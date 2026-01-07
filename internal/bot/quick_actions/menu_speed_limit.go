package quick_actions

import (
	"cws/internal/bot/ui"
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowSpeedLimitMenu(chatId int64) {
	logger.Debugf("Showing speed limit menu for user %d", chatId)
	_, _, messageID, ok := h.getClientsAndMenuMessageOrReplyWithMainMenu(chatId, ui.Msg(ui.MsgSpeedLimitMenuNoClientsText))
	if ok {
		text := ui.Msg(ui.MsgSpeedLimitMenuText)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.Speed10),
				ui.Button(ui.Speed100),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.Speed500),
				ui.Button(ui.Speed1000),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.Speed2000),
				ui.Button(ui.Speed5000),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.Speed10000),
				ui.Button(ui.Speed50000),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.SpeedCustom),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.SpeedRemove),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.ButtonWithData(ui.Back, "quick_actions"),
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
