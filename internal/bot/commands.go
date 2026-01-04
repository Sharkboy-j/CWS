package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler struct {
	bot            *tgbotapi.BotAPI
	repo           *storage.Repository
	stateMgr       *StateManager
	msgSender      messaging.MessageSender
	isStartCommand bool // Track if current operation is from /start command
}

func NewCommandHandler(bot *tgbotapi.BotAPI, repo *storage.Repository, stateMgr *StateManager, msgSender messaging.MessageSender) *CommandHandler {
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
		ch.isStartCommand = true
		ch.ShowMainMenu(chatId)
		ch.isStartCommand = false
	case "check":
		ch.handleCheckCommand(chatId)
	case "clients":
		ch.HandleClientsCommand(chatId)
	default:
		logger.Warn("Пользователь %d выполнил неизвестную команду: /%s", chatId, command)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "Неизвестная команда", nil)
	}
}

func (ch *CommandHandler) handleCheckCommand(chatId int64) {
	logger.Debugf("Пользователь %d выполнил команду Check", chatId)
	ch.ShowMainMenu(chatId)
}

func (ch *CommandHandler) ShowMainMenu(chatId int64) {
	logger.Debugf("Показ главного меню для пользователя %d", chatId)
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	// For /start command, always send new message to ensure menu appears after chat clearing
	if ch.isStartCommand {
		messageID = 0
	}

	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
	}

	var text strings.Builder
	if len(clients) > 0 {
		for _, client := range clients {
			qbClient, qbErr := qbit.New(ctx, client)
			if qbErr != nil {
				logger.Error("Ошибка при подключении к клиенту %s: %v", client.Name, qbErr)
				text.WriteString(fmt.Sprintf("❌ *%s* - ошибка подключения\n\n", client.Name))

				continue
			}

			transferInfo, transferErr := qbClient.GetTransferInfo(ctx)
			if transferErr != nil {
				logger.Error("Ошибка при получении информации о передаче для клиента %s: %v", client.Name, transferErr)
				text.WriteString(fmt.Sprintf("⚠️ *%s* - ошибка получения данных\n\n", client.Name))

				continue
			}

			text.WriteString(fmt.Sprintf("🔹 *%s* ", client.Name))

			formatSpeed := func(bytesPerSec int64) string {
				if bytesPerSec == 0 {
					return "0 B/s"
				}
				if bytesPerSec < 1024 {
					return fmt.Sprintf("%d B/s", bytesPerSec)
				}
				if bytesPerSec < 1024*1024 {
					return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
				}

				return fmt.Sprintf("%.1f MB/s", float64(bytesPerSec)/(1024*1024))
			}

			formatLimit := func(bytesPerSec int64) string {
				if bytesPerSec == 0 {
					return ""
				}

				mbPerSec := float64(bytesPerSec) / (1024 * 1024)

				return fmt.Sprintf("%.2f МБ/с", mbPerSec)
			}

			text.WriteString(fmt.Sprintf("⬇️ %s", formatSpeed(transferInfo.DownloadSpeed)))
			if transferInfo.DownloadLimit > 0 {
				text.WriteString(fmt.Sprintf(" \\[%s]", formatLimit(transferInfo.DownloadLimit)))
			}
			text.WriteString(" ")

			text.WriteString(fmt.Sprintf("⬆️ %s", formatSpeed(transferInfo.UploadSpeed)))
			if transferInfo.UploadLimit > 0 {
				text.WriteString(fmt.Sprintf(" \\[%s]", formatLimit(transferInfo.UploadLimit)))
			}
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.CheckTorrents),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.QuickActionsMenu),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SettingsMenu),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.AddTorrentFile),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SearchTorrent),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.MonitorTorrent),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text.String(), &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *CommandHandler) ShowSettingsMenu(chatId int64) {
	logger.Debugf("Показ меню настроек для пользователя %d", chatId)
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	text := "⚙️ *Настройки*\n\nВыберите раздел настроек:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.ClientsMenu),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.Variables),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.MainMenu),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении меню настроек для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

// ShowVariablesMenu displays variables; current values are fetched from DB (defaults applied).
func (ch *CommandHandler) ShowVariablesMenu(chatId int64) {
	// Show list of editable variables. Actual editing opens when user selects a variable.
	logger.Debugf("Показ списка переменных для пользователя %d", chatId)
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	// fetch current value to show next to variable name
	ctx := context.Background()
	count, err := ch.repo.GetRecommendedTorrents(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении recommended_torrents для пользователя %d: %v", chatId, err)
		count = 3
	}

	text := "🔧 *Переменные*\n\nВыберите переменную для редактирования:"
	varLabel := fmt.Sprintf("%s: %d", ui.Text(ui.RecommendedTorrents), count)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Data(varLabel, "edit_recommended_torrents"),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SettingsMenu),
			ui.Button(ui.MainMenu),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении списка переменных для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

// ShowVariablesMenuWithValue is like ShowVariablesMenu but uses provided count
// to display next to the variable (useful immediately after saving).
func (ch *CommandHandler) ShowVariablesMenuWithValue(chatId int64, count int) {
	logger.Debugf("Показ списка переменных (override) для пользователя %d", chatId)
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	text := "🔧 *Переменные*\n\nВыберите переменную для редактирования:"
	varLabel := fmt.Sprintf("%s: %d", ui.Text(ui.RecommendedTorrents), count)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Data(varLabel, "edit_recommended_torrents"),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SettingsMenu),
			ui.Button(ui.MainMenu),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении списка переменных для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

// ShowEditRecommendedTorrents displays editor for recommended torrents count.
func (ch *CommandHandler) ShowEditRecommendedTorrents(chatId int64) {
	logger.Debugf("Показ редактора recommended torrents для пользователя %d", chatId)
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	ctx := context.Background()
	count, err := ch.repo.GetRecommendedTorrents(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении recommended_torrents для пользователя %d: %v", chatId, err)
		count = 3
	}

	text := fmt.Sprintf("🔧 *Редактирование переменной*\n\nРекомендуемое количество торрентов на странице выбора мониторинга: *%d*\n\nВыберите новое значение:", count)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Data("1", "set_recommended_torrents_1"), ui.Data("2", "set_recommended_torrents_2"), ui.Data("3", "set_recommended_torrents_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Data("4", "set_recommended_torrents_4"), ui.Data("5", "set_recommended_torrents_5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.SpeedCustom, "edit_recommended_torrents_input"),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.Variables),
			ui.Button(ui.MainMenu),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении редактора переменной для пользователя %d: %v", chatId, err)

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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "Ошибка при получении списка клиентов", nil)

		return
	}
	logger.Debugf("Пользователь %d имеет %d клиентов для проверки", chatId, len(clients))

	messageID := ch.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := "📋 *Проверка активных торрентов*\n\nКлиенты не найдены. Добавьте клиента для проверки."
		sendErr := sendNoClientsMessage(ch.msgSender, ch.stateMgr, chatId, text)
		if sendErr != nil {
			return
		}

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
			ui.Data(
				fmt.Sprintf("🔍 Проверить %s", client.Name),
				fmt.Sprintf("check_client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.MainMenu),
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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "Ошибка при получении списка клиентов", nil)

		return
	}
	logger.Debugf("Пользователь %d имеет %d клиентов", chatId, len(clients))

	messageID := ch.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := "📋 *Клиенты qBittorrent*\n\nКлиенты не найдены. Добавьте первого клиента."
		err = sendNoClientsMessage(ch.msgSender, ch.stateMgr, chatId, text)
		if err != nil {
			return
		}

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
			ui.Data(
				fmt.Sprintf("🔧 %s", client.Name),
				fmt.Sprintf("client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.AddClient),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.MainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text.String(), &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}
