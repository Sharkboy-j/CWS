package telegram

import (
	"context"
	"cws/database"
	"cws/logger"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler struct {
	bot       *tgbotapi.BotAPI
	repo      *database.Repository
	stateMgr  *StateManager
	msgSender *MessageSender
}

func NewCommandHandler(bot *tgbotapi.BotAPI, repo *database.Repository, stateMgr *StateManager, msgSender *MessageSender) *CommandHandler {
	return &CommandHandler{
		bot:       bot,
		repo:      repo,
		stateMgr:  stateMgr,
		msgSender: msgSender,
	}
}

func (ch *CommandHandler) HandleCommand(message *tgbotapi.Message) {
	command := message.Command()
	chatId := message.Chat.ID

	switch command {
	case "menu", "start":
		ch.ShowMainMenu(chatId)
	case "check":
		ch.handleCheckCommand(chatId)
	case "clients":
		ch.HandleClientsCommand(chatId)
	default:
		logger.Warn("Пользователь %d выполнил неизвестную команду: /%s", chatId, command)
		msg := tgbotapi.NewMessage(chatId, "Неизвестная команда")
		ch.msgSender.Send(msg)
	}
}

func (ch *CommandHandler) handleCheckCommand(chatId int64) {
	logger.Debugf("Пользователь %d выполнил команду Check", chatId)
	ch.ShowMainMenu(chatId)
}

func (ch *CommandHandler) ShowMainMenu(chatId int64) {
	logger.Debugf("Показ главного меню для пользователя %d", chatId)
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	text := "🏠 *Главное меню*\n\nВыберите действие:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Првоерить обновления", "check_torrents"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Клиенты", "clients"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📥 Добавить торрент файл", "add_torrent_file"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔎 Поиск торрента", "search_torrent"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Мониторинг торрента", "monitor_torrent"),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *CommandHandler) ShowCheckClientsList(chatId int64) {
	logger.Debugf("Пользователь %d запросил список клиентов для проверки", chatId)
	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении списка клиентов")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d имеет %d клиентов для проверки", chatId, len(clients))

	messageID := ch.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := "📋 *Проверка активных торрентов*\n\nКлиенты не найдены. Добавьте клиента для проверки."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if err != nil {
			logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
			return
		}
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		return
	}
	var text strings.Builder
	text.WriteString("📋 *Проверка активных торрентов*\n\n")
	text.WriteString("Выберите клиента для проверки:\n\n")

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, client := range clients {
		sslText := "🔒"
		if !client.SSL {
			sslText = "🔓"
		}
		text.WriteString(fmt.Sprintf("%s *%s*\n", sslText, client.Name))
		text.WriteString(fmt.Sprintf("   `%s:%d`\n\n", client.Host, client.Port))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🔍 Проверить %s", client.Name),
				fmt.Sprintf("check_client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text.String(), &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *CommandHandler) HandleClientsCommand(chatId int64) {
	logger.Debugf("Пользователь %d запросил список клиентов", chatId)
	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении списка клиентов")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d имеет %d клиентов", chatId, len(clients))

	messageID := ch.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := "📋 *Клиенты qBittorrent*\n\nКлиенты не найдены. Добавьте первого клиента."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if err != nil {
			logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
			return
		}
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		return
	}
	var text strings.Builder
	text.WriteString("📋 *Клиенты qBittorrent*\n\n")

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, client := range clients {
		sslText := "🔒"
		if !client.SSL {
			sslText = "🔓"
		}
		text.WriteString(fmt.Sprintf("%s *%s*\n", sslText, client.Name))
		text.WriteString(fmt.Sprintf("   `%s:%d`\n\n", client.Host, client.Port))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🔧 %s", client.Name),
				fmt.Sprintf("client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text.String(), &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}
