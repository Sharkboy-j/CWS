package bot

import (
	"context"
	"cws/config"
	"cws/internal/bot/monitoring"
	"cws/internal/bot/quick_actions"
	"cws/internal/bot/ui"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
	"fmt"
	"time"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ClientHandler struct {
	repo                 *storage.Repository
	msgSender            messaging.MessageSender
	notifySender         messaging.MessageSender
	stateMgr             *StateManager
	cmdHdlr              *CommandHandler
	cfg                  *config.Config
	mainBotUsername      string
	missingTorrentsCache map[int64][]missingTorrentInfo
	checkResultsCache    map[int64]*CheckResultsCache
	torrentFilesCache    map[int64]*TorrentFileCache
	torrentMonitorCache  map[int64]*TorrentMonitorCache
	torrentMonitorSvc    monitoring.TorrentMonitorService
	torrentSearchSvc     *TorrentSearchService
	quickActionsHandler  *quick_actions.Handler
}

type TorrentMonitorCache struct {
	ClientID int64
	Torrents []TorrentMonitorItem
}

type TorrentMonitorItem struct {
	Hash string
	Name string
}

type TorrentFileCache struct {
	ClientID     int64
	FileName     string
	FileData     []byte
	SavePaths    []string
	DefaultPath  string
	ExistingPath string
	TorrentName  string
	SelectedPath string
	ExistingHash string
}

type CheckResultsCache struct {
	Results            []ClientCheckResult
	TotalDuration      time.Duration
	LastCheckTime      *time.Time
	AllMissingTorrents []missingTorrentInfo
}

func NewClientHandler(repo *storage.Repository, msgSender messaging.MessageSender, notifySender messaging.MessageSender, stateMgr *StateManager, cfg *config.Config, mainBotUsername string) *ClientHandler {
	quickActionsHandler := quick_actions.NewHandler(repo, msgSender, &stateManagerAdapter{stateMgr: stateMgr})

	return &ClientHandler{
		repo:                 repo,
		msgSender:            msgSender,
		notifySender:         notifySender,
		stateMgr:             stateMgr,
		cfg:                  cfg,
		mainBotUsername:      mainBotUsername,
		missingTorrentsCache: make(map[int64][]missingTorrentInfo),
		checkResultsCache:    make(map[int64]*CheckResultsCache),
		torrentFilesCache:    make(map[int64]*TorrentFileCache),
		torrentMonitorCache:  make(map[int64]*TorrentMonitorCache),
		torrentMonitorSvc:    monitoring.NewTorrentMonitorService(repo, msgSender, stateMgr.GetMenuMessage, stateMgr.SetMenuMessage),
		torrentSearchSvc:     NewTorrentSearchService(repo, msgSender, stateMgr),
		quickActionsHandler:  quickActionsHandler,
	}
}

type stateManagerAdapter struct {
	stateMgr *StateManager
}

func (s *stateManagerAdapter) GetMenuMessage(chatId int64) int {
	return s.stateMgr.GetMenuMessage(chatId)
}

func (s *stateManagerAdapter) SetMenuMessage(chatId int64, messageID int) {
	s.stateMgr.SetMenuMessage(chatId, messageID)
}

func (ch *ClientHandler) SetCommandHandler(cmdHdlr *CommandHandler) {
	ch.cmdHdlr = cmdHdlr
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.SetCommandHandler(&commandHandlerAdapter{cmdHdlr: cmdHdlr})
		ch.quickActionsHandler.SetStateSetter(&stateSetterAdapter{stateMgr: ch.stateMgr})
	}
}

type commandHandlerAdapter struct {
	cmdHdlr *CommandHandler
}

func (c *commandHandlerAdapter) ShowMainMenu(chatId int64) {
	c.cmdHdlr.ShowMainMenu(chatId)
}

type stateSetterAdapter struct {
	stateMgr *StateManager
}

func (s *stateSetterAdapter) SetUserState(chatId int64, state string) {
	s.stateMgr.SetUserState(chatId, state)
}

func (ch *ClientHandler) ShowClientDetails(chatId int64, clientID int64) {
	client, ok := ch.getClientByIDWithErrorHandling(chatId, clientID)
	if !ok {
		return
	}

	sslText := "Нет"
	if client.SSL {
		sslText = "Да"
	}

	text := fmt.Sprintf("🔧 *%s*\n\n"+
		"Host: `%s`\n"+
		"Port: `%d`\n"+
		"Username: `%s`\n"+
		"SSL: `%s`\n",
		client.Name, client.Host, client.Port, client.Username, sslText)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Edit, fmt.Sprintf("edit_client_%d", clientID)),
			ui.ButtonWithData(ui.Delete, fmt.Sprintf("delete_client_%d", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.BackToList),
			ui.Button(ui.MainMenu),
		),
	)

	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) ShowDeleteConfirmation(chatId int64, clientID int64) {
	client, ok := ch.getClientByIDWithErrorHandling(chatId, clientID)
	if !ok {
		return
	}

	text := fmt.Sprintf("⚠️ *Подтверждение удаления*\n\nВы уверены, что хотите удалить клиента *%s*?\n\nЭто действие нельзя отменить!", client.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.ConfirmDelete, fmt.Sprintf("confirm_delete_%d", clientID)),
			ui.ButtonWithData(ui.Cancel, fmt.Sprintf("client_%d", clientID)),
		),
	)

	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) DeleteClient(chatId int64, clientID int64) {
	client, ok := ch.getClientByIDWithErrorHandling(chatId, clientID)
	if !ok {
		return
	}

	ctx := context.Background()

	clientName := client.Name

	err := ch.repo.DeleteClient(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при удалении клиента %d для пользователя %d: %v", clientID, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "Ошибка при удалении клиента. Попробуйте снова.", nil)

		return
	}

	logger.Debugf("Пользователь %d успешно удалил клиента: ID=%d, Name=%s", chatId, clientID, clientName)

	if ch.cmdHdlr != nil {
		ch.cmdHdlr.HandleClientsCommand(chatId)
	}
}

func (ch *ClientHandler) ShowClientsForTorrentAdd(chatId int64) {
	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		errorText := "❌ Ошибка при получении списка клиентов"
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, errorText, nil)

		return
	}

	if len(clients) == 0 {
		errorText := "📥 *Добавление торрент файла*\n\nКлиенты не найдены. Добавьте клиента для загрузки торрента."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.AddClient),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := ch.stateMgr.GetMenuMessage(chatId)
		newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if sendErr == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}

		return
	}

	text := "📥 *Добавление торрент файла*\n\nВыберите клиент для загрузки торрента:"
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, client := range clients {
		sslText := "🔒"
		if !client.SSL {
			sslText = "🔓"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				fmt.Sprintf("%s %s", sslText, client.Name),
				fmt.Sprintf("add_torrent_client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.MainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err == nil {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

func (ch *ClientHandler) ShowClientsForTorrentMonitor(chatId int64) {
	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		errorText := "❌ Ошибка при получении списка клиентов"
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, errorText, nil)

		return
	}

	if len(clients) == 0 {
		errorText := "📊 *Мониторинг торрента*\n\nКлиенты не найдены. Добавьте клиента для мониторинга."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.AddClient),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := ch.stateMgr.GetMenuMessage(chatId)
		newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if sendErr == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}

		return
	}

	text := "📊 *Мониторинг торрента*\n\nВыберите клиент:"
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, client := range clients {
		sslText := "🔒"
		if !client.SSL {
			sslText = "🔓"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				fmt.Sprintf("%s %s", sslText, client.Name),
				fmt.Sprintf("monitor_torrent_client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.MainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err == nil {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

func (ch *ClientHandler) ShowClientsForTorrentMonitorWithHash(chatId int64, hash string) {
	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при получении списка клиентов", nil)

		return
	}

	if len(clients) == 0 {
		errorText := "📊 *Мониторинг торрента*\n\nКлиенты не найдены. Добавьте клиента для мониторинга."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.AddClient),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := ch.stateMgr.GetMenuMessage(chatId)
		newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if sendErr == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}

		return
	}

	text := "📊 *Мониторинг торрента*\n\nВыберите клиент:"
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, client := range clients {
		sslText := "🔒"
		if !client.SSL {
			sslText = "🔓"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				fmt.Sprintf("%s %s", sslText, client.Name),
				fmt.Sprintf("monitor_torrent_start_%d_%s", client.ID, hash),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.MainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err == nil {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

func (ch *ClientHandler) StartTorrentMonitorDialog(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при получении данных клиента", nil)

		return
	}

	qbClient, err := qbit.New(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту для мониторинга: %v", err)
	} else {
		torrents, torrentsErr := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
		if torrentsErr == nil && len(torrents) > 0 {
			sortedTorrents := make([]qbittorrent.Torrent, len(torrents))
			copy(sortedTorrents, torrents)

			for i := 0; i < len(sortedTorrents)-1; i++ {
				for j := i + 1; j < len(sortedTorrents); j++ {
					if sortedTorrents[i].AddedOn < sortedTorrents[j].AddedOn {
						sortedTorrents[i], sortedTorrents[j] = sortedTorrents[j], sortedTorrents[i]
					}
				}
			}
			// Build full list of monitor items (we'll paginate the view).
			var monitorItems []TorrentMonitorItem
			for _, torrent := range sortedTorrents {
				hash := torrent.InfohashV1
				if hash == "" {
					hash = torrent.InfohashV2
				}
				if hash != "" {
					monitorItems = append(monitorItems, TorrentMonitorItem{
						Hash: hash,
						Name: torrent.Name,
					})
				}
			}

			if len(monitorItems) > 0 {
				ch.torrentMonitorCache[chatId] = &TorrentMonitorCache{
					ClientID: clientID,
					Torrents: monitorItems,
				}

				// show first page (page 0)
				ch.stateMgr.SetUserState(chatId, fmt.Sprintf("monitor_torrent_hash_%d", clientID))
				ch.ShowTorrentMonitorPage(chatId, clientID, 0)

				return
			}
		}
	}

	ch.StartTorrentMonitorManualInput(chatId, clientID)
}

// ShowTorrentMonitorPage renders a page of torrents for monitoring selection.
func (ch *ClientHandler) ShowTorrentMonitorPage(chatId int64, clientID int64, page int) {
	cache, exists := ch.torrentMonitorCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID || len(cache.Torrents) == 0 {
		logger.Warn("Кэш торрентов для мониторинга не найден для пользователя %d", chatId)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: данные не найдены. Начните заново.", nil)

		return
	}

	ctx := context.Background()
	// determine page size from user's recommended torrents setting (default 3)
	pageSize := 3
	if cnt, err := ch.repo.GetRecommendedTorrents(ctx, chatId); err == nil && cnt > 0 {
		pageSize = cnt
	}

	total := len(cache.Torrents)
	maxPage := (total - 1) / pageSize
	if page < 0 {
		page = 0
	}
	if page > maxPage {
		page = maxPage
	}

	text := "📊 *Мониторинг торрента*\n\nВыберите торрент или введите хеш вручную:"
	var rows [][]tgbotapi.InlineKeyboardButton

	start := page * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		item := cache.Torrents[i]
		name := item.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}
		// use global index in callback so handler can find correct hash
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				fmt.Sprintf("📁 %s", name),
				fmt.Sprintf("monitor_torrent_hash_btn_%d_%d", clientID, i),
			),
		))
	}

	// pagination buttons
	var navRow []tgbotapi.InlineKeyboardButton
	if page > 0 {
		navRow = append(navRow, ui.ButtonWithData(ui.PrevPage, fmt.Sprintf("monitor_torrent_page_%d_%d", clientID, page-1)))
	}
	if page < maxPage {
		navRow = append(navRow, ui.ButtonWithData(ui.NextPage, fmt.Sprintf("monitor_torrent_page_%d_%d", clientID, page+1)))
	}
	if len(navRow) > 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(navRow...))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.ButtonWithData(ui.ManualHashInput, fmt.Sprintf("monitor_torrent_manual_%d", clientID)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.ButtonWithData(ui.Cancel, "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if sendErr != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, sendErr)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) StartTorrentMonitorManualInput(chatId int64, clientID int64) {
	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("monitor_torrent_hash_%d", clientID))
	text := "📊 *Мониторинг торрента*\n\nВведите хеш торрента для мониторинга:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "main_menu"),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {

		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {

		return fmt.Sprintf("%.2fs", d.Seconds())
	}

	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)

	return fmt.Sprintf("%dm %.2fs", minutes, seconds)
}

func (ch *ClientHandler) formatTimeInUserTimezone(ctx context.Context, chatId int64, t time.Time) string {
	timezone, err := ch.repo.GetUserTimezone(ctx, chatId)
	if err != nil {
		logger.Warn("Ошибка при получении часового пояса для пользователя %d: %v, используем Europe/Minsk", chatId, err)
		timezone = "Europe/Minsk"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		logger.Warn("Ошибка при загрузке локации %s для пользователя %d: %v, используем Europe/Minsk", timezone, chatId, err)
		loc, _ = time.LoadLocation("Europe/Minsk")
		if loc == nil {
			loc = time.UTC
		}
	}

	return t.In(loc).Format("02.01.2006 15:04:05")
}

func (ch *ClientHandler) StartAddTorrentDialog(chatId int64, clientID int64) {
	ctx := context.Background()
	client, ok := ch.getClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return
	}

	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("add_torrent_wait_file_%d", clientID))
	text := fmt.Sprintf("📥 *Добавление торрент файла*\n\nКлиент: *%s*\n\n📎 Отправьте торрент файл (.torrent):", client.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "cancel_add_torrent"),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) HandleTorrentFileReceived(ctx context.Context, chatId int64, clientID int64, fileData []byte, fileName string) {
	qbClient, _, ok := ch.getQbClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return
	}

	torrentInfo, err := qbit.ParseTorrentFile(fileData)
	if err != nil {
		logger.Warn("Ошибка при парсинге торрент файла: %v", err)
		torrentInfo = nil
	}

	var existingPath string
	var existingHash string
	var torrentName string
	if torrentInfo != nil {
		torrentName = torrentInfo.Name
		allTorrents, getErr := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
		if getErr == nil {
			existingTorrent := qbit.FindTorrentByName(allTorrents, torrentInfo.Name)
			if existingTorrent != nil && existingTorrent.SavePath != "" {
				existingPath = existingTorrent.SavePath
				existingHash = existingTorrent.InfohashV1
				if existingHash == "" {
					existingHash = existingTorrent.InfohashV2
				}
				logger.Info("Найден существующий торрент с таким же названием: %s, путь: %s, hash: %s", torrentInfo.Name, existingPath, existingHash)
			}
		}
	}

	savePaths, err := qbClient.GetTorrentSavePaths(ctx)
	if err != nil {
		logger.Warn("Ошибка при получении путей сохранения: %v", err)
		savePaths = []string{}
	}

	defaultPath, err := qbClient.GetDefaultSavePath(ctx)
	if err != nil {
		logger.Warn("Ошибка при получении пути по умолчанию: %v", err)
		defaultPath = ""
	}

	ch.torrentFilesCache[chatId] = &TorrentFileCache{
		ClientID:     clientID,
		FileName:     fileName,
		FileData:     fileData,
		SavePaths:    savePaths,
		DefaultPath:  defaultPath,
		ExistingPath: existingPath,
		ExistingHash: existingHash,
		TorrentName:  torrentName,
	}

	ch.ShowSavePathSelection(chatId, clientID, savePaths, defaultPath, existingPath, torrentName)
}

func (ch *ClientHandler) ShowSavePathSelection(chatId int64, clientID int64, savePaths []string, defaultPath string, existingPath string, torrentName string) {
	text := "📁 *Выберите путь сохранения*\n\n"

	if torrentName != "" {
		text += fmt.Sprintf("Торрент: `%s`\n\n", torrentName)
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	if existingPath != "" {
		text += fmt.Sprintf("⭐ *Рекомендуется* (используется для этого торрента):\n`%s`\n\n", existingPath)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				fmt.Sprintf("⭐ Рекомендуется: %s", truncatePath(existingPath, 999)),
				fmt.Sprintf("select_save_path_%d_%d", clientID, -2),
			),
		))
	}

	if defaultPath != "" && defaultPath != existingPath {
		text += fmt.Sprintf("По умолчанию: `%s`\n\n", defaultPath)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				fmt.Sprintf("📂 По умолчанию (%s)", truncatePath(defaultPath, 999)),
				fmt.Sprintf("select_save_path_%d_%d", clientID, -1),
			),
		))
	}

	if len(savePaths) > 0 {
		text += "Существующие пути:\n"
		shownCount := 0
		for i, path := range savePaths {
			if shownCount >= 10 {

				break
			}
			if path == existingPath || path == defaultPath {

				continue
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				ui.Data(
					fmt.Sprintf("📂 %s", truncatePath(path, 999)),
					fmt.Sprintf("select_save_path_%d_%d", clientID, i),
				),
			))
			shownCount++
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.ButtonWithData(ui.ManualPathInput, fmt.Sprintf("custom_save_path_%d", clientID)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.ButtonWithData(ui.Cancel, "cancel_add_torrent"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) HandleSavePathSelection(chatId int64, clientID int64, pathIndex int) {
	cache, exists := ch.torrentFilesCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID {
		logger.Warn("Кэш торрент файла не найден для пользователя %d", chatId)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: данные торрента не найдены. Начните заново.", nil)

		return
	}

	var savePath string
	if pathIndex == -2 {
		savePath = cache.ExistingPath
		if savePath == "" {
			logger.Warn("Путь из существующего торрента не найден")
			_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: путь не найден", nil)

			return
		}
	} else if pathIndex == -1 {
		savePath = cache.DefaultPath
	} else if pathIndex >= 0 && pathIndex < len(cache.SavePaths) {
		savePath = cache.SavePaths[pathIndex]
	} else {
		logger.Warn("Неверный индекс пути: %d", pathIndex)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: неверный путь", nil)

		return
	}

	cache.SelectedPath = savePath
	ch.torrentFilesCache[chatId] = cache

	ch.ShowSkipHashCheckQuestion(chatId, clientID, savePath)
}

func (ch *ClientHandler) StartCustomSavePathDialog(chatId int64, clientID int64) {
	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("add_torrent_custom_path_%d", clientID))
	text := "✏️ *Ввод пути сохранения*\n\nВведите путь для сохранения торрента:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "cancel_add_torrent"),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) ShowSkipHashCheckQuestion(chatId int64, clientID int64, savePath string) {
	text := fmt.Sprintf("⚙️ *Настройки добавления торрента*\n\nПуть сохранения: `%s`\n\n❓ Пропустить проверку хеша при добавлении?\n\n_Проверка хеша может занять время, но гарантирует целостность данных._", savePath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.SkipHashYes, fmt.Sprintf("skip_hash_check_%d_true", clientID)),
			ui.ButtonWithData(ui.SkipHashNo, fmt.Sprintf("skip_hash_check_%d_false", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "cancel_add_torrent"),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) CancelAddTorrent(chatId int64) {
	ch.stateMgr.DeleteUserState(chatId)
	delete(ch.torrentFilesCache, chatId)
	if ch.cmdHdlr != nil {
		ch.cmdHdlr.ShowMainMenu(chatId)
	}
}

func (ch *ClientHandler) AddTorrentToClient(ctx context.Context, chatId int64, clientID int64, fileData []byte, savePath string, skipHashCheck bool) {
	qbClient, _, ok := ch.getQbClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return
	}

	err := qbClient.AddTorrentFile(ctx, fileData, savePath, "", skipHashCheck)
	if err != nil {
		logger.Error("Ошибка при добавлении торрента: %v", err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, fmt.Sprintf("❌ Ошибка при добавлении торрента: %v", err), nil)

		return
	}

	cache, exists := ch.torrentFilesCache[chatId]
	if exists && cache != nil && cache.ExistingHash != "" {
		ch.ShowDeleteExistingTorrentQuestion(chatId, clientID, cache.ExistingHash, cache.TorrentName)

		return
	}

	var newTorrentHash string
	torrentInfo, err := qbit.ParseTorrentFile(fileData)
	if err == nil && torrentInfo != nil {
		newTorrentHash = torrentInfo.InfoHash
		logger.Debug("Извлечен hash нового торрента: %s", newTorrentHash)
	} else {
		logger.Warn("Не удалось извлечь hash из торрент файла: %v", err)
	}

	ch.finalizeTorrentFlow(ctx, chatId, clientID, newTorrentHash)
}

func (ch *ClientHandler) HandleSkipHashCheckSelection(chatId int64, clientID int64, skipHashCheck bool) {
	ctx := context.Background()
	cache, exists := ch.torrentFilesCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID {
		logger.Warn("Кэш торрент файла не найден для пользователя %d", chatId)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: данные торрента не найдены. Начните заново.", nil)

		return
	}

	if cache.SelectedPath == "" {
		logger.Warn("Выбранный путь не найден в кэше для пользователя %d", chatId)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: путь сохранения не выбран. Начните заново.", nil)

		return
	}

	ch.AddTorrentToClient(ctx, chatId, clientID, cache.FileData, cache.SelectedPath, skipHashCheck)
}

func (ch *ClientHandler) ShowDeleteExistingTorrentQuestion(chatId int64, clientID int64, existingHash string, torrentName string) {
	text := fmt.Sprintf("✅ \n\n⚠️ Найден существующий торрент с таким же названием:\n`%s`\n\n❓ Удалить старый торрент?", torrentName)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.ConfirmDelete, fmt.Sprintf("delete_existing_torrent_%d", clientID)),
			ui.ButtonWithData(ui.KeepExisting, fmt.Sprintf("keep_existing_torrent_%d", clientID)),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

// ShowDeleteFilesQuestion показывает вопрос об удалении файлов вместе с торрентом
func (ch *ClientHandler) ShowDeleteFilesQuestion(chatId int64, clientID int64, hash string) {
	text := "🗑️ *Удаление торрента*\n\n❓ Удалить файлы вместе с торрентом?\n\n_Если выбрать \"Да\", файлы будут удалены с диска._"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.DeleteFilesYes, fmt.Sprintf("confirm_delete_torrent_%d_true", clientID)),
			ui.ButtonWithData(ui.DeleteFilesNo, fmt.Sprintf("confirm_delete_torrent_%d_false", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, fmt.Sprintf("keep_existing_torrent_%d", clientID)),
		),
	)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) HandleDeleteExistingTorrent(chatId int64, clientID int64, hash string, deleteFiles bool) {
	ctx := context.Background()
	qbClient, client, ok := ch.getQbClientByIDOrReply(ctx, chatId, clientID)
	if !ok {
		return
	}

	err := qbClient.DeleteTorrent(ctx, hash, deleteFiles)
	if err != nil {
		logger.Error("Ошибка при удалении торрента: %v", err)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, fmt.Sprintf("❌ Ошибка при удалении торрента: %v", err), nil)

		return
	}

	ch.stateMgr.DeleteUserState(chatId)

	cache, exists := ch.torrentFilesCache[chatId]
	var newTorrentHash string
	if exists && cache != nil {
		torrentInfo, parseErr := qbit.ParseTorrentFile(cache.FileData)
		if parseErr == nil && torrentInfo != nil {
			newTorrentHash = torrentInfo.InfoHash
		}
	}
	delete(ch.torrentFilesCache, chatId)

	filesText := ""
	if deleteFiles {
		filesText = " и файлы"
	}
	text := fmt.Sprintf("✅ Торрент успешно удален%s из клиента *%s*", filesText, client.Name)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, _ := ch.msgSender.SendOrEdit(chatId, messageID, text, nil)
	if newMessageID > 0 {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}

	ch.finalizeTorrentFlow(ctx, chatId, clientID, newTorrentHash)
}

func (ch *ClientHandler) HandleKeepExistingTorrent(chatId int64, clientID int64) {
	ctx := context.Background()

	cache, exists := ch.torrentFilesCache[chatId]
	var newTorrentHash string
	if exists && cache != nil {
		torrentInfo, parseErr := qbit.ParseTorrentFile(cache.FileData)
		if parseErr == nil && torrentInfo != nil {
			newTorrentHash = torrentInfo.InfoHash
			logger.Debug("Извлечен hash нового торрента из кэша: %s", newTorrentHash)
		} else {
			logger.Warn("Не удалось извлечь hash из торрент файла в кэше: %v", parseErr)
		}
	}

	ch.finalizeTorrentFlow(ctx, chatId, clientID, newTorrentHash)
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {

		return path
	}

	return "..." + path[len(path)-maxLen+3:]
}

func (ch *ClientHandler) ShowQuickActionsMenu(chatId int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.ShowQuickActionsMenu(chatId)
	}
}

func (ch *ClientHandler) HandlePauseAllTorrents(chatId int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.HandlePauseAllTorrents(chatId)
	}
}

func (ch *ClientHandler) HandleResumeAllTorrents(chatId int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.HandleResumeAllTorrents(chatId)
	}
}

func (ch *ClientHandler) ShowSpeedLimitMenu(chatId int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.ShowSpeedLimitMenu(chatId)
	}
}

func (ch *ClientHandler) StartCustomSpeedLimitDialog(chatId int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.StartCustomSpeedLimitDialog(chatId)
	}
}

func (ch *ClientHandler) HandleLimitSpeedBytes(chatId int64, limitBytesPerSec int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.HandleLimitSpeedBytes(chatId, limitBytesPerSec)
	}
}

func (ch *ClientHandler) HandleRemoveSpeedLimits(chatId int64) {
	if ch.quickActionsHandler != nil {
		ch.quickActionsHandler.HandleRemoveSpeedLimits(chatId)
	}
}
