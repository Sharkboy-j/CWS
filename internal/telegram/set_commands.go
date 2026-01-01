package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"cws/logger"
)

func (bs *BotService) SetCommands(commands []tgbotapi.BotCommand) error {
	setCommands := tgbotapi.NewSetMyCommands(commands...)
	_, err := bs.bot.Request(setCommands)
	if err != nil {
		return err
	}

	logger.Info("Команды бота установлены успешно")

	return nil
}
