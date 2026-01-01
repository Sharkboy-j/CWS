package telegram

import (
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageSender struct {
	bot *tgbotapi.BotAPI
}

func NewMessageSender(bot *tgbotapi.BotAPI) *MessageSender {
	return &MessageSender{bot: bot}
}

func (ms *MessageSender) Send(msg tgbotapi.MessageConfig) error {
	chatId := msg.ChatID
	logger.Debug("Отправка сообщения пользователю %d", chatId)
	_, err := ms.bot.Send(msg)
	if err != nil {
		logger.Error("Ошибка при отправке сообщения пользователю %d: %v", chatId, err)
		return err
	}
	logger.Debug("Сообщение успешно отправлено пользователю %d", chatId)
	return nil
}

func (ms *MessageSender) SendOrEdit(chatId int64, messageID int, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) (int, error) {
	if messageID == 0 {
		msg := tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = "markdown"
		if replyMarkup != nil {
			msg.ReplyMarkup = replyMarkup
		}
		sentMsg, err := ms.bot.Send(msg)
		if err != nil {
			logger.Error("Ошибка при отправке сообщения пользователю %d: %v", chatId, err)
			return 0, err
		}
		logger.Debug("Отправлено новое сообщение пользователю %d, messageID=%d", chatId, sentMsg.MessageID)
		return sentMsg.MessageID, nil
	} else {
		editMsg := tgbotapi.NewEditMessageText(chatId, messageID, text)
		editMsg.ParseMode = "markdown"
		if replyMarkup != nil {
			editMsg.ReplyMarkup = replyMarkup
		}
		_, err := ms.bot.Send(editMsg)
		if err != nil {
			logger.Warn("Не удалось обновить сообщение %d для пользователя %d: %v, отправляем новое", messageID, chatId, err)
			return ms.SendOrEdit(chatId, 0, text, replyMarkup)
		}
		logger.Debug("Обновлено сообщение %d для пользователя %d", messageID, chatId)
		return messageID, nil
	}
}

func (ms *MessageSender) DeleteMessage(chatId int64, messageID int) {
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
