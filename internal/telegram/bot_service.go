package telegram

import (
	"context"
	"cws/internal"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/logger"
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot          *tgbotapi.BotAPI
	chatId       int64
	token        string
	repo         *storage.Repository
	stateMgr     StateManager
	msgSender    messaging.MessageSender
	cmdHdlr      CommandHandler
	callbackHdlr CallbackHandler
	dialogHdlr   DialogHandler
	docHdlr      DocumentHandler
}

type BotServiceConfig struct {
	Token        string
	Repo         *storage.Repository
	StateMgr     StateManager
	MsgSender    messaging.MessageSender
	CmdHdlr      CommandHandler
	CallbackHdlr CallbackHandler
	DialogHdlr   DialogHandler
	DocHdlr      DocumentHandler
}

func NewBotService(cfg BotServiceConfig) (*BotService, error) {
	bot, err := initBot(cfg.Token)
	if err != nil {
		return nil, err
	}

	service := &BotService{
		bot:          bot,
		token:        cfg.Token,
		repo:         cfg.Repo,
		stateMgr:     cfg.StateMgr,
		msgSender:    cfg.MsgSender,
		cmdHdlr:      cfg.CmdHdlr,
		callbackHdlr: cfg.CallbackHdlr,
		dialogHdlr:   cfg.DialogHdlr,
		docHdlr:      cfg.DocHdlr,
	}

	return service, nil
}

func (bs *BotService) GetBot() *tgbotapi.BotAPI {
	return bs.bot
}

func (bs *BotService) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		chatId := update.CallbackQuery.Message.Chat.ID
		username := update.CallbackQuery.From.UserName
		logger.Debugf("Пользователь %d (@%s) нажал на кнопку: %s", chatId, username, update.CallbackQuery.Data)
		bs.callbackHdlr.HandleCallbackQuery(update.CallbackQuery)

		return
	}

	if update.Message != nil && update.Message.IsCommand() {
		chatId := update.Message.Chat.ID
		username := update.Message.From.UserName
		command := update.Message.Command()
		messageID := update.Message.MessageID
		logger.Debugf("Пользователь %d (@%s) выполнил команду: /%s", chatId, username, command)

		bs.cmdHdlr.HandleCommand(update.Message)

		bs.msgSender.DeleteMessage(chatId, messageID)

		return
	}

	if update.Message != nil {
		chatId := update.Message.Chat.ID
		username := update.Message.From.UserName
		logger.Debugf("Пользователь %d (@%s) отправил сообщение: %s", chatId, username, update.Message.Text)

		if update.Message.Location != nil {
			bs.handleUserLocation(chatId, update.Message.Location)
		}

		if update.Message.Document != nil {
			bs.handleDocument(chatId, update.Message)

			return
		}

		bs.dialogHdlr.HandleMessage(update.Message)
	}
}

func (bs *BotService) handleUserLocation(chatId int64, location *tgbotapi.Location) {
	ctx := context.Background()
	logger.Debugf("Пользователь %d отправил локацию: lat=%.6f, lon=%.6f", chatId, location.Latitude, location.Longitude)

	timezone := internal.DetermineTimezoneByCoordinates(location.Latitude, location.Longitude)

	err := bs.repo.SetUserTimezone(ctx, chatId, timezone)
	if err != nil {
		logger.Error("Ошибка при сохранении часового пояса для пользователя %d: %v", chatId, err)

		return
	}

	logger.Info("Часовой пояс пользователя %d установлен: %s", chatId, timezone)
	_, _ = bs.msgSender.SendOrEdit(chatId, 0, fmt.Sprintf("✅ Часовой пояс установлен: %s", timezone), nil)
}

func (bs *BotService) handleDocument(chatId int64, message *tgbotapi.Message) {
	ctx := context.Background()
	document := message.Document
	if document == nil {
		logger.Warn("Документ не найден в сообщении")

		return
	}

	fileURL, err := bs.bot.GetFileDirectURL(document.FileID)
	if err != nil {
		logger.Error("Ошибка при получении URL файла: %v", err)
		_, _ = bs.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при получении файла", nil)

		return
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		logger.Error("Ошибка при скачивании файла: %v", err)
		_, _ = bs.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при скачивании файла", nil)

		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Ошибка при чтении файла: %v", err)
		_, _ = bs.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при чтении файла", nil)

		return
	}

	if bs.docHdlr != nil {
		bs.docHdlr.HandleDocument(ctx, chatId, document, fileData)
	}

	bs.msgSender.DeleteMessage(chatId, message.MessageID)
}
