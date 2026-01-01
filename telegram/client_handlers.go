package telegram

import (
	"context"
	"cws/config"
	"cws/database"
	"cws/logger"
	"cws/qBit"
	"fmt"
	"time"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ClientHandler struct {
	repo                 *database.Repository
	msgSender            *MessageSender
	stateMgr             *StateManager
	cmdHdlr              *CommandHandler
	cfg                  *config.Config
	missingTorrentsCache map[int64][]missingTorrentInfo
	checkResultsCache    map[int64]*CheckResultsCache
	torrentFilesCache    map[int64]*TorrentFileCache
	torrentMonitorCache  map[int64]*TorrentMonitorCache // Кэш торрентов для выбора мониторинга
	torrentMonitorSvc    *TorrentMonitorService         // Сервис мониторинга торрентов
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
	ExistingPath string // Путь из существующего торрента с таким же хешем
	TorrentName  string // Название торрента
	SelectedPath string // Выбранный путь сохранения
	ExistingHash string // Hash существующего торрента (если найден)
}

type CheckResultsCache struct {
	Results            []ClientCheckResult
	TotalDuration      time.Duration
	LastCheckTime      *time.Time
	AllMissingTorrents []missingTorrentInfo
}

func NewClientHandler(repo *database.Repository, msgSender *MessageSender, stateMgr *StateManager, cfg *config.Config) *ClientHandler {
	return &ClientHandler{
		repo:                 repo,
		msgSender:            msgSender,
		stateMgr:             stateMgr,
		cfg:                  cfg,
		missingTorrentsCache: make(map[int64][]missingTorrentInfo),
		checkResultsCache:    make(map[int64]*CheckResultsCache),
		torrentFilesCache:    make(map[int64]*TorrentFileCache),
		torrentMonitorCache:  make(map[int64]*TorrentMonitorCache),
		torrentMonitorSvc:    NewTorrentMonitorService(repo, msgSender, stateMgr),
	}
}

func (ch *ClientHandler) SetCommandHandler(cmdHdlr *CommandHandler) {
	ch.cmdHdlr = cmdHdlr
}

func (ch *ClientHandler) ShowClientDetails(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался просмотреть несуществующий клиент %d", chatId, clientID)
		msg := tgbotapi.NewMessage(chatId, "Клиент не найден или у вас нет доступа")
		ch.msgSender.Send(msg)
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
			tgbotapi.NewInlineKeyboardButtonData("✏️ Изменить", fmt.Sprintf("edit_client_%d", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("delete_client_%d", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к списку", "clients"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
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
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для удаления пользователем %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался удалить несуществующий клиент %d", chatId, clientID)
		msg := tgbotapi.NewMessage(chatId, "Клиент не найден или у вас нет доступа")
		ch.msgSender.Send(msg)
		return
	}

	text := fmt.Sprintf("⚠️ *Подтверждение удаления*\n\nВы уверены, что хотите удалить клиента *%s*?\n\nЭто действие нельзя отменить!", client.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", fmt.Sprintf("confirm_delete_%d", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", fmt.Sprintf("client_%d", clientID)),
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
	ctx := context.Background()

	// Сначала получаем имя клиента для сообщения
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для удаления пользователем %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался удалить несуществующий клиент %d", chatId, clientID)
		msg := tgbotapi.NewMessage(chatId, "Клиент не найден или у вас нет доступа")
		ch.msgSender.Send(msg)
		return
	}

	clientName := client.Name

	// Удаляем клиента
	err = ch.repo.DeleteClient(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при удалении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "Ошибка при удалении клиента. Попробуйте снова.")
		ch.msgSender.Send(msg)
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
		msg := tgbotapi.NewMessage(chatId, errorText)
		ch.msgSender.Send(msg)
		return
	}

	if len(clients) == 0 {
		errorText := "📥 *Добавление торрент файла*\n\nКлиенты не найдены. Добавьте клиента для загрузки торрента."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := ch.stateMgr.GetMenuMessage(chatId)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if err == nil {
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
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", sslText, client.Name),
				fmt.Sprintf("add_torrent_client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err == nil {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

// ShowClientsForTorrentMonitor показывает список клиентов для выбора мониторинга торрента
func (ch *ClientHandler) ShowClientsForTorrentMonitor(chatId int64) {
	ctx := context.Background()
	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		errorText := "❌ Ошибка при получении списка клиентов"
		msg := tgbotapi.NewMessage(chatId, errorText)
		ch.msgSender.Send(msg)
		return
	}

	if len(clients) == 0 {
		errorText := "📊 *Мониторинг торрента*\n\nКлиенты не найдены. Добавьте клиента для мониторинга."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := ch.stateMgr.GetMenuMessage(chatId)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if err == nil {
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
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", sslText, client.Name),
				fmt.Sprintf("monitor_torrent_client_%d", client.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err == nil {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	}
}

// StartTorrentMonitorDialog запускает диалог для ввода хеша торрента
func (ch *ClientHandler) StartTorrentMonitorDialog(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту для мониторинга: %v", err)
		// Продолжаем без списка торрентов
	} else {
		// Получаем список торрентов
		torrents, err := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
		if err == nil && len(torrents) > 0 {
			// Сортируем по времени добавления (более новые первыми)
			// Создаем копию для сортировки
			sortedTorrents := make([]qbittorrent.Torrent, len(torrents))
			copy(sortedTorrents, torrents)

			// Сортируем по AddedOn (более новые первыми)
			for i := 0; i < len(sortedTorrents)-1; i++ {
				for j := i + 1; j < len(sortedTorrents); j++ {
					if sortedTorrents[i].AddedOn < sortedTorrents[j].AddedOn {
						sortedTorrents[i], sortedTorrents[j] = sortedTorrents[j], sortedTorrents[i]
					}
				}
			}

			// Берем три последних
			maxTorrents := 3
			if len(sortedTorrents) < maxTorrents {
				maxTorrents = len(sortedTorrents)
			}

			// Сохраняем торренты в кэш
			var monitorItems []TorrentMonitorItem
			for i := 0; i < maxTorrents; i++ {
				torrent := sortedTorrents[i]
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

				ch.stateMgr.SetUserState(chatId, fmt.Sprintf("monitor_torrent_hash_%d", clientID))
				text := "📊 *Мониторинг торрента*\n\nВыберите торрент или введите хеш вручную:"
				var rows [][]tgbotapi.InlineKeyboardButton

				// Добавляем кнопки с последними торрентами (используем индекс вместо hash)
				for i, item := range monitorItems {
					// Обрезаем название если слишком длинное
					name := item.Name
					if len(name) > 40 {
						name = name[:37] + "..."
					}
					rows = append(rows, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(
							fmt.Sprintf("📁 %s", name),
							fmt.Sprintf("monitor_torrent_hash_btn_%d_%d", clientID, i),
						),
					))
				}

				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✏️ Ввести хеш вручную", fmt.Sprintf("monitor_torrent_manual_%d", clientID)),
				))
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "main_menu"),
				))

				keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
				messageID := ch.stateMgr.GetMenuMessage(chatId)
				newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
				if err != nil {
					logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
					return
				}
				ch.stateMgr.SetMenuMessage(chatId, newMessageID)
				return
			}
		}
	}

	// Если не удалось получить торренты, показываем обычный диалог ввода
	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("monitor_torrent_hash_%d", clientID))
	text := "📊 *Мониторинг торрента*\n\nВведите хеш торрента для мониторинга:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "main_menu"),
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

// StartTorrentMonitorManualInput запускает диалог для ручного ввода хеша
func (ch *ClientHandler) StartTorrentMonitorManualInput(chatId int64, clientID int64) {
	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("monitor_torrent_hash_%d", clientID))
	text := "📊 *Мониторинг торрента*\n\nВведите хеш торрента для мониторинга:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "main_menu"),
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
	} else {
		minutes := int(d.Minutes())
		seconds := d.Seconds() - float64(minutes*60)
		return fmt.Sprintf("%dm %.2fs", minutes, seconds)
	}
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
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("add_torrent_wait_file_%d", clientID))
	text := fmt.Sprintf("📥 *Добавление торрент файла*\n\nКлиент: *%s*\n\n📎 Отправьте торрент файл (.torrent):", client.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_add_torrent"),
		),
	)
	// Используем GetMenuMessage, чтобы редактировать то же сообщение, где был список клиентов
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) HandleTorrentFileReceived(ctx context.Context, chatId int64, clientID int64, fileData []byte, fileName string) {
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту %s для пользователя %d: %v", client.Name, chatId, err)
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("❌ Ошибка при подключении к клиенту *%s*", client.Name))
		ch.msgSender.Send(msg)
		return
	}

	// Парсим торрент файл для извлечения хеша
	torrentInfo, err := qBit.ParseTorrentFile(fileData)
	if err != nil {
		logger.Warn("Ошибка при парсинге торрент файла: %v", err)
		torrentInfo = nil
	}

	var existingPath string
	var existingHash string
	var torrentName string
	if torrentInfo != nil {
		torrentName = torrentInfo.Name
		// Проверяем, есть ли уже торрент с таким названием
		allTorrents, err := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
		if err == nil {
			existingTorrent := qBit.FindTorrentByName(allTorrents, torrentInfo.Name)
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

	savePaths, err := qBit.GetTorrentSavePaths(ctx, qbClient)
	if err != nil {
		logger.Warn("Ошибка при получении путей сохранения: %v", err)
		savePaths = []string{}
	}

	defaultPath, err := qBit.GetDefaultSavePath(ctx, qbClient)
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

	// Если найден существующий торрент с таким же хешем, предлагаем его путь первым
	if existingPath != "" {
		text += fmt.Sprintf("⭐ *Рекомендуется* (используется для этого торрента):\n`%s`\n\n", existingPath)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("⭐ Рекомендуется: %s", truncatePath(existingPath, 999)),
				fmt.Sprintf("select_save_path_%d_%d", clientID, -2),
			),
		))
	}

	if defaultPath != "" && defaultPath != existingPath {
		text += fmt.Sprintf("По умолчанию: `%s`\n\n", defaultPath)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
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
			// Пропускаем путь, если он уже показан как рекомендуемый или по умолчанию
			if path == existingPath || path == defaultPath {
				continue
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("📂 %s", truncatePath(path, 999)),
					fmt.Sprintf("select_save_path_%d_%d", clientID, i),
				),
			))
			shownCount++
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✏️ Ввести путь вручную", fmt.Sprintf("custom_save_path_%d", clientID)),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_add_torrent"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	// Используем GetMenuMessage, чтобы редактировать то же сообщение
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
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка: данные торрента не найдены. Начните заново.")
		ch.msgSender.Send(msg)
		return
	}

	var savePath string
	if pathIndex == -2 {
		// Используем путь из существующего торрента
		savePath = cache.ExistingPath
		if savePath == "" {
			logger.Warn("Путь из существующего торрента не найден")
			msg := tgbotapi.NewMessage(chatId, "❌ Ошибка: путь не найден")
			ch.msgSender.Send(msg)
			return
		}
	} else if pathIndex == -1 {
		savePath = cache.DefaultPath
	} else if pathIndex >= 0 && pathIndex < len(cache.SavePaths) {
		savePath = cache.SavePaths[pathIndex]
	} else {
		logger.Warn("Неверный индекс пути: %d", pathIndex)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка: неверный путь")
		ch.msgSender.Send(msg)
		return
	}

	// Сохраняем выбранный путь в кэш
	cache.SelectedPath = savePath
	ch.torrentFilesCache[chatId] = cache

	// Показываем вопрос о пропуске проверки хеша
	ch.ShowSkipHashCheckQuestion(chatId, clientID, savePath)
}

func (ch *ClientHandler) StartCustomSavePathDialog(chatId int64, clientID int64) {
	ch.stateMgr.SetUserState(chatId, fmt.Sprintf("add_torrent_custom_path_%d", clientID))
	text := "✏️ *Ввод пути сохранения*\n\nВведите путь для сохранения торрента:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_add_torrent"),
		),
	)
	// Используем GetMenuMessage, чтобы редактировать то же сообщение
	messageID := ch.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

// ShowSkipHashCheckQuestion показывает вопрос о пропуске проверки хеша
func (ch *ClientHandler) ShowSkipHashCheckQuestion(chatId int64, clientID int64, savePath string) {
	text := fmt.Sprintf("⚙️ *Настройки добавления торрента*\n\nПуть сохранения: `%s`\n\n❓ Пропустить проверку хеша при добавлении?\n\n_Проверка хеша может занять время, но гарантирует целостность данных._", savePath)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, пропустить", fmt.Sprintf("skip_hash_check_%d_true", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет, проверить", fmt.Sprintf("skip_hash_check_%d_false", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_add_torrent"),
		),
	)
	// Используем GetMenuMessage, чтобы редактировать то же сообщение
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
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту %s для пользователя %d: %v", client.Name, chatId, err)
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("❌ Ошибка при подключении к клиенту *%s*", client.Name))
		ch.msgSender.Send(msg)
		return
	}

	err = qBit.AddTorrentFile(ctx, qbClient, fileData, savePath, "", skipHashCheck)
	if err != nil {
		logger.Error("Ошибка при добавлении торрента: %v", err)
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("❌ Ошибка при добавлении торрента: %v", err))
		ch.msgSender.Send(msg)
		return
	}

	// Проверяем, был ли найден существующий торрент
	cache, exists := ch.torrentFilesCache[chatId]
	if exists && cache != nil && cache.ExistingHash != "" {
		// Предлагаем удалить старый торрент
		ch.ShowDeleteExistingTorrentQuestion(chatId, clientID, cache.ExistingHash, cache.TorrentName)
		return
	}

	// Получаем hash нового торрента из файла (fileData доступен как параметр)
	var newTorrentHash string
	torrentInfo, err := qBit.ParseTorrentFile(fileData)
	if err == nil && torrentInfo != nil {
		newTorrentHash = torrentInfo.InfoHash
		logger.Debug("Извлечен hash нового торрента: %s", newTorrentHash)
	} else {
		logger.Warn("Не удалось извлечь hash из торрент файла: %v", err)
	}

	ch.stateMgr.DeleteUserState(chatId)
	delete(ch.torrentFilesCache, chatId)

	// Если удалось получить hash, начинаем мониторинг
	if newTorrentHash != "" {
		logger.Debug("Запуск мониторинга торрента для пользователя %d, hash: %s", chatId, newTorrentHash)
		ch.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, newTorrentHash)
	} else {
		logger.Warn("Hash не получен, переход в главное меню для пользователя %d", chatId)
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowMainMenu(chatId)
		}
	}
}

// HandleSkipHashCheckSelection обрабатывает выбор опции пропуска проверки хеша
func (ch *ClientHandler) HandleSkipHashCheckSelection(chatId int64, clientID int64, skipHashCheck bool) {
	ctx := context.Background()
	cache, exists := ch.torrentFilesCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID {
		logger.Warn("Кэш торрент файла не найден для пользователя %d", chatId)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка: данные торрента не найдены. Начните заново.")
		ch.msgSender.Send(msg)
		return
	}

	if cache.SelectedPath == "" {
		logger.Warn("Выбранный путь не найден в кэше для пользователя %d", chatId)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка: путь сохранения не выбран. Начните заново.")
		ch.msgSender.Send(msg)
		return
	}

	ch.AddTorrentToClient(ctx, chatId, clientID, cache.FileData, cache.SelectedPath, skipHashCheck)
}

// ShowDeleteExistingTorrentQuestion показывает вопрос об удалении существующего торрента
func (ch *ClientHandler) ShowDeleteExistingTorrentQuestion(chatId int64, clientID int64, existingHash string, torrentName string) {
	// Hash уже сохранен в кэше, используем только clientID в callback
	text := fmt.Sprintf("✅ *Торрент успешно добавлен*\n\n⚠️ Найден существующий торрент с таким же названием:\n`%s`\n\n❓ Удалить старый торрент?", torrentName)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", fmt.Sprintf("delete_existing_torrent_%d", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет, оставить", fmt.Sprintf("keep_existing_torrent_%d", clientID)),
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
	// Hash уже сохранен в кэше, используем только clientID в callback
	text := "🗑️ *Удаление торрента*\n\n❓ Удалить файлы вместе с торрентом?\n\n_Если выбрать \"Да\", файлы будут удалены с диска._"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить файлы", fmt.Sprintf("confirm_delete_torrent_%d_true", clientID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет, только торрент", fmt.Sprintf("confirm_delete_torrent_%d_false", clientID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", fmt.Sprintf("keep_existing_torrent_%d", clientID)),
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

// HandleDeleteExistingTorrent обрабатывает удаление существующего торрента
func (ch *ClientHandler) HandleDeleteExistingTorrent(chatId int64, clientID int64, hash string, deleteFiles bool) {
	ctx := context.Background()
	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при получении данных клиента")
		ch.msgSender.Send(msg)
		return
	}

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту %s для пользователя %d: %v", client.Name, chatId, err)
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("❌ Ошибка при подключении к клиенту *%s*", client.Name))
		ch.msgSender.Send(msg)
		return
	}

	err = qBit.DeleteTorrent(ctx, qbClient, hash, deleteFiles)
	if err != nil {
		logger.Error("Ошибка при удалении торрента: %v", err)
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("❌ Ошибка при удалении торрента: %v", err))
		ch.msgSender.Send(msg)
		return
	}

	ch.stateMgr.DeleteUserState(chatId)

	// Получаем hash нового торрента из кэша и начинаем мониторинг
	cache, exists := ch.torrentFilesCache[chatId]
	var newTorrentHash string
	if exists && cache != nil {
		torrentInfo, err := qBit.ParseTorrentFile(cache.FileData)
		if err == nil && torrentInfo != nil {
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

	// Если удалось получить hash нового торрента, начинаем мониторинг
	if newTorrentHash != "" {
		ch.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, newTorrentHash)
	} else {
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowMainMenu(chatId)
		}
	}
}

// HandleKeepExistingTorrent обрабатывает решение оставить существующий торрент
func (ch *ClientHandler) HandleKeepExistingTorrent(chatId int64, clientID int64) {
	ctx := context.Background()

	// Получаем hash нового торрента из кэша перед его удалением
	cache, exists := ch.torrentFilesCache[chatId]
	var newTorrentHash string
	if exists && cache != nil {
		torrentInfo, err := qBit.ParseTorrentFile(cache.FileData)
		if err == nil && torrentInfo != nil {
			newTorrentHash = torrentInfo.InfoHash
			logger.Debug("Извлечен hash нового торрента из кэша: %s", newTorrentHash)
		} else {
			logger.Warn("Не удалось извлечь hash из торрент файла в кэше: %v", err)
		}
	}

	ch.stateMgr.DeleteUserState(chatId)
	delete(ch.torrentFilesCache, chatId)

	// Если удалось получить hash, начинаем мониторинг
	if newTorrentHash != "" {
		logger.Debug("Запуск мониторинга торрента для пользователя %d, hash: %s", chatId, newTorrentHash)
		ch.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, newTorrentHash)
	} else {
		logger.Warn("Hash не получен, переход в главное меню для пользователя %d", chatId)
		if ch.cmdHdlr != nil {
			ch.cmdHdlr.ShowMainMenu(chatId)
		}
	}
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	return "..." + path[len(path)-maxLen+3:]
}
