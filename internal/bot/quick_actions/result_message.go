package quick_actions

import (
	"cws/internal/bot/ui"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"cws/logger"
)

func (h *Handler) sendOrEditResult(
	chatId int64,
	messageID int,
	text string,
	successCount int,
	failCount int,
	failedClients []string,
	keyboard *tgbotapi.InlineKeyboardMarkup,
) bool {
	if failCount > 0 {
		text += ui.Msgf(ui.MsgResultErrorsHeaderFmt, failCount)
		for _, name := range failedClients {
			text += ui.Msgf(ui.MsgResultErrorsItemFmt, name)
		}
	}

	text += ui.Msgf(ui.MsgResultTotalsFmt, successCount, failCount)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return false
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)

	return true
}

func (h *Handler) sendOrEditResultWithMainMenu(
	chatId int64,
	messageID int,
	text string,
	successCount int,
	failCount int,
	failedClients []string,
) bool {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.MainMenu),
		),
	)

	return h.sendOrEditResult(chatId, messageID, text, successCount, failCount, failedClients, &keyboard)
}
