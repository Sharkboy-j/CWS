package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/dialogstate"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/logger"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DialogHandler struct {
	repo       *storage.Repository
	msgSender  messaging.MessageSender
	stateMgr   *StateManager
	clientHdlr *ClientHandler
	cmdHdlr    *CommandHandler
}

func NewDialogHandler(repo *storage.Repository, msgSender messaging.MessageSender, stateMgr *StateManager, clientHdlr *ClientHandler) *DialogHandler {
	return &DialogHandler{
		repo:       repo,
		msgSender:  msgSender,
		stateMgr:   stateMgr,
		clientHdlr: clientHdlr,
	}
}

func (dh *DialogHandler) SetCommandHandler(cmdHdlr *CommandHandler) {
	dh.cmdHdlr = cmdHdlr
}

func (dh *DialogHandler) HandleMessage(message *tgbotapi.Message) {
	chatId := message.Chat.ID
	messageID := message.MessageID
	state, exists := dh.stateMgr.GetUserState(chatId)
	if !exists {
		logger.Debug("Пользователь %d отправил сообщение, но не в процессе диалога", chatId)

		return
	}

	text := message.Text
	if text == "" {
		logger.Debug("Пользователь %d отправил пустое сообщение", chatId)

		return
	}

	logger.Debug("Обработка сообщения от пользователя %d в состоянии: %s, текст: %s", chatId, state, text)

	const separator = "|||"

	defer dh.msgSender.DeleteMessage(chatId, messageID)

	switch {
	case strings.HasPrefix(state, string(dialogstate.StateAddClientName)), strings.HasPrefix(state, string(dialogstate.StateEditClientName)):
		dh.handleClientName(chatId, text, state, separator)
	case strings.HasPrefix(state, string(dialogstate.StateAddClientHost)), strings.HasPrefix(state, string(dialogstate.StateEditClientHost)):
		dh.handleClientHost(chatId, text, state, separator)
	case strings.HasPrefix(state, string(dialogstate.StateAddClientPort)), strings.HasPrefix(state, string(dialogstate.StateEditClientPort)):
		dh.handleClientPort(chatId, text, state, separator)
	case strings.HasPrefix(state, string(dialogstate.StateAddClientUsername)), strings.HasPrefix(state, string(dialogstate.StateEditClientUsername)):
		dh.handleClientUsername(chatId, text, state, separator)
	case strings.HasPrefix(state, string(dialogstate.StateAddClientPassword)), strings.HasPrefix(state, string(dialogstate.StateEditClientPassword)):
		dh.handleClientPassword(chatId, text, state, separator)

	case strings.HasPrefix(state, string(dialogstate.StateAddTorrentCustom)+"_"):
		dh.handleAddTorrentCustomPath(chatId, text, state)
	case strings.HasPrefix(state, string(dialogstate.StateMonitorTorrent)+"_"):
		dh.handleMonitorTorrentHash(chatId, text, state)
	case state == string(dialogstate.StateSearchTorrent):
		dh.handleTorrentSearchQuery(chatId, text)
	case state == string(dialogstate.StateCustomSpeedLimit):
		dh.handleCustomSpeedLimit(chatId, text)
	case state == string(dialogstate.StateEditRecommended):
		dh.handleEditRecommendedTorrentsInput(chatId, text)
	default:
		logger.Warn("Неизвестное состояние для пользователя %d: %s, текст: %s", chatId, state, text)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogUnknownStateStartOver), nil)
	}
}

type clientDialogMode int

const (
	clientDialogModeAdd clientDialogMode = iota
	clientDialogModeEdit
)

type clientStatePair struct {
	add  dialogstate.State
	edit dialogstate.State
}

func (p clientStatePair) state(mode clientDialogMode) dialogstate.State {
	if mode == clientDialogModeEdit {
		return p.edit
	}

	return p.add
}

var (
	clientStateNamePair = clientStatePair{add: dialogstate.StateAddClientName, edit: dialogstate.StateEditClientName}
	clientStateHostPair = clientStatePair{add: dialogstate.StateAddClientHost, edit: dialogstate.StateEditClientHost}
	clientStatePortPair = clientStatePair{add: dialogstate.StateAddClientPort, edit: dialogstate.StateEditClientPort}
	clientStateUserPair = clientStatePair{add: dialogstate.StateAddClientUsername, edit: dialogstate.StateEditClientUsername}
	clientStatePassPair = clientStatePair{add: dialogstate.StateAddClientPassword, edit: dialogstate.StateEditClientPassword}
	clientStateSSLPair  = clientStatePair{add: dialogstate.StateAddClientSSL, edit: dialogstate.StateEditClientSSL}
)

func (m clientDialogMode) prefix() string {
	if m == clientDialogModeEdit {
		return "edit_client"
	}

	return "add_client"
}

func (m clientDialogMode) invalidStateMsg() ui.MsgID {
	if m == clientDialogModeEdit {
		return ui.MsgDialogInvalidStateEditStartOver
	}

	return ui.MsgDialogInvalidStateAddStartOver
}

func (m clientDialogMode) sslCallbacks() (string, string) {
	if m == clientDialogModeEdit {
		return "set_edit_ssl_true", "set_edit_ssl_false"
	}

	return "set_ssl_true", "set_ssl_false"
}

func buildClientState(mode clientDialogMode, pair clientStatePair, separator string, parts ...string) string {
	base := string(pair.state(mode))
	if len(parts) == 0 {
		return base
	}

	return base + separator + strings.Join(parts, separator)
}

func parseClientState(state string, pair clientStatePair, separator string) (clientDialogMode, []string, bool) {
	addPrefix := string(pair.add)
	editPrefix := string(pair.edit)

	switch {
	case strings.HasPrefix(state, addPrefix):
		return clientDialogModeAdd, extractStateParts(state, addPrefix, separator), true
	case strings.HasPrefix(state, editPrefix):
		return clientDialogModeEdit, extractStateParts(state, editPrefix, separator), true
	default:
		return clientDialogModeAdd, nil, false
	}
}

func detectClientMode(state string) clientDialogMode {
	if strings.HasPrefix(state, "edit_client_") {
		return clientDialogModeEdit
	}

	return clientDialogModeAdd
}

func extractStateParts(state, prefix, separator string) []string {
	remainder := strings.TrimPrefix(state, prefix)
	remainder = strings.TrimPrefix(remainder, separator)
	if remainder == "" {
		return nil
	}

	return strings.Split(remainder, separator)
}

func (dh *DialogHandler) cancelButton(mode clientDialogMode) tgbotapi.InlineKeyboardButton {
	callback := "cancel_add_client"
	if mode == clientDialogModeEdit {
		callback = "cancel_edit_client"
	}

	return ui.ButtonWithData(ui.Cancel, callback)
}

func (dh *DialogHandler) cancelKeyboard(mode clientDialogMode) *tgbotapi.InlineKeyboardMarkup {
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(dh.cancelButton(mode)),
	)

	return &markup
}

func (dh *DialogHandler) sendDialogPrompt(chatId int64, text string, mode clientDialogMode, keyboard *tgbotapi.InlineKeyboardMarkup) bool {
	if keyboard == nil {
		keyboard = dh.cancelKeyboard(mode)
	} else {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(dh.cancelButton(mode)))
	}

	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, text, keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return false
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)

	return true
}

func (dh *DialogHandler) resetClientDialog(chatId int64, mode clientDialogMode, state string, pair clientStatePair) {
	logger.Warn("Неверный формат состояния %s для пользователя %d: %s, шаг: %s", mode.prefix(), chatId, state, pair.state(mode))
	dh.stateMgr.DeleteUserState(chatId)
	_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(mode.invalidStateMsg()), nil)
}

func (dh *DialogHandler) handleClientName(chatId int64, text, state, separator string) {
	mode, parts, ok := parseClientState(state, clientStateNamePair, separator)
	if !ok {
		dh.resetClientDialog(chatId, detectClientMode(state), state, clientStateNamePair)

		return
	}

	logger.Debugf("Пользователь %d ввел имя клиента: %s", chatId, text)

	nextParts := []string{text}
	if mode == clientDialogModeEdit {
		if len(parts) < 2 {
			dh.resetClientDialog(chatId, mode, state, clientStateNamePair)

			return
		}
		nextParts = []string{parts[0], text}
	}

	dh.stateMgr.SetUserState(chatId, buildClientState(mode, clientStateHostPair, separator, nextParts...))
	dh.sendDialogPrompt(chatId, ui.Msg(ui.MsgDialogEnterHost), mode, nil)
}

func (dh *DialogHandler) handleClientHost(chatId int64, text, state, separator string) {
	mode, parts, ok := parseClientState(state, clientStateHostPair, separator)
	if !ok {
		dh.resetClientDialog(chatId, detectClientMode(state), state, clientStateHostPair)

		return
	}

	minParts := 1
	if mode == clientDialogModeEdit {
		minParts = 2
	}
	if len(parts) < minParts {
		dh.resetClientDialog(chatId, mode, state, clientStateHostPair)

		return
	}

	logger.Debugf("Пользователь %d ввел host: %s", chatId, text)

	nextParts := make([]string, 0, len(parts)+1)
	if mode == clientDialogModeEdit {
		nextParts = append(nextParts, parts[0])
	}
	nextParts = append(nextParts, parts[len(parts)-1], text)

	dh.stateMgr.SetUserState(chatId, buildClientState(mode, clientStatePortPair, separator, nextParts...))
	dh.sendDialogPrompt(chatId, ui.Msg(ui.MsgDialogEnterPort), mode, nil)
}

func (dh *DialogHandler) handleClientPort(chatId int64, text, state, separator string) {
	mode, parts, ok := parseClientState(state, clientStatePortPair, separator)
	if !ok {
		dh.resetClientDialog(chatId, detectClientMode(state), state, clientStatePortPair)

		return
	}

	minParts := 2
	if mode == clientDialogModeEdit {
		minParts = 3
	}
	if len(parts) < minParts {
		dh.resetClientDialog(chatId, mode, state, clientStatePortPair)

		return
	}

	port, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		messageText := ui.Msg(ui.MsgDialogPortMustBeNumberTryAgain)
		messageID := dh.stateMgr.GetDialogMessage(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)

		return
	}
	logger.Debugf("Пользователь %d ввел port: %d", chatId, port)

	nextParts := make([]string, 0, len(parts)+1)
	if mode == clientDialogModeEdit {
		nextParts = append(nextParts, parts[0])
	}
	nextParts = append(nextParts, parts[len(parts)-2], parts[len(parts)-1], strconv.FormatInt(port, 10))

	dh.stateMgr.SetUserState(chatId, buildClientState(mode, clientStateUserPair, separator, nextParts...))
	dh.sendDialogPrompt(chatId, ui.Msg(ui.MsgDialogEnterUsername), mode, nil)
}

func (dh *DialogHandler) handleClientUsername(chatId int64, text, state, separator string) {
	mode, parts, ok := parseClientState(state, clientStateUserPair, separator)
	if !ok {
		dh.resetClientDialog(chatId, detectClientMode(state), state, clientStateUserPair)

		return
	}

	minParts := 3
	if mode == clientDialogModeEdit {
		minParts = 4
	}
	if len(parts) < minParts {
		dh.resetClientDialog(chatId, mode, state, clientStateUserPair)

		return
	}

	logger.Debugf("Пользователь %d ввел username: %s", chatId, text)

	nextParts := make([]string, 0, len(parts)+1)
	if mode == clientDialogModeEdit {
		nextParts = append(nextParts, parts[0])
	}
	nextParts = append(nextParts, parts[len(parts)-3], parts[len(parts)-2], parts[len(parts)-1], text)

	dh.stateMgr.SetUserState(chatId, buildClientState(mode, clientStatePassPair, separator, nextParts...))
	dh.sendDialogPrompt(chatId, ui.Msg(ui.MsgDialogEnterPassword), mode, nil)
}

func (dh *DialogHandler) handleClientPassword(chatId int64, text, state, separator string) {
	mode, parts, ok := parseClientState(state, clientStatePassPair, separator)
	if !ok {
		dh.resetClientDialog(chatId, detectClientMode(state), state, clientStatePassPair)

		return
	}

	minParts := 4
	if mode == clientDialogModeEdit {
		minParts = 5
	}
	if len(parts) < minParts {
		dh.resetClientDialog(chatId, mode, state, clientStatePassPair)

		return
	}

	logger.Debugf("Пользователь %d ввел password", chatId)

	nextParts := append(append([]string{}, parts...), text)
	dh.stateMgr.SetUserState(chatId, buildClientState(mode, clientStateSSLPair, separator, nextParts...))

	messageText := ui.Msg(ui.MsgDialogUseSSL)
	yesCallback, noCallback := mode.sslCallbacks()
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Yes, yesCallback),
			ui.ButtonWithData(ui.No, noCallback),
		),
	)
	dh.sendDialogPrompt(chatId, messageText, mode, &keyboard)
}

func (dh *DialogHandler) StartEditClientDialog(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := dh.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для редактирования пользователем %d: %v", clientID, chatId, err)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorGetClientData), nil)

		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался редактировать несуществующий клиент %d", chatId, clientID)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorClientNotFoundOrNoAccess), nil)

		return
	}

	const separator = "|||"
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("%s%s%d%s%s", dialogstate.StateEditClientName, separator, clientID, separator, client.Name))
	messageText := ui.Msgf(ui.MsgDialogEditClientStartFmt, client.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "cancel_edit_client"),
		),
	)
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	if messageID == 0 {
		messageID = dh.stateMgr.GetMenuMessage(chatId)
	}
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) StartAddClientDialog(chatId int64) {
	dh.stateMgr.SetUserState(chatId, string(dialogstate.StateAddClientName))
	text := ui.Msg(ui.MsgDialogAddClientStart)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "cancel_add_client"),
		),
	)
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	if messageID == 0 {
		messageID = dh.stateMgr.GetMenuMessage(chatId)
	}
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) FinishEditClient(chatId int64, ssl bool) {
	const separator = "|||"

	state, exists := dh.stateMgr.GetUserState(chatId)
	if !exists {
		logger.Warn("Состояние не найдено или неверный префикс для пользователя %d: exists=%v, state=%s", chatId, exists, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogSessionExpiredStartOver), nil)

		return
	}

	var clientID, port int64
	var clientName, host, username, password string
	var err error
	ctx := context.Background()
	var client *storage.Client

	logger.Debug("Обработка завершения редактирования для пользователя %d, состояние: %s", chatId, state)
	parts := strings.Split(state, separator)
	if strings.HasPrefix(state, string(dialogstate.StateEditClientSSL)) {
		clientID, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			dh.stateMgr.DeleteUserState(chatId)
			_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidClientIDStartOver), nil)

			return
		}

		clientName = parts[2]
		host = parts[3]
		portStr := parts[4]
		port, err = strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			dh.stateMgr.DeleteUserState(chatId)
			_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidPortStartOver), nil)

			return
		}
		username = parts[5]
		password = parts[6]

	} else if strings.HasPrefix(state, string(dialogstate.StateAddClientSSL)) {
		clientName = parts[1]
		host = parts[2]
		portStr := parts[3]
		port, err = strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			dh.stateMgr.DeleteUserState(chatId)
			_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidPortStartOver), nil)

			return
		}
		username = parts[4]
		password = parts[5]
	} else {
		logger.Warn("Неверный формат состояния %s для пользователя %d: %s (частей: %d, ожидается 7)", dialogstate.StateEditClientSSL, chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidDataStartOver), nil)

		return
	}

	client = &storage.Client{
		ID:       clientID,
		UserID:   chatId,
		Name:     clientName,
		Host:     host,
		Port:     int32(port),
		Username: username,
		Password: password,
		SSL:      ssl,
	}

	if client.ID != 0 {
		err = dh.repo.UpdateClient(ctx, client, chatId)
		if err != nil {
			logger.Error("Ошибка при обновлении/добавлении клиента %d для пользователя %d: %v", clientID, chatId, err)
			_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogUpdateClientErrorTryAgain), nil)
			dh.stateMgr.DeleteUserState(chatId)

			return
		}

		logger.Debugf("Пользователь %d успешно обновил клиента: ID=%d, Name=%s, Host=%s:%d",
			chatId, clientID, clientName, host, port)
	} else {
		client, err = dh.repo.CreateClient(ctx, client)
		if err != nil {
			logger.Error("Ошибка при создании клиента для пользователя %d: %v", chatId, err)
			_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogCreateClientErrorTryAgain), nil)
			dh.stateMgr.DeleteUserState(chatId)

			return
		}

		logger.Debugf("Пользователь %d успешно создал клиента: ID=%d, Name=%s, Host=%s:%d",
			chatId, client.ID, client.Name, client.Host, client.Port)
	}

	dh.stateMgr.DeleteUserState(chatId)

	dialogMessageID := dh.stateMgr.GetDialogMessage(chatId)
	if dialogMessageID > 0 {
		dh.stateMgr.SetMenuMessage(chatId, dialogMessageID)
	}
	dh.stateMgr.SetDialogMessage(chatId, 0)

	if dh.cmdHdlr != nil {
		dh.cmdHdlr.HandleClientsCommand(chatId)
	} else {
		logger.Warn("Command handler is not initialized, unable to refresh clients list for user %d", chatId)
	}
}

func (dh *DialogHandler) handleAddTorrentCustomPath(chatId int64, text, state string) {
	logger.Debugf("Пользователь %d ввел путь сохранения: %s", chatId, text)

	parts := strings.Split(state, "_")
	if len(parts) < 5 {
		logger.Warn("Неверный формат состояния %s для пользователя %d: %s", dialogstate.StateAddTorrentCustom, chatId, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidStateStartOver), nil)

		return
	}
	clientIDStr := parts[4]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Error("Ошибка при парсинге ID клиента: %v", err)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidClientID), nil)

		return
	}

	cache, exists := dh.clientHdlr.torrentFilesCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID {
		logger.Warn("Кэш торрент файла не найден для пользователя %d", chatId)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgErrorTorrentDataNotFoundStartOver), nil)

		return
	}

	cache.SelectedPath = text
	dh.clientHdlr.torrentFilesCache[chatId] = cache

	dh.clientHdlr.ShowSkipHashCheckQuestion(chatId, clientID, text)
}

func (dh *DialogHandler) handleMonitorTorrentHash(chatId int64, text, state string) {
	logger.Debugf("Пользователь %d ввел хеш торрента для мониторинга: %s", chatId, text)

	parts := strings.Split(state, "_")
	if len(parts) < 4 {
		logger.Warn("Неверный формат состояния %s для пользователя %d: %s", dialogstate.StateMonitorTorrent, chatId, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidStateStartOver), nil)

		return
	}
	clientIDStr := parts[3]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Неверный ID клиента в состоянии %s для пользователя %d: %s", dialogstate.StateMonitorTorrent, chatId, clientIDStr)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogInvalidClientIDStartOver), nil)

		return
	}

	hash := strings.TrimSpace(strings.ToUpper(text))

	if hash == "" {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogMonitorHashEmptyPrompt), nil)

		return
	}

	if len(hash) != 40 {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogMonitorHashLengthPrompt), nil)

		return
	}

	dh.stateMgr.DeleteUserState(chatId)

	ctx := context.Background()
	dh.clientHdlr.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, hash)
}

func (dh *DialogHandler) handleTorrentSearchQuery(chatId int64, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogSearchQueryEmptyPrompt), nil)

		return
	}

	dh.stateMgr.DeleteUserState(chatId)

	dh.clientHdlr.torrentSearchSvc.SearchTorrents(chatId, query)
}

func (dh *DialogHandler) handleCustomSpeedLimit(chatId int64, speedText string) {
	speedText = strings.TrimSpace(speedText)
	if speedText == "" {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogSpeedEmptyPrompt), nil)

		return
	}

	speedMB, err := strconv.ParseFloat(speedText, 64)
	if err != nil {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogSpeedInvalidFormatPrompt), nil)

		return
	}

	if speedMB <= 0 {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogSpeedMustBePositivePrompt), nil)

		return
	}

	dh.stateMgr.DeleteUserState(chatId)

	speedBytesPerSec := int64(speedMB * 1024 * 1024)
	dh.clientHdlr.HandleLimitSpeedBytes(chatId, speedBytesPerSec)
}

func (dh *DialogHandler) handleEditRecommendedTorrentsInput(chatId int64, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		messageText := ui.Msg(ui.MsgDialogRecommendedEmptyPrompt)
		messageID := dh.stateMgr.GetDialogMessage(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)

		return
	}

	n, err := strconv.Atoi(text)
	if err != nil || n <= 0 || n > 100 {
		messageText := ui.Msg(ui.MsgDialogRecommendedInvalidPrompt)
		messageID := dh.stateMgr.GetDialogMessage(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)

		return
	}

	ctx := context.Background()
	if err = dh.repo.SetRecommendedTorrents(ctx, chatId, n); err != nil {
		logger.Error("Ошибка при сохранении recommended_torrents для пользователя %d: %v", chatId, err)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgDialogRecommendedSaveErrorTryAgain), nil)

		return
	}

	dialogMessageID := dh.stateMgr.GetDialogMessage(chatId)
	if dialogMessageID > 0 {
		dh.msgSender.DeleteMessage(chatId, dialogMessageID)
		dh.stateMgr.SetDialogMessage(chatId, 0)
	}
	dh.stateMgr.DeleteUserState(chatId)

	if dh.cmdHdlr != nil {
		dh.cmdHdlr.ShowVariablesMenu(chatId)
	}
}
