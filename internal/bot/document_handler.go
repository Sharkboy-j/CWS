package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/dialogstate"
	"cws/internal/telegram/messaging"
	"cws/logger"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DocumentHandler struct {
	stateMgr   *StateManager
	msgSender  messaging.MessageSender
	clientHdlr *ClientHandler
}

func NewDocumentHandler(stateMgr *StateManager, msgSender messaging.MessageSender, clientHdlr *ClientHandler) *DocumentHandler {
	return &DocumentHandler{
		stateMgr:   stateMgr,
		msgSender:  msgSender,
		clientHdlr: clientHdlr,
	}
}

func (dh *DocumentHandler) HandleDocument(ctx context.Context, chatId int64, document *tgbotapi.Document, fileData []byte) {
	state, exists := dh.stateMgr.GetUserState(chatId)
	if !exists || !strings.HasPrefix(state, string(dialogstate.StateAddTorrentWait)+"_") {
		logger.Debug("Пользователь %d отправил файл, но не в процессе добавления торрента", chatId)

		return
	}

	prefix := string(dialogstate.StateAddTorrentWait) + "_"
	clientIDStr := strings.TrimPrefix(state, prefix)
	if clientIDStr == state {
		logger.Warn("Неверный формат состояния для добавления торрента: %s", state)

		return
	}
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Error("Ошибка при парсинге ID клиента из состояния: %v", err)

		return
	}

	if !strings.HasSuffix(strings.ToLower(document.FileName), ".torrent") {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorSendTorrentFilePrompt), nil)

		return
	}

	logger.Debugf("Пользователь %d отправил торрент файл %s для клиента %d", chatId, document.FileName, clientID)

	text := ui.Msgf(ui.MsgDocumentProcessingTorrentFileFmt, document.FileName)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "cancel_add_torrent"),
		),
	)
	messageID := dh.stateMgr.GetMenuMessage(chatId)
	_, err = dh.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Warn("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
	}

	dh.clientHdlr.HandleTorrentFileReceived(ctx, chatId, clientID, fileData, document.FileName)
}
