package quick_actions

import (
	"cws/internal/bot/ui"
	"cws/internal/dialogstate"
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) StartCustomSpeedLimitDialog(chatId int64) {
	logger.Debugf("Starting custom speed limit dialog for user %d", chatId)
	if h.stateSetter != nil {
		h.stateSetter.SetUserState(chatId, string(dialogstate.StateCustomSpeedLimit))
	}
	text := ui.Msg(ui.MsgSpeedLimitCustomPromptText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "quick_actions"),
		),
	)
	messageID := h.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}
