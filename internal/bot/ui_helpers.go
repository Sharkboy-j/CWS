package bot

import (
	"fmt"

	"cws/internal/bot/ui"
	"cws/internal/telegram/messaging"
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func paginateRange(total, pageSize, page int) (pageOut, totalPages, start, end int) {
	if pageSize <= 0 {
		pageSize = 1
	}
	if total < 0 {
		total = 0
	}

	totalPages = (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start = page * pageSize
	end = start + pageSize
	if end > total {
		end = total
	}

	return page, totalPages, start, end
}

func truncateButtonLabel(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}

	return s[:max-3] + "..."
}

func buildMissingTorrentRowButtons(info missingTorrentInfo) []tgbotapi.InlineKeyboardButton {
	btnText := truncateButtonLabel(info.name, 60)
	row := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonURL(btnText, info.url),
	}
	if info.hash != "" {
		row = append(row, ui.Data(ui.IconChart, fmt.Sprintf("monitor_from_missing_%s", info.hash)))
	}

	return row
}

func sendSingleButtonPromptWithState(
	stateMgr *StateManager,
	msgSender messaging.MessageSender,
	chatId int64,
	userState string,
	text string,
	button tgbotapi.InlineKeyboardButton,
	logContext string,
) error {
	if userState != "" {
		stateMgr.SetUserState(chatId, userState)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)
	messageID := stateMgr.GetMenuMessage(chatId)
	newMessageID, err := msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("%s %d: %v", logContext, chatId, err)

		return err
	}
	stateMgr.SetMenuMessage(chatId, newMessageID)

	return nil
}
