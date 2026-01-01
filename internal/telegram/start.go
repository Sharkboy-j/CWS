package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"cws/logger"
)

func (bs *BotService) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bs.bot.GetUpdatesChan(u)
	logger.Info("Бот запущен и ожидает обновлений...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Получен сигнал остановки, завершаем работу бота...")
			bs.bot.StopReceivingUpdates()

			return nil
		case update := <-updates:
			logger.Debug("Получено обновление: UpdateID=%d", update.UpdateID)
			bs.handleUpdate(update)
		}
	}
}
