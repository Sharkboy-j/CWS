package telegram

import (
	"sync"

	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botInstance *tgbotapi.BotAPI
	botOnce     sync.Once
	botMutex    sync.RWMutex
)

func initBot(token string) (*tgbotapi.BotAPI, error) {
	botMutex.RLock()
	if botInstance != nil {
		botMutex.RUnlock()
		logger.Debug("Bot instance already exists, returning existing instance")

		return botInstance, nil
	}
	botMutex.RUnlock()

	var initErr error
	botOnce.Do(func() {
		logger.Debugf("try init bot '%s'", token)
		bot, err := tgbotapi.NewBotAPI(token)
		if err != nil {
			initErr = err

			return
		}

		logger.Infof("telegram bot authorized as '%s'", bot.Self.UserName)

		botMutex.Lock()
		botInstance = bot
		botMutex.Unlock()
	})

	if initErr != nil {
		return nil, initErr
	}

	botMutex.RLock()
	defer botMutex.RUnlock()

	return botInstance, nil
}
