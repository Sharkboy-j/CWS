package messaging

import (
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (ms *messageSender) DeleteMessage(chatId int64, messageID int) {
	if messageID == 0 {
		return
	}
	deleteMsg := tgbotapi.NewDeleteMessage(chatId, messageID)
	_, err := ms.bot.Send(deleteMsg)
	if err != nil {
		logger.Debug("Не удалось удалить сообщение %d для пользователя %d: %v (возможно, уже удалено)", messageID, chatId, err)
	} else {
		logger.Debug("Удалено сообщение %d для пользователя %d", messageID, chatId)
	}
}
