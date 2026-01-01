package telegram

import (
	"bytes"
	"context"
	"cws/config"
	"cws/database"
	"cws/logger"
	"encoding/json"
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot          *tgbotapi.BotAPI
	chatId       int64
	token        string
	repo         *database.Repository
	stateMgr     *StateManager
	msgSender    *MessageSender
	cmdHdlr      *CommandHandler
	callbackHdlr *CallbackHandler
	dialogHdlr   *DialogHandler
	clientHdlr   *ClientHandler
	autoChecker  *AutoChecker
}

func NewBotService(token string, chatId int64, repo *database.Repository, cfg *config.Config) (*BotService, error) {
	bot, err := InitBot(token)
	if err != nil {
		return nil, err
	}

	stateMgr, err := NewStateManager(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	msgSender := NewMessageSender(bot)

	service := &BotService{
		bot:       bot,
		chatId:    chatId,
		token:     token,
		repo:      repo,
		stateMgr:  stateMgr,
		msgSender: msgSender,
	}

	service.cmdHdlr = NewCommandHandler(bot, repo, stateMgr, msgSender)
	service.clientHdlr = NewClientHandler(repo, msgSender, stateMgr, cfg)
	service.dialogHdlr = NewDialogHandler(repo, msgSender, stateMgr, service.clientHdlr)
	service.callbackHdlr = NewCallbackHandler(bot, stateMgr, msgSender, service.dialogHdlr, service.clientHdlr)

	service.clientHdlr.SetCommandHandler(service.cmdHdlr)
	service.dialogHdlr.SetCommandHandler(service.cmdHdlr)
	service.callbackHdlr.SetCommandHandler(service.cmdHdlr)

	if !cfg.ManualCheckOnly && cfg.DurationSeconds > 0 {
		service.autoChecker = NewAutoChecker(service.clientHdlr, cfg.DurationSeconds)
	}

	err = service.setupCommands()
	if err != nil {
		logger.Warn("Не удалось установить команды бота: %v", err)
	}

	err = service.setupMenuButton()
	if err != nil {
		logger.Warn("Не удалось установить Menu Button: %v", err)
	}

	logger.Info("Бот инициализирован для пользователя %d", chatId)

	return service, nil
}

func (bs *BotService) StartAutoChecker(ctx context.Context) {
	if bs.autoChecker != nil {
		go bs.autoChecker.Start(ctx)
	}
}

func (bs *BotService) setupCommands() error {
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

	setCommands := tgbotapi.NewSetMyCommands(commands...)
	_, err := bs.bot.Request(setCommands)
	if err != nil {
		return err
	}

	logger.Info("Команды бота установлены успешно")
	return nil
}

func (bs *BotService) setupMenuButton() error {
	menuButton := map[string]interface{}{
		"type": "commands",
	}

	requestBody := map[string]interface{}{
		"chat_id":     bs.chatId,
		"menu_button": menuButton,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setChatMenuButton", bs.token)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	logger.Info("Menu Button установлен успешно")
	return nil
}

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

func (bs *BotService) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		chatId := update.CallbackQuery.Message.Chat.ID
		username := update.CallbackQuery.From.UserName
		logger.Info("Пользователь %d (@%s) нажал на кнопку: %s", chatId, username, update.CallbackQuery.Data)
		bs.callbackHdlr.HandleCallbackQuery(update.CallbackQuery)
		return
	}

	if update.Message != nil && update.Message.IsCommand() {
		chatId := update.Message.Chat.ID
		username := update.Message.From.UserName
		command := update.Message.Command()
		messageID := update.Message.MessageID
		logger.Info("Пользователь %d (@%s) выполнил команду: /%s", chatId, username, command)

		bs.cmdHdlr.HandleCommand(update.Message)

		bs.msgSender.DeleteMessage(chatId, messageID)
		return
	}

	if update.Message != nil {
		chatId := update.Message.Chat.ID
		username := update.Message.From.UserName
		logger.Debug("Пользователь %d (@%s) отправил сообщение: %s", chatId, username, update.Message.Text)
		bs.dialogHdlr.HandleMessage(update.Message)
	}
}
