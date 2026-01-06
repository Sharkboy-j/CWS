package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/internal/textutil"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processRutrackerResults(ctx context.Context, qbClient qbit.Service, torrentByHash map[string]qbittorrent.Torrent, rutrackerResults map[string]*int) (int, []missingTorrentInfo) {
	foundCount := 0
	for _, topicID := range rutrackerResults {
		if topicID != nil {
			foundCount++
		}
	}

	var missingTorrents []missingTorrentInfo

	if rutrackerResults != nil {
		for hash, torrent := range torrentByHash {
			topicID, exists := rutrackerResults[hash]
			if !exists || topicID == nil {
				props, err := qbClient.GetTorrentPropertiesCached(ctx, torrent.Hash)
				url := ""
				if err == nil && props != nil {
					url = textutil.ExtractURLFromComment(props.Comment)
				}
				missingTorrents = append(missingTorrents, missingTorrentInfo{
					name: torrent.Name,
					hash: hash,
					url:  url,
				})
			}
		}
	} else {
		for hash, torrent := range torrentByHash {
			props, err := qbClient.GetTorrentPropertiesCached(ctx, torrent.Hash)
			url := ""
			if err == nil && props != nil {
				url = textutil.ExtractURLFromComment(props.Comment)
			}
			missingTorrents = append(missingTorrents, missingTorrentInfo{
				name: torrent.Name,
				hash: hash,
				url:  url,
			})
		}
	}

	return foundCount, missingTorrents
}

func sendNoClientsMessage(msgSender messaging.MessageSender, stateMgr *StateManager, chatId int64, text string) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.AddClient),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.MainMenu),
		),
	)
	messageID := stateMgr.GetMenuMessage(chatId)
	newMessageID, err := msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return err
	}
	stateMgr.SetMenuMessage(chatId, newMessageID)

	return nil
}

func (ch *ClientHandler) getClientByIDWithErrorHandling(chatId int64, clientID int64) (*storage.Client, bool) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorGetClientData), nil)

		return nil, false
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался получить доступ к несуществующему клиенту %d", chatId, clientID)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorClientNotFoundOrNoAccess), nil)

		return nil, false
	}

	return client, true
}

func (ch *ClientHandler) getClientByIDOrReply(ctx context.Context, chatId int64, clientID int64) (*storage.Client, bool) {
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorGetClientDataWithEmoji), nil)

		return nil, false
	}

	return client, true
}

func (ch *ClientHandler) getQbClientByIDOrReply(ctx context.Context, chatId int64, clientID int64) (qbit.Service, *storage.Client, bool) {
	client, ok := ch.getClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return nil, nil, false
	}

	qbClient, err := qbit.New(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту %s для пользователя %d: %v", client.Name, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msgf(ui.MsgErrorConnectClientFmt, client.Name), nil)

		return nil, nil, false
	}

	return qbClient, client, true
}

func (ch *ClientHandler) finalizeTorrentFlow(ctx context.Context, chatId int64, clientID int64, newTorrentHash string) {
	ch.stateMgr.DeleteUserState(chatId)
	delete(ch.torrentFilesCache, chatId)

	if newTorrentHash != "" {
		logger.Debug("Запуск мониторинга торрента для пользователя %d, hash: %s", chatId, newTorrentHash)
		ch.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, newTorrentHash)

		return
	}

	logger.Warn("Hash не получен, переход в главное меню для пользователя %d", chatId)
	if ch.cmdHdlr != nil {
		ch.cmdHdlr.ShowMainMenu(chatId)
	}
}
