package main

import (
	"context"
	"cws/config"
	"cws/internal/bot"
	"cws/internal/storage"
	"cws/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancelFunction := context.WithCancel(context.Background())
	defer cancelFunction()

	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		logger.Errorf("err during reading config enviroment: %s", err)
		os.Exit(1)
	}

	logger.SetLevelFromString(cfg.LogLevel)
	logger.Debugf("Логирование инициализировано с уровнем: %s", cfg.LogLevel)

	repo, err := storage.NewRepository(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		logger.Errorf("err during database initialization: %s", err)
		os.Exit(1)
	}
	defer func() {
		_ = repo.Close()
	}()

	botService, err := bot.NewBotService(cfg.TelegramToken, repo, cfg)
	if err != nil {
		logger.Errorf("err during bot initialization: %s", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if startErr := botService.Start(ctx); startErr != nil {
			logger.Errorf("Ошибка при работе бота: %v", startErr)
		}
	}()

	botService.StartAutoChecker(ctx)

	<-sigChan
	logger.Info("Получен сигнал завершения, останавливаем сервис...")
	cancelFunction()

	os.Exit(0)
}
