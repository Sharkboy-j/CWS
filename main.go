package main

import (
	"context"
	"cws/config"
	"cws/database"
	"cws/logger"
	"cws/telegram"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancelFunction := context.WithCancel(context.Background())
	defer cancelFunction()

	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		log.Fatalf("err during reading config enviroment: %s", err)
	}

	logger.SetLevelFromString(cfg.LogLevel)
	logger.Info("Логирование инициализировано с уровнем: %s", cfg.LogLevel)

	repo, err := database.NewRepository(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("err during database initialization: %s", err)
	}
	defer repo.Close()

	botService, err := telegram.NewBotService(cfg.TelegramToken, cfg.ChatId, repo, cfg)
	if err != nil {
		log.Fatalf("err during bot initialization: %s", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := botService.Start(ctx); err != nil {
			log.Printf("Ошибка при работе бота: %v", err)
		}
	}()

	botService.StartAutoChecker(ctx)

	<-sigChan
	log.Println("Получен сигнал завершения, останавливаем сервис...")
	cancelFunction()

	os.Exit(0)
}
