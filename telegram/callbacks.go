package telegram

import (
	"cws/logger"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	bot        *tgbotapi.BotAPI
	stateMgr   *StateManager
	msgSender  *MessageSender
	dialogHdlr *DialogHandler
	clientHdlr *ClientHandler
	cmdHdlr    *CommandHandler
}

func NewCallbackHandler(bot *tgbotapi.BotAPI, stateMgr *StateManager, msgSender *MessageSender, dialogHdlr *DialogHandler, clientHdlr *ClientHandler) *CallbackHandler {
	return &CallbackHandler{
		bot:        bot,
		stateMgr:   stateMgr,
		msgSender:  msgSender,
		dialogHdlr: dialogHdlr,
		clientHdlr: clientHdlr,
	}
}

func (ch *CallbackHandler) SetCommandHandler(cmdHdlr *CommandHandler) {
	ch.cmdHdlr = cmdHdlr
}

func (ch *CallbackHandler) HandleCallbackQuery(query *tgbotapi.CallbackQuery) {
	chatId := query.Message.Chat.ID
	data := query.Data

	logger.Debug("Обработка callback от пользователя %d: %s", chatId, data)

	callback := tgbotapi.NewCallback(query.ID, "")
	ch.bot.Request(callback)

	switch {
	case data == "main_menu":
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowMainMenu(chatId)
		}
	case data == "check_torrents":
		ch.clientHdlr.CheckAllClients(chatId)
	case data == "add_client":
		logger.Debugf("Пользователь %d начал добавление нового клиента", chatId)
		ch.dialogHdlr.StartAddClientDialog(chatId)
	case strings.HasPrefix(data, "check_client_"):
		ch.handleCheckClient(chatId, data)
	case strings.HasPrefix(data, "recheck_client_"):
		ch.handleRecheckClient(chatId, data)
	case strings.HasPrefix(data, "client_"):
		ch.handleClientDetails(chatId, data)
	case data == "cancel_add_client":
		logger.Debugf("Пользователь %d отменил добавление клиента", chatId)
		ch.stateMgr.DeleteUserState(chatId)
		ch.stateMgr.SetDialogMessage(chatId, 0) // Очищаем message_id диалога
		msg := tgbotapi.NewMessage(chatId, "Добавление клиента отменено")
		ch.msgSender.Send(msg)
	case data == "cancel_edit_client":
		logger.Debugf("Пользователь %d отменил редактирование клиента", chatId)
		ch.stateMgr.DeleteUserState(chatId)
		ch.stateMgr.SetDialogMessage(chatId, 0) // Очищаем message_id диалога
		msg := tgbotapi.NewMessage(chatId, "Редактирование клиента отменено")
		ch.msgSender.Send(msg)
	case data == "set_ssl_true":
		ch.dialogHdlr.FinishAddClient(chatId, true)
	case data == "set_ssl_false":
		ch.dialogHdlr.FinishAddClient(chatId, false)
	case data == "clients":
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.HandleClientsCommand(chatId)
		}
	case strings.HasPrefix(data, "delete_client_"):
		ch.handleDeleteClient(chatId, data)
	case strings.HasPrefix(data, "confirm_delete_"):
		ch.handleConfirmDelete(chatId, data)
	case strings.HasPrefix(data, "edit_client_"):
		ch.handleEditClient(chatId, data)
	case data == "set_edit_ssl_true":
		ch.dialogHdlr.FinishEditClient(chatId, true)
	case data == "set_edit_ssl_false":
		ch.dialogHdlr.FinishEditClient(chatId, false)
	}
}

func (ch *CallbackHandler) handleClientDetails(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		msg := tgbotapi.NewMessage(chatId, "Ошибка: неверный ID клиента")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d запросил детали клиента %d", chatId, clientID)
	ch.clientHdlr.ShowClientDetails(chatId, clientID)
}

func (ch *CallbackHandler) handleDeleteClient(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "delete_client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		msg := tgbotapi.NewMessage(chatId, "Ошибка: неверный ID клиента")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d запросил удаление клиента %d", chatId, clientID)
	ch.clientHdlr.ShowDeleteConfirmation(chatId, clientID)
}

func (ch *CallbackHandler) handleConfirmDelete(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "confirm_delete_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		msg := tgbotapi.NewMessage(chatId, "Ошибка: неверный ID клиента")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d подтвердил удаление клиента %d", chatId, clientID)
	ch.clientHdlr.DeleteClient(chatId, clientID)
}

func (ch *CallbackHandler) handleEditClient(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "edit_client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		msg := tgbotapi.NewMessage(chatId, "Ошибка: неверный ID клиента")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d начал редактирование клиента %d", chatId, clientID)
	ch.dialogHdlr.StartEditClientDialog(chatId, clientID)
}

func (ch *CallbackHandler) handleCheckClient(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "check_client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		msg := tgbotapi.NewMessage(chatId, "Ошибка: неверный ID клиента")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d запросил проверку активных торрентов для клиента %d", chatId, clientID)
	ch.clientHdlr.CheckClientTorrents(chatId, clientID)
}

func (ch *CallbackHandler) handleRecheckClient(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "recheck_client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		msg := tgbotapi.NewMessage(chatId, "Ошибка: неверный ID клиента")
		ch.msgSender.Send(msg)
		return
	}
	logger.Debugf("Пользователь %d запросил повторную проверку активных торрентов для клиента %d", chatId, clientID)
	ch.clientHdlr.CheckClientTorrents(chatId, clientID)
}
