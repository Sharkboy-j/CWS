package quick_actions

import (
	"cws/internal/bot/ui"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"cws/logger"
)

func (h *Handler) sendOrEditResultWithMainMenu(
	chatId int64,
	messageID int,
	text string,
	successCount int,
	failCount int,
	failedClients []string,
) bool {
	if failCount > 0 {
		text += fmt.Sprintf("\n❌ Ошибки (%d):\n", failCount)
		for _, name := range failedClients {
			text += fmt.Sprintf("  • %s\n", name)
		}
	}

	text += fmt.Sprintf("\nВсего обработано: %d успешно, %d с ошибками", successCount, failCount)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.MainMenu),
		),
	)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return false
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)

	return true
}
