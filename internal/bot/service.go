package bot

import (
	"context"
	"cws/config"
	"cws/internal/storage"
	"cws/internal/telegram"
	"cws/internal/telegram/messaging"
	"cws/logger"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	telegramService *telegram.BotService
	autoChecker     *AutoChecker
}

func NewBotService(token string, repo *storage.Repository, cfg *config.Config) (*Service, error) {
	stateMgr, err := NewStateManager(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	telegramCfg := telegram.BotServiceConfig{
		Token:        token,
		Repo:         repo,
		StateMgr:     stateMgr,
		MsgSender:    nil,
		CmdHdlr:      nil,
		CallbackHdlr: nil,
		DialogHdlr:   nil,
		DocHdlr:      nil,
	}

	telegramService, err := telegram.NewBotService(telegramCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram service: %w", err)
	}

	bot := telegramService.GetBot()
	msgSender := messaging.NewMessageSender(bot)

	cmdHdlr := NewCommandHandler(bot, repo, stateMgr, msgSender)
	clientHdlr := NewClientHandler(repo, msgSender, stateMgr, cfg)
	dialogHdlr := NewDialogHandler(repo, msgSender, stateMgr, clientHdlr)
	callbackHdlr := NewCallbackHandler(bot, stateMgr, msgSender, dialogHdlr, clientHdlr)
	docHdlr := NewDocumentHandler(stateMgr, msgSender, clientHdlr)

	clientHdlr.SetCommandHandler(cmdHdlr)
	dialogHdlr.SetCommandHandler(cmdHdlr)
	callbackHdlr.SetCommandHandler(cmdHdlr)

	telegramService.SetHandlers(msgSender, cmdHdlr, callbackHdlr, dialogHdlr, docHdlr)

	var ac *AutoChecker
	if !cfg.ManualCheckOnly && cfg.DurationSeconds > 0 {
		ac = NewAutoChecker(clientHdlr, cfg.DurationSeconds)
	}

	commands := []tgbotapi.BotCommand{
		{
			Command:     "menu",
			Description: "Главное меню",
		},
		{
			Command:     "check",
			Description: "Проверить статус",
		},
		{
			Command:     "clients",
			Description: "Управление клиентами qBittorrent",
		},
	}

	err = telegramService.SetCommands(commands)
	if err != nil {
		logger.Warn("Не удалось установить команды бота: %v", err)
	}

	service := &Service{
		telegramService: telegramService,
		autoChecker:     ac,
	}

	logger.Debugf("Bot service initialized")

	return service, nil
}

func (bs *Service) Start(ctx context.Context) error {
	return bs.telegramService.Start(ctx)
}

func (bs *Service) StartAutoChecker(ctx context.Context) {
	if bs.autoChecker != nil {
		go bs.autoChecker.Start(ctx)
	}
}
