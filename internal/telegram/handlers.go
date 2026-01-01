package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler interface {
	HandleCommand(message *tgbotapi.Message)
}

type CallbackHandler interface {
	HandleCallbackQuery(query *tgbotapi.CallbackQuery)
}

type DialogHandler interface {
	HandleMessage(message *tgbotapi.Message)
}

type DocumentHandler interface {
	HandleDocument(ctx context.Context, chatId int64, document *tgbotapi.Document, fileData []byte)
}

type StateManager interface {
	GetUserState(chatId int64) (string, bool)
	GetMenuMessage(chatId int64) int
	SetMenuMessage(chatId int64, messageID int)
}
