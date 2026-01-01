package messaging

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type messageSender struct {
	bot *tgbotapi.BotAPI
}

type MessageSender interface {
	SendOrEdit(chatId int64, messageID int, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) (int, error)
	DeleteMessage(chatId int64, messageID int)
}

func NewMessageSender(bot *tgbotapi.BotAPI) MessageSender {
	return &messageSender{bot: bot}
}
