package quick_actions

import (
	"cws/internal/bot/ui"
	"cws/logger"
)

func (h *Handler) ShowPauseTorrentsMenu(chatId int64) {
	_, _, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}

	text := ui.Msg(ui.MsgQuickActionsPauseMenuText)
	keyboard := h.pauseTorrentsKeyboard()

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}
