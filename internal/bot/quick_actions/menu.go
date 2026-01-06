package quick_actions

import (
	"context"
	"cws/internal/bot/ui"
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
		_, _ = h.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgClientsListError), nil)

		return
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := ui.Msg(ui.MsgQuickActionsMenuNoClientsText)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		newMessageID, sendErr := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if sendErr != nil {
			logger.Error("Error sending message for user %d: %v", chatId, sendErr)

			return
		}
		h.stateMgr.SetMenuMessage(chatId, newMessageID)

		return
	}

	text := ui.Msg(ui.MsgQuickActionsMenuText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.PauseAllTorrents),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.ResumeAllTorrents),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SpeedLimitMenu),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.MainMenu),
		),
	)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (h *Handler) getClientsAndMenuMessageOrReply(chatId int64) (context.Context, []*storage.Client, int, bool) {
	ctx := context.Background()
	clients, err := h.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Error getting clients for user %d: %v", chatId, err)
		_, _ = h.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgClientsListErrorWithEmoji), nil)

		return nil, nil, 0, false
	}

	if len(clients) == 0 {
		_, _ = h.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgQuickActionsNoClients), nil)

		return nil, nil, 0, false
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)

	return ctx, clients, messageID, true
}
