package quick_actions

import (
	"context"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	repo        *storage.Repository
	msgSender   messaging.MessageSender
	stateMgr    StateManager
	cmdHdlr     CommandHandler
	stateSetter StateSetter
}

type StateSetter interface {
	SetUserState(chatId int64, state string)
}

type StateManager interface {
	GetMenuMessage(chatId int64) int
	SetMenuMessage(chatId int64, messageID int)
}

type CommandHandler interface {
	ShowMainMenu(chatId int64)
}

func NewHandler(repo *storage.Repository, msgSender messaging.MessageSender, stateMgr StateManager) *Handler {
	return &Handler{
		repo:      repo,
		msgSender: msgSender,
		stateMgr:  stateMgr,
	}
}

func (h *Handler) SetCommandHandler(cmdHdlr CommandHandler) {
	h.cmdHdlr = cmdHdlr
}

func (h *Handler) SetStateSetter(stateSetter StateSetter) {
	h.stateSetter = stateSetter
}

func (h *Handler) ShowQuickActionsMenu(chatId int64) {
	logger.Debugf("Showing quick actions menu for user %d", chatId)
	ctx := context.Background()
	clients, err := h.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Error getting clients for user %d: %v", chatId, err)
		_, _ = h.msgSender.SendOrEdit(chatId, 0, "Ошибка при получении списка клиентов", nil)

		return
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := "⚡ *Быстрые действия*\n\nКлиенты не найдены. Добавьте клиента для использования быстрых действий."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if err != nil {
			logger.Error("Error sending message for user %d: %v", chatId, err)

			return
		}
		h.stateMgr.SetMenuMessage(chatId, newMessageID)

		return
	}

	text := "⚡ *Быстрые действия*\n\nВыберите действие:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏸ Остановить все раздачи", "quick_action_pause_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶ Запустить все раздачи", "quick_action_resume_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚦 Ограничить скорость", "quick_action_limit_speed_menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}
