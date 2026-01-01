package telegram

import (
	"context"
	"cws/config"
	"cws/database"
	"cws/logger"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ClientHandler struct {
	repo                  *database.Repository
	msgSender             *MessageSender
	stateMgr              *StateManager
	cmdHdlr               *CommandHandler // Для показа списка клиентов после удаления
	cfg                   *config.Config
	missingTorrentsCache  map[int64][]missingTorrentInfo // Кэш мёртвых торрентов для пагинации (chatId -> список торрентов)
	checkResultsCache     map[int64]*CheckResultsCache   // Кэш результатов проверки для пагинации
}

type CheckResultsCache struct {
	Results           []ClientCheckResult
	TotalDuration     time.Duration
	LastCheckTime     *time.Time
	AllMissingTorrents []missingTorrentInfo // Все мёртвые торренты со всех клиентов
}

func NewClientHandler(repo *database.Repository, msgSender *MessageSender, stateMgr *StateManager, cfg *config.Config) *ClientHandler {
	return &ClientHandler{
		repo:                 repo,
		msgSender:            msgSender,
		stateMgr:             stateMgr,
		cfg:                  cfg,
		missingTorrentsCache: make(map[int64][]missingTorrentInfo),
		checkResultsCache:    make(map[int64]*CheckResultsCache),
	}
}

func (ch *ClientHandler) SetCommandHandler(cmdHdlr *CommandHandler) {
	ch.cmdHdlr = cmdHdlr
}

func (ch *ClientHandler) ShowClientDetails(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался просмотреть несуществующий клиент %d", chatId, clientID)
		msg := tgbotapi.NewMessage(chatId, "Клиент не найден или у вас нет доступа")
		ch.msgSender.Send(msg)
		return
	}

	sslText := "Нет"
	if client.SSL {
		sslText = "Да"
	}

	text := fmt.Sprintf("🔧 *%s*\n\n"+
		"Host: `%s`\n"+
		"Port: `%d`\n"+
		"Username: `%s`\n"+
		"SSL: `%s`\n",
		client.Name, client.Host, client.Port, client.Username, sslText)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Изменить", fmt.Sprintf("edit_client_%d", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("delete_client_%d", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к списку", "clients"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)

	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) ShowDeleteConfirmation(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для удаления пользователем %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался удалить несуществующий клиент %d", chatId, clientID)
		msg := tgbotapi.NewMessage(chatId, "Клиент не найден или у вас нет доступа")
		ch.msgSender.Send(msg)
		return
	}

	text := fmt.Sprintf("⚠️ *Подтверждение удаления*\n\nВы уверены, что хотите удалить клиента *%s*?\n\nЭто действие нельзя отменить!", client.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", fmt.Sprintf("confirm_delete_%d", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", fmt.Sprintf("client_%d", clientID)),
		),
	)

	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) DeleteClient(chatId int64, clientID int64) {
	ctx := context.Background()

	// Сначала получаем имя клиента для сообщения
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для удаления пользователем %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался удалить несуществующий клиент %d", chatId, clientID)
		msg := tgbotapi.NewMessage(chatId, "Клиент не найден или у вас нет доступа")
		ch.msgSender.Send(msg)
		return
	}

	clientName := client.Name

	// Удаляем клиента
	err = ch.repo.DeleteClient(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при удалении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при удалении клиента. Попробуйте снова.")
		ch.msgSender.Send(msg)
		return
	}

	logger.Debugf("Пользователь %d успешно удалил клиента: ID=%d, Name=%s", chatId, clientID, clientName)

	if ch.cmdHdlr != nil {
		ch.cmdHdlr.HandleClientsCommand(chatId)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := d.Seconds() - float64(minutes*60)
		return fmt.Sprintf("%dm %.2fs", minutes, seconds)
	}
}
