package telegram

import (
	"context"
	"cws/logger"
	"time"
)

type AutoChecker struct {
	clientHdlr *ClientHandler
	interval   time.Duration
	lastCheck  map[int64]time.Time // Время последней проверки для каждого пользователя
}

func NewAutoChecker(clientHdlr *ClientHandler, intervalSeconds int) *AutoChecker {
	return &AutoChecker{
		clientHdlr: clientHdlr,
		interval:   time.Duration(intervalSeconds) * time.Second,
		lastCheck:  make(map[int64]time.Time),
	}
}

func (ac *AutoChecker) Start(ctx context.Context) {
	logger.Info("Автоматическая проверка запущена с интервалом %v", ac.interval)

	go ac.runCheck(ctx)

	ticker := time.NewTicker(ac.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Автоматическая проверка остановлена")
			return
		case <-ticker.C:
			go ac.runCheck(ctx)
		}
	}
}

func (ac *AutoChecker) runCheck(ctx context.Context) {
	logger.Info("Начало автоматической проверки всех пользователей")
	startTime := time.Now()

	userIDs, err := ac.clientHdlr.repo.GetAllUserIDs(ctx)
	if err != nil {
		logger.Error("Ошибка при получении списка пользователей: %v", err)
		return
	}

	if len(userIDs) == 0 {
		logger.Debug("Нет пользователей для проверки")
		return
	}

	logger.Info("Найдено %d пользователей для автоматической проверки", len(userIDs))

	for _, userID := range userIDs {
		select {
		case <-ctx.Done():
			logger.Info("Автоматическая проверка прервана")
			return
		default:
		}

		logger.Debug("Автоматическая проверка для пользователя %d", userID)
		ac.clientHdlr.CheckAllClientsAuto(userID)
		ac.lastCheck[userID] = time.Now()
	}

	elapsed := time.Since(startTime)
	logger.Info("Автоматическая проверка завершена за %v для %d пользователей", elapsed, len(userIDs))
}

func (ac *AutoChecker) GetLastCheckTime(userID int64) (time.Time, bool) {
	t, exists := ac.lastCheck[userID]
	return t, exists
}
