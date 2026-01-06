package bot

import (
	"context"
	"cws/config"
	"cws/internal/bot/ui"
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
	notifyBot       *tgbotapi.BotAPI
	repo            *storage.Repository
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

	var notifySender messaging.MessageSender
	var notifyBot *tgbotapi.BotAPI
	if cfg.TelegramMsgToken != "" {
		nb, initErr := tgbotapi.NewBotAPI(cfg.TelegramMsgToken)
		if initErr != nil {
			return nil, fmt.Errorf("failed to init notification telegram bot: %w", initErr)
		}
		logger.Infof("notification telegram bot authorized as '%s'", nb.Self.UserName)
		notifyBot = nb
		notifySender = messaging.NewMessageSender(nb)
		cmdHdlr.SetNotifyBotUsername(nb.Self.UserName)
	}

	clientHdlr := NewClientHandler(repo, msgSender, notifySender, stateMgr, cfg, bot.Self.UserName)
	dialogHdlr := NewDialogHandler(repo, msgSender, stateMgr, clientHdlr)
	callbackHdlr := NewCallbackHandler(bot, stateMgr, msgSender, dialogHdlr, clientHdlr)
	docHdlr := NewDocumentHandler(stateMgr, msgSender, clientHdlr)

	clientHdlr.SetCommandHandler(cmdHdlr)
	cmdHdlr.SetClientHandler(clientHdlr)
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
			Description: ui.Msg(ui.MsgBotCommandMenuDescription),
		},
		{
			Command:     "check",
			Description: ui.Msg(ui.MsgBotCommandCheckDescription),
		},
		{
			Command:     "clients",
			Description: ui.Msg(ui.MsgBotCommandClientsDescription),
		},
	}

	err = telegramService.SetCommands(commands)
	if err != nil {
		logger.Warn("Не удалось установить команды бота: %v", err)
	}

	service := &Service{
		telegramService: telegramService,
		autoChecker:     ac,
		notifyBot:       notifyBot,
		repo:            repo,
	}

	logger.Debugf("Bot service initialized")

	return service, nil
}

func (bs *Service) Start(ctx context.Context) error {
	if bs.notifyBot != nil {
		go StartNotifyBot(ctx, bs.notifyBot, bs.repo)
	}

	return bs.telegramService.Start(ctx)
}

func (bs *Service) StartAutoChecker(ctx context.Context) {
	if bs.autoChecker != nil {
		go bs.autoChecker.Start(ctx)
	}
}
