package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/dialogstate"
	"cws/internal/telegram/messaging"
	"cws/logger"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	bot        *tgbotapi.BotAPI
	stateMgr   *StateManager
	msgSender  messaging.MessageSender
	dialogHdlr *DialogHandler
	clientHdlr *ClientHandler
	cmdHdlr    *CommandHandler
}

func NewCallbackHandler(bot *tgbotapi.BotAPI, stateMgr *StateManager, msgSender messaging.MessageSender, dialogHdlr *DialogHandler, clientHdlr *ClientHandler) *CallbackHandler {
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

func (ch *CallbackHandler) parseClientIDAndGetTorrentCache(chatId int64, clientIDStr string) (int64, *TorrentFileCache, bool) {
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

		return 0, nil, false
	}

	cache, exists := ch.clientHdlr.torrentFilesCache[chatId]
	if !exists || cache == nil || cache.ExistingHash == "" {
		logger.Warn("Кэш торрент файла не найден или hash отсутствует для пользователя %d", chatId)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorTorrentDataNotFoundStartOver), nil)

		return clientID, nil, false
	}

	return clientID, cache, true
}

func (ch *CallbackHandler) HandleCallbackQuery(query *tgbotapi.CallbackQuery) {
	chatId := query.Message.Chat.ID
	data := query.Data

	logger.Debug("Обработка callback от пользователя %d: %s", chatId, data)

	callback := tgbotapi.NewCallback(query.ID, "")
	_, _ = ch.bot.Request(callback)

	switch {
	case data == "monitor_torrent_stop":
		ch.clientHdlr.torrentMonitorSvc.StopTorrentMonitoring(chatId)
		ch.clientHdlr.ShowClientsForTorrentMonitor(chatId)
	case data == "back_to_torrents":
		ch.clientHdlr.torrentMonitorSvc.StopTorrentMonitoring(chatId)

		if state, ok := ch.stateMgr.GetUserState(chatId); ok && strings.HasPrefix(state, string(dialogstate.StateMonitorTorrent)+"_") {
			clientIDStr := strings.TrimPrefix(state, string(dialogstate.StateMonitorTorrent)+"_")
			if clientID, err := strconv.ParseInt(clientIDStr, 10, 64); err == nil {
				ch.clientHdlr.StartTorrentMonitorDialog(chatId, clientID)

				return
			}
		}

		ch.clientHdlr.ShowClientsForTorrentMonitor(chatId)
	case data == "main_menu":
		ch.clientHdlr.torrentMonitorSvc.StopTorrentMonitoring(chatId)
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowMainMenu(chatId)
		}
	case data == "check_torrents":
		ch.clientHdlr.CheckAllClients(chatId)
	case data == "quick_actions":
		ch.clientHdlr.ShowQuickActionsMenu(chatId)
	case data == "quick_action_pause_menu":
		ch.clientHdlr.ShowPauseTorrentsMenu(chatId)
	case data == "quick_action_resume_menu":
		ch.clientHdlr.ShowResumeTorrentsMenu(chatId)
	case data == "quick_action_pause_all":
		ch.clientHdlr.HandlePauseAllTorrents(chatId)
	case data == "quick_action_resume_all":
		ch.clientHdlr.HandleResumeAllTorrents(chatId)
	case data == "quick_action_pause_rutracker":
		ch.clientHdlr.HandlePauseRutrackerTorrents(chatId)
	case data == "quick_action_resume_rutracker":
		ch.clientHdlr.HandleResumeRutrackerTorrents(chatId)
	case data == "quick_action_pause_non_rutracker":
		ch.clientHdlr.HandlePauseNonRutrackerTorrents(chatId)
	case data == "quick_action_resume_non_rutracker":
		ch.clientHdlr.HandleResumeNonRutrackerTorrents(chatId)
	case data == "quick_action_limit_speed_menu":
		ch.clientHdlr.ShowSpeedLimitMenu(chatId)
	case strings.HasPrefix(data, "quick_action_limit_speed_"):
		ch.handleSpeedLimitSelection(chatId, data)
	case data == "quick_action_limit_speed_custom":
		ch.clientHdlr.StartCustomSpeedLimitDialog(chatId)
	case data == "quick_action_remove_speed_limits":
		ch.clientHdlr.HandleRemoveSpeedLimits(chatId)
	case data == "add_torrent_file":
		ch.clientHdlr.ShowClientsForTorrentAdd(chatId)
	case data == "monitor_torrent":
		ch.clientHdlr.ShowClientsForTorrentMonitor(chatId)
	case data == "settings":
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowSettingsMenu(chatId)
		}
	case data == "toggle_notifications":
		ctx := context.Background()
		enabled, err := ch.clientHdlr.repo.GetNotificationsEnabled(ctx, chatId)
		if err != nil {
			logger.Warn("Failed to read notifications_enabled for user %d: %v", chatId, err)
			enabled = true
		}
		if setErr := ch.clientHdlr.repo.SetNotificationsEnabled(ctx, chatId, !enabled); setErr != nil {
			logger.Warn("Failed to set notifications_enabled for user %d: %v", chatId, setErr)
		}
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowSettingsMenu(chatId)
		}
	case data == "variables":
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowVariablesMenu(chatId)
		}
	case data == "edit_timezone":
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowTimezoneMenu(chatId, 0)
		}
	case strings.HasPrefix(data, "edit_timezone_page_"):
		if ch.cmdHdlr != nil {
			pageStr := strings.TrimPrefix(data, "edit_timezone_page_")
			page, err := strconv.Atoi(pageStr)
			if err != nil {
				logger.Warn("Пользователь %d отправил неверный номер страницы таймзон: %s", chatId, pageStr)

				return
			}
			ch.cmdHdlr.ShowTimezoneMenu(chatId, page)
		}
	case data == string(dialogstate.StateEditRecommended):
		ch.stateMgr.SetUserState(chatId, string(dialogstate.StateEditRecommended))
		text := ui.Msg(ui.MsgVariablesRecommendedTorrentsPrompt)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.ButtonWithData(ui.Cancel, "variables"),
			),
		)
		messageID := ch.stateMgr.GetDialogMessage(chatId)
		newMessageID, _ := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if newMessageID != 0 {
			ch.stateMgr.SetDialogMessage(chatId, newMessageID)
		}
	case data == "search_torrent":
		ch.clientHdlr.torrentSearchSvc.StartTorrentSearchDialog(chatId)
	case strings.HasPrefix(data, "search_torrent_select_"):
		ch.handleSearchTorrentSelect(chatId, data)
	case strings.HasPrefix(data, "page_search_"):
		pageStr := strings.TrimPrefix(data, "page_search_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			logger.Warn("Пользователь %d отправил неверный номер страницы поиска: %s", chatId, pageStr)

			return
		}
		logger.Debugf("Пользователь %d запросил страницу %d результатов поиска", chatId, page)
		ch.clientHdlr.torrentSearchSvc.ShowSearchResultsPage(chatId, page)
	case strings.HasPrefix(data, "monitor_torrent_page_"):
		rest := strings.TrimPrefix(data, "monitor_torrent_page_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) != 2 {
			logger.Warn("Пользователь %d отправил неверный формат страницы мониторинга: %s", chatId, data)

			return
		}
		clientIDStr := parts[0]
		pageStr := parts[1]
		clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
		if err != nil {
			logger.Warn("Пользователь %d отправил неверный ID клиента для страницы мониторинга: %s", chatId, clientIDStr)

			return
		}
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			logger.Warn("Пользователь %d отправил неверный номер страницы мониторинга: %s", chatId, pageStr)

			return
		}
		logger.Debugf("Пользователь %d запросил страницу %d списка торрентов для клиента %d", chatId, page, clientID)
		ch.clientHdlr.ShowTorrentMonitorPage(chatId, clientID, page)
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
		ch.stateMgr.SetDialogMessage(chatId, 0)
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.HandleClientsCommand(chatId)
		}
	case data == "cancel_edit_client":
		logger.Debugf("Пользователь %d отменил редактирование клиента", chatId)
		ch.stateMgr.DeleteUserState(chatId)
		ch.stateMgr.SetDialogMessage(chatId, 0)
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.HandleClientsCommand(chatId)
		}
	case data == "set_ssl_true":
		ch.dialogHdlr.FinishEditClient(chatId, true)
	case data == "set_ssl_false":
		ch.dialogHdlr.FinishEditClient(chatId, false)
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
	case data == "page_info":

		return
	case strings.HasPrefix(data, "page_missing_"):
		pageStr := strings.TrimPrefix(data, "page_missing_")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			logger.Warn("Пользователь %d отправил неверный номер страницы: %s", chatId, pageStr)

			return
		}
		logger.Debugf("Пользователь %d запросил страницу %d мёртвых торрентов", chatId, page)
		ch.clientHdlr.ShowMissingTorrentsPage(chatId, page)
	case strings.HasPrefix(data, "add_torrent_client_"):
		ch.handleAddTorrentClient(chatId, data)
	case strings.HasPrefix(data, "select_save_path_"):
		ch.handleSelectSavePath(chatId, data)
	case data == "cancel_add_torrent":
		ch.handleCancelAddTorrent(chatId)
	case strings.HasPrefix(data, "custom_save_path_"):
		ch.handleCustomSavePath(chatId, data)
	case strings.HasPrefix(data, "skip_hash_check_"):
		ch.handleSkipHashCheck(chatId, data)
	case strings.HasPrefix(data, "delete_existing_torrent_"):
		ch.handleDeleteExistingTorrent(chatId, data)
	case strings.HasPrefix(data, "keep_existing_torrent_"):
		ch.handleKeepExistingTorrent(chatId, data)
	case strings.HasPrefix(data, "confirm_delete_torrent_"):
		ch.handleConfirmDeleteTorrent(chatId, data)
	case strings.HasPrefix(data, "monitor_torrent_client_"):
		ch.handleMonitorTorrentClient(chatId, data)
	case strings.HasPrefix(data, "monitor_torrent_start_"):
		ch.handleMonitorTorrentStart(chatId, data)
	case strings.HasPrefix(data, "monitor_from_missing_"):
		hash := strings.TrimPrefix(data, "monitor_from_missing_")
		hash = strings.TrimSpace(hash)
		if hash != "" {
			ch.clientHdlr.ShowClientsForTorrentMonitorWithHash(chatId, hash)
		}
	case strings.HasPrefix(data, "monitor_torrent_hash_btn_"):
		ch.handleMonitorTorrentHashButton(chatId, data)
	case strings.HasPrefix(data, "monitor_torrent_manual_"):
		ch.handleMonitorTorrentManual(chatId, data)
	case strings.HasPrefix(data, "monitor_pause_"):
		ch.handleMonitorPause(chatId, data)
	case strings.HasPrefix(data, "monitor_resume_"):
		ch.handleMonitorResume(chatId, data)
	case strings.HasPrefix(data, "monitor_delete_confirm_"):
		ch.handleMonitorDeleteConfirm(chatId, data)
	case strings.HasPrefix(data, "monitor_delete_cancel_"):
		ch.handleMonitorDeleteCancel(chatId, data)
	case strings.HasPrefix(data, "monitor_delete_"):
		ch.handleMonitorDelete(chatId, data)
	case data == "edit_recommended_torrents":
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowEditRecommendedTorrents(chatId)
		}
	case strings.HasPrefix(data, "set_recommended_torrents_"):
		ch.handleSetRecommendedTorrents(chatId, data)
	case strings.HasPrefix(data, "set_timezone_"):
		ch.handleSetTimezone(chatId, data)
	}
}

func (ch *CallbackHandler) handleMonitorTorrentClient(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "monitor_torrent_client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

		return
	}
	logger.Debugf("Пользователь %d выбрал клиента %d для мониторинга торрента", chatId, clientID)
	ch.clientHdlr.StartTorrentMonitorDialog(chatId, clientID)
}

func (ch *CallbackHandler) handleMonitorTorrentStart(chatId int64, data string) {
	rest := strings.TrimPrefix(data, "monitor_torrent_start_")
	parts := strings.SplitN(rest, "_", 2)

	clientIDStr := ""
	hash := ""
	if len(parts) == 2 {
		clientIDStr = parts[0]
		hash = strings.TrimSpace(parts[1])
	} else if len(parts) == 1 {
		clientIDStr = parts[0]
		cached, exists := ch.clientHdlr.monitorStartHashCache[chatId]
		if exists {
			hash = strings.TrimSpace(cached)
		}
	} else {
		logger.Warn("Invalid monitor start callback format from user %d: %s", chatId, data)

		return
	}

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid client ID in monitor start callback from user %d: %s", chatId, clientIDStr)

		return
	}
	if hash == "" {
		return
	}

	ctx := context.Background()
	delete(ch.clientHdlr.torrentMonitorCache, chatId)
	ch.clientHdlr.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, hash)
}

func (ch *CallbackHandler) handleMonitorTorrentHashButton(chatId int64, data string) {
	prefix := "monitor_torrent_hash_btn_"
	if !strings.HasPrefix(data, prefix) {
		logger.Warn("Пользователь %d отправил неверный формат выбора хеша: %s", chatId, data)

		return
	}

	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		logger.Warn("Пользователь %d отправил неверный формат выбора хеша: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	indexStr := parts[1]

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)

		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный индекс торрента: %s", chatId, indexStr)

		return
	}

	cache, exists := ch.clientHdlr.torrentMonitorCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID {
		logger.Warn("Кэш торрентов для мониторинга не найден для пользователя %d", chatId)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorDataNotFoundStartOver), nil)

		return
	}

	if index < 0 || index >= len(cache.Torrents) {
		logger.Warn("Неверный индекс торрента %d для пользователя %d (всего: %d)", index, chatId, len(cache.Torrents))
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidTorrentIndex), nil)

		return
	}

	hash := cache.Torrents[index].Hash
	logger.Debugf("Пользователь %d выбрал торрент для мониторинга, клиент: %d, hash: %s", chatId, clientID, hash)

	delete(ch.clientHdlr.torrentMonitorCache, chatId)

	ctx := context.Background()
	ch.clientHdlr.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, hash)
}

func (ch *CallbackHandler) handleMonitorTorrentManual(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "monitor_torrent_manual_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

		return
	}
	logger.Debugf("Пользователь %d выбрал ручной ввод хеша для клиента %d", chatId, clientID)
	ch.clientHdlr.StartTorrentMonitorManualInput(chatId, clientID)
}

func (ch *CallbackHandler) handleMonitorPause(chatId int64, data string) {
	prefix := "monitor_pause_"
	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		logger.Warn("Invalid monitor pause callback format from user %d: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	hash := parts[1]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid client ID in monitor pause callback from user %d: %s", chatId, clientIDStr)

		return
	}

	ctx := context.Background()
	qbClient, _, ok := ch.clientHdlr.getQbClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return
	}

	if err = qbClient.PauseTorrent(ctx, hash); err != nil {
		logger.Error("Error pausing torrent %s for user %d: %v", hash, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorPauseTorrent), nil)

		return
	}
}

func (ch *CallbackHandler) handleMonitorResume(chatId int64, data string) {
	prefix := "monitor_resume_"
	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		logger.Warn("Invalid monitor resume callback format from user %d: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	hash := parts[1]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid client ID in monitor resume callback from user %d: %s", chatId, clientIDStr)

		return
	}

	ctx := context.Background()
	qbClient, _, ok := ch.clientHdlr.getQbClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return
	}

	if err = qbClient.ResumeTorrent(ctx, hash); err != nil {
		logger.Error("Error resuming torrent %s for user %d: %v", hash, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorResumeTorrent), nil)

		return
	}
}

func (ch *CallbackHandler) handleMonitorDelete(chatId int64, data string) {
	prefix := "monitor_delete_"
	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		logger.Warn("Invalid monitor delete callback format from user %d: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	hash := strings.TrimSpace(parts[1])
	if hash == "" {
		return
	}

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid client ID in monitor delete callback from user %d: %s", chatId, clientIDStr)

		return
	}

	ch.clientHdlr.torrentMonitorSvc.StopTorrentMonitoring(chatId)

	text := ui.Msg(ui.MsgDeleteFilesQuestionText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.DeleteFilesYes, fmt.Sprintf("monitor_delete_confirm_%d_%s_true", clientID, hash)),
			ui.ButtonWithData(ui.DeleteFilesNo, fmt.Sprintf("monitor_delete_confirm_%d_%s_false", clientID, hash)),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, fmt.Sprintf("monitor_delete_cancel_%d_%s", clientID, hash)),
		),
	)

	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if sendErr != nil {
		logger.Error("Error sending monitor delete confirmation to user %d: %v", chatId, sendErr)

		return
	}
	if newMessageID != 0 && newMessageID != messageID {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

func (ch *CallbackHandler) handleMonitorDeleteCancel(chatId int64, data string) {
	prefix := "monitor_delete_cancel_"
	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		logger.Warn("Invalid monitor delete cancel callback format from user %d: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	hash := strings.TrimSpace(parts[1])
	if hash == "" {
		return
	}

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid client ID in monitor delete cancel callback from user %d: %s", chatId, clientIDStr)

		return
	}

	ctx := context.Background()
	ch.clientHdlr.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, hash)
}

func (ch *CallbackHandler) handleMonitorDeleteConfirm(chatId int64, data string) {
	prefix := "monitor_delete_confirm_"
	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 3)
	if len(parts) != 3 {
		logger.Warn("Invalid monitor delete confirm callback format from user %d: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	hash := strings.TrimSpace(parts[1])
	deleteFilesStr := strings.TrimSpace(parts[2])
	if hash == "" || deleteFilesStr == "" {
		return
	}

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid client ID in monitor delete confirm callback from user %d: %s", chatId, clientIDStr)

		return
	}

	deleteFiles, err := strconv.ParseBool(deleteFilesStr)
	if err != nil {
		logger.Warn("Invalid deleteFiles flag in monitor delete confirm callback from user %d: %s", chatId, deleteFilesStr)

		return
	}

	ctx := context.Background()
	qbClient, client, ok := ch.clientHdlr.getQbClientByIDOrReply(ctx, chatId, clientID)
	if ok {
	} else {
		return
	}

	err = qbClient.DeleteTorrent(ctx, hash, deleteFiles)
	if err != nil {
		logger.Error("Error deleting torrent %s for user %d: %v", hash, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msgs(ui.MsgErrorDeleteTorrentFmt, err), nil)

		return
	}

	filesText := ""
	if deleteFiles {
		filesText = ui.Msg(ui.MsgTorrentDeletedFilesSuffix)
	}
	text := ui.Msgs(ui.MsgTorrentDeletedSuccessFmt, filesText, client.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.BackToTorrents),
			ui.Button(ui.MainMenu),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, _ := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if newMessageID != 0 && newMessageID != messageID {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

func (ch *CallbackHandler) handleSetRecommendedTorrents(chatId int64, data string) {
	prefix := "set_recommended_torrents_"
	if !strings.HasPrefix(data, prefix) {
		return
	}
	nStr := strings.TrimPrefix(data, prefix)
	n, err := strconv.Atoi(nStr)
	if err != nil {
		logger.Warn("Invalid recommended torrents value from user %d: %s", chatId, nStr)

		return
	}

	ctx := context.Background()
	if err = ch.clientHdlr.repo.SetRecommendedTorrents(ctx, chatId, n); err != nil {
		logger.Error("Failed to save recommended torrents for user %d: %v", chatId, err)

		return
	}

	if ch.cmdHdlr != nil {
		ch.cmdHdlr.ShowVariablesMenu(chatId)
	}
}

func (ch *CallbackHandler) handleSetTimezone(chatId int64, data string) {
	prefix := "set_timezone_"
	if !strings.HasPrefix(data, prefix) {
		return
	}
	repr := strings.TrimPrefix(data, prefix)
	tz := strings.ReplaceAll(repr, "|", "/")

	ctx := context.Background()
	if err := ch.clientHdlr.repo.SetUserTimezone(ctx, chatId, tz); err != nil {
		logger.Error("Ошибка при сохранении часового пояса для пользователя %d: %v", chatId, err)

		return
	}

	if ch.cmdHdlr != nil {
		ch.cmdHdlr.ShowVariablesMenu(chatId)
	}
}

func (ch *CallbackHandler) handleSearchTorrentSelect(chatId int64, data string) {
	indexStr := strings.TrimPrefix(data, "search_torrent_select_")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный индекс результата поиска: %s", chatId, indexStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidResultIndex), nil)

		return
	}

	result, err := ch.clientHdlr.torrentSearchSvc.GetSearchResult(chatId, index)
	if err != nil {
		logger.Warn("Ошибка при получении результата поиска для пользователя %d: %v", chatId, err)

		return
	}
	logger.Debugf("Пользователь %d выбрал торрент для мониторинга из результатов поиска: клиент %d, hash: %s", chatId, result.ClientID, result.Hash)

	ch.clientHdlr.torrentSearchSvc.ClearSearchCache(chatId)

	ctx := context.Background()
	ch.clientHdlr.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, result.ClientID, result.Hash)
}

func (ch *CallbackHandler) handleAddTorrentClient(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "add_torrent_client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

		return
	}
	logger.Debugf("Пользователь %d выбрал клиента %d для добавления торрента", chatId, clientID)
	ch.clientHdlr.StartAddTorrentDialog(chatId, clientID)
}

func (ch *CallbackHandler) handleSelectSavePath(chatId int64, data string) {
	parts := strings.Split(data, "_")
	if len(parts) < 4 {
		logger.Warn("Пользователь %d отправил неверный формат выбора пути: %s", chatId, data)

		return
	}
	clientIDStr := parts[3]
	pathIndexStr := parts[4]

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)

		return
	}

	pathIndex, err := strconv.Atoi(pathIndexStr)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный индекс пути: %s", chatId, pathIndexStr)

		return
	}

	logger.Debugf("Пользователь %d выбрал путь сохранения %d для клиента %d", chatId, pathIndex, clientID)
	ch.clientHdlr.HandleSavePathSelection(chatId, clientID, pathIndex)
}

func (ch *CallbackHandler) handleCancelAddTorrent(chatId int64) {
	logger.Debugf("Пользователь %d отменил добавление торрента", chatId)
	ch.clientHdlr.CancelAddTorrent(chatId)
}

func (ch *CallbackHandler) handleCustomSavePath(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "custom_save_path_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)

		return
	}
	logger.Debugf("Пользователь %d выбрал ввод пути вручную для клиента %d", chatId, clientID)
	ch.clientHdlr.StartCustomSavePathDialog(chatId, clientID)
}

func (ch *CallbackHandler) handleSkipHashCheck(chatId int64, data string) {
	parts := strings.Split(data, "_")
	if len(parts) < 5 {
		logger.Warn("Пользователь %d отправил неверный формат выбора пропуска проверки хеша: %s", chatId, data)

		return
	}
	clientIDStr := parts[3]
	skipStr := parts[4]

	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)

		return
	}

	skipHashCheck := skipStr == "true"
	logger.Debugf("Пользователь %d выбрал пропуск проверки хеша: %v для клиента %d", chatId, skipHashCheck, clientID)
	ch.clientHdlr.HandleSkipHashCheckSelection(chatId, clientID, skipHashCheck)
}

func (ch *CallbackHandler) handleDeleteExistingTorrent(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "delete_existing_torrent_")
	clientID, cache, ok := ch.parseClientIDAndGetTorrentCache(chatId, clientIDStr)
	if !ok {
		return
	}

	logger.Debugf("Пользователь %d выбрал удаление существующего торрента для клиента %d, hash: %s", chatId, clientID, cache.ExistingHash)
	ch.clientHdlr.ShowDeleteFilesQuestion(chatId, clientID, cache.ExistingHash)
}

func (ch *CallbackHandler) handleKeepExistingTorrent(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "keep_existing_torrent_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)

		return
	}
	logger.Debugf("Пользователь %d решил оставить существующий торрент для клиента %d", chatId, clientID)
	ch.clientHdlr.HandleKeepExistingTorrent(chatId, clientID)
}

func (ch *CallbackHandler) handleConfirmDeleteTorrent(chatId int64, data string) {
	prefix := "confirm_delete_torrent_"
	if !strings.HasPrefix(data, prefix) {
		logger.Warn("Пользователь %d отправил неверный формат подтверждения удаления: %s", chatId, data)

		return
	}

	rest := strings.TrimPrefix(data, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		logger.Warn("Пользователь %d отправил неверный формат подтверждения удаления: %s", chatId, data)

		return
	}

	clientIDStr := parts[0]
	deleteFilesStr := parts[1]

	clientID, cache, ok := ch.parseClientIDAndGetTorrentCache(chatId, clientIDStr)
	if !ok {
		return
	}

	deleteFiles := deleteFilesStr == "true"
	logger.Debugf("Пользователь %d подтвердил удаление торрента для клиента %d, hash: %s, deleteFiles: %v", chatId, clientID, cache.ExistingHash, deleteFiles)
	ch.clientHdlr.HandleDeleteExistingTorrent(chatId, clientID, cache.ExistingHash, deleteFiles)
}

func (ch *CallbackHandler) handleClientDetails(chatId int64, data string) {
	clientIDStr := strings.TrimPrefix(data, "client_")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверный ID клиента: %s", chatId, clientIDStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

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
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidClientID), nil)

		return
	}
	logger.Debugf("Пользователь %d запросил повторную проверку активных торрентов для клиента %d", chatId, clientID)
	ch.clientHdlr.CheckClientTorrents(chatId, clientID)
}

func (ch *CallbackHandler) handleSpeedLimitSelection(chatId int64, data string) {
	speedStr := strings.TrimPrefix(data, "quick_action_limit_speed_")
	speedMBx100, err := strconv.ParseInt(speedStr, 10, 64)
	if err != nil {
		logger.Warn("Пользователь %d отправил неверное значение скорости: %s", chatId, speedStr)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorInvalidSpeedValue), nil)

		return
	}
	speedMB := float64(speedMBx100) / 100.0
	logger.Debugf("Пользователь %d выбрал ограничение скорости: %.2f МБ/с", chatId, speedMB)
	speedBytesPerSec := int64(speedMB * 1024 * 1024)
	ch.clientHdlr.HandleLimitSpeedBytes(chatId, speedBytesPerSec)
}
