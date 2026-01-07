package quick_actions

import (
	"cws/internal/bot/ui"
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowQuickActionsMenu(chatId int64) {
	logger.Debugf("Showing quick actions menu for user %d", chatId)
	_, _, messageID, ok := h.getClientsAndMenuMessageOrReplyWithMainMenu(chatId, ui.Msg(ui.MsgQuickActionsMenuNoClientsText))
	if ok {
		text := ui.Msg(ui.MsgQuickActionsMenuText)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.PauseTorrentsMenu),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.ResumeTorrentsMenu),
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
	} else {
		return
	}
}
