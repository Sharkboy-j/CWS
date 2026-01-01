package messaging

import (
	"strings"

	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (ms *messageSender) SendOrEdit(chatId int64, messageID int, text string, replyMarkup *tgbotapi.InlineKeyboardMarkup) (int, error) {
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
	}
	editMsg := tgbotapi.NewEditMessageText(chatId, messageID, text)
	editMsg.ParseMode = "markdown"
	if replyMarkup != nil {
		editMsg.ReplyMarkup = replyMarkup
	}
	response, err := ms.bot.Send(editMsg)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "message to edit not found") ||
			strings.Contains(errStr, "message not found") ||
			strings.Contains(errStr, "bad request") ||
			strings.Contains(errStr, "message can't be edited") {
			logger.Warn("Сообщение %d для пользователя %d не найдено или не может быть отредактировано: %v, отправляем новое", messageID, chatId, err)
		} else {
			logger.Warn("Не удалось обновить сообщение %d для пользователя %d: %v, отправляем новое", messageID, chatId, err)
		}
		msg := tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = "markdown"
		if replyMarkup != nil {
			msg.ReplyMarkup = replyMarkup
		}
		sentMsg, sendErr := ms.bot.Send(msg)
		if sendErr != nil {
			logger.Error("Ошибка при отправке нового сообщения пользователю %d после неудачного редактирования: %v", chatId, sendErr)

			return 0, sendErr
		}
		logger.Debug("Отправлено новое сообщение пользователю %d вместо редактирования, messageID=%d", chatId, sentMsg.MessageID)

		return sentMsg.MessageID, nil
	}
	if response.MessageID == 0 || response.MessageID != messageID {
		logger.Warn("Ответ от API при редактировании сообщения %d для пользователя %d некорректный (response.MessageID=%d), отправляем новое", messageID, chatId, response.MessageID)
		msg := tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = "markdown"
		if replyMarkup != nil {
			msg.ReplyMarkup = replyMarkup
		}
		sentMsg, sendErr := ms.bot.Send(msg)
		if sendErr != nil {
			logger.Error("Ошибка при отправке нового сообщения пользователю %d после некорректного ответа: %v", chatId, sendErr)

			return 0, sendErr
		}
		logger.Debug("Отправлено новое сообщение пользователю %d после некорректного ответа, messageID=%d", chatId, sentMsg.MessageID)

		return sentMsg.MessageID, nil
	}
	logger.Debug("Обновлено сообщение %d для пользователя %d", messageID, chatId)

	return messageID, nil
}
