package telegram

import (
	"context"
	"cws/database"
	"cws/logger"
	"cws/qBit"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TorrentMonitor struct {
	ChatID   int64
	ClientID int64
	Hash     string
	Stop     chan bool
}

type TorrentMonitorService struct {
	repo              *database.Repository
	msgSender         *MessageSender
	stateMgr          *StateManager
	torrentMonitoring map[int64]*TorrentMonitor
}

func NewTorrentMonitorService(repo *database.Repository, msgSender *MessageSender, stateMgr *StateManager) *TorrentMonitorService {
	return &TorrentMonitorService{
		repo:              repo,
		msgSender:         msgSender,
		stateMgr:          stateMgr,
		torrentMonitoring: make(map[int64]*TorrentMonitor),
	}
}

// StartTorrentMonitoring запускает мониторинг прогресса торрента
func (tms *TorrentMonitorService) StartTorrentMonitoring(ctx context.Context, chatId int64, clientID int64, hash string) {
	if monitor, exists := tms.torrentMonitoring[chatId]; exists {
		close(monitor.Stop)
		delete(tms.torrentMonitoring, chatId)
	}

	monitor := &TorrentMonitor{
		ChatID:   chatId,
		ClientID: clientID,
		Hash:     hash,
		Stop:     make(chan bool),
	}
	tms.torrentMonitoring[chatId] = monitor

	client, err := tms.repo.GetClientByID(ctx, clientID, chatId)
	if err == nil && client != nil {
		text := fmt.Sprintf("✅ *\n\nКлиент: *%s*\n\n⏳ Обработка...", client.Name)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := tms.stateMgr.GetMenuMessage(chatId)
		tms.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	}

	go tms.monitorTorrentProgress(ctx, monitor)
}

// StopTorrentMonitoring останавливает мониторинг торрента
func (tms *TorrentMonitorService) StopTorrentMonitoring(chatId int64) {
	if monitor, exists := tms.torrentMonitoring[chatId]; exists {
		close(monitor.Stop)
		delete(tms.torrentMonitoring, chatId)
	}
}

// monitorTorrentProgress периодически обновляет информацию о прогрессе торрента
func (tms *TorrentMonitorService) monitorTorrentProgress(ctx context.Context, monitor *TorrentMonitor) {
	time.Sleep(1 * time.Second)

	tms.updateTorrentProgress(ctx, monitor)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-monitor.Stop:
			logger.Debug("Мониторинг торрента остановлен для пользователя %d", monitor.ChatID)
			return
		case <-ctx.Done():
			logger.Debug("Контекст отменен, остановка мониторинга для пользователя %d", monitor.ChatID)
			return
		case <-ticker.C:
			tms.updateTorrentProgress(ctx, monitor)
		}
	}
}

// updateTorrentProgress обновляет сообщение с информацией о прогрессе торрента
func (tms *TorrentMonitorService) updateTorrentProgress(ctx context.Context, monitor *TorrentMonitor) {
	client, err := tms.repo.GetClientByID(ctx, monitor.ClientID, monitor.ChatID)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для мониторинга: %v", monitor.ClientID, err)
		tms.StopTorrentMonitoring(monitor.ChatID)
		return
	}

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту для мониторинга: %v", err)
		return
	}

	torrents, err := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
	if err != nil {
		logger.Warn("Ошибка при получении торрентов для мониторинга: %v", err)
		return
	}

	var torrent *qbittorrent.Torrent
	hashUpper := strings.ToUpper(monitor.Hash)
	for i := range torrents {
		if strings.ToUpper(torrents[i].InfohashV1) == hashUpper || strings.ToUpper(torrents[i].InfohashV2) == hashUpper {
			torrent = &torrents[i]
			break
		}
	}

	if torrent == nil {
		logger.Debug("Торрент не найден для мониторинга, hash: %s, продолжаем попытки...", monitor.Hash)
		text := fmt.Sprintf("✅ \n\nКлиент: *%s*\n\n⏳ Торрент обрабатывается qBittorrent...\n\n_Обработка..._", client.Name)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := tms.stateMgr.GetMenuMessage(monitor.ChatID)
		tms.msgSender.SendOrEdit(monitor.ChatID, messageID, text, &keyboard)
		return
	}

	numPeers := int(torrent.NumSeeds) + int(torrent.NumIncomplete)

	var torrentURL string
	props, err := qbClient.GetTorrentPropertiesCtx(ctx, monitor.Hash)
	if err == nil {
		torrentURL = tms.extractURLFromComment(props.Comment)
	}

	text := tms.formatTorrentProgress(torrent, client.Name, numPeers)

	var rows [][]tgbotapi.InlineKeyboardButton
	if torrentURL != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 Открыть раздачу", torrentURL),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	messageID := tms.stateMgr.GetMenuMessage(monitor.ChatID)
	_, err = tms.msgSender.SendOrEdit(monitor.ChatID, messageID, text, &keyboard)
	if err != nil {
		logger.Warn("Ошибка при обновлении прогресса торрента: %v", err)
	}
}

// formatTorrentProgress форматирует информацию о прогрессе торрента
func (tms *TorrentMonitorService) formatTorrentProgress(torrent *qbittorrent.Torrent, clientName string, numPeers int) string {
	var status string
	var progress float64

	switch torrent.State {
	case "downloading":
		status = "⬇️ Загрузка"
		progress = torrent.Progress * 100
	case "uploading", "stalledUP":
		status = "⬆️ Раздача"
		progress = 100.0
	case "checkingUP", "checkingDL", "checkingResumeData":
		status = "🔍 Проверка"
		progress = torrent.Progress * 100
	case "queuedUP", "queuedDL":
		status = "⏳ В очереди"
		progress = torrent.Progress * 100
	case "pausedUP", "pausedDL":
		status = "⏸ Остановлен"
		progress = torrent.Progress * 100
	case "error":
		status = "❌ Ошибка"
		progress = torrent.Progress * 100
	case "missingFiles":
		status = "⚠️ Отсутствуют файлы"
		progress = torrent.Progress * 100
	default:
		status = "ℹ️ " + string(torrent.State)
		progress = torrent.Progress * 100
	}

	formatSize := func(bytes int64) string {
		const unit = 1024
		if bytes < unit {
			return fmt.Sprintf("%d B", bytes)
		}
		div, exp := int64(unit), 0
		for n := bytes / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
	}

	formatSpeed := func(bytesPerSec int64) string {
		if bytesPerSec < 1024 {
			return fmt.Sprintf("%d B/s", bytesPerSec)
		}
		return formatSize(bytesPerSec) + "/s"
	}

	text := fmt.Sprintf("📊 *Прогресс торрента*\n\n")
	text += fmt.Sprintf("Клиент: *%s*\n\n", clientName)
	text += fmt.Sprintf("📁 *%s*\n\n", torrent.Name)
	text += fmt.Sprintf("Статус: %s\n", status)
	text += fmt.Sprintf("Прогресс: *%.1f%%*\n\n", progress)
	text += fmt.Sprintf("⬇️ Загрузка: %s\n", formatSpeed(torrent.DlSpeed))
	text += fmt.Sprintf("⬆️ Отдача: %s\n", formatSpeed(torrent.UpSpeed))
	text += fmt.Sprintf("📤 Всего отдано: %s\n", formatSize(torrent.Uploaded))
	text += fmt.Sprintf("👥 Сиды: %d | Пиры: %d\n\n", torrent.NumSeeds, numPeers)
	text += fmt.Sprintf("📦 Размер: %s / %s", formatSize(torrent.Completed), formatSize(torrent.Size))

	return text
}

// extractURLFromComment извлекает URL из комментария торрента
func (tms *TorrentMonitorService) extractURLFromComment(comment string) string {
	if comment == "" {
		return ""
	}

	urlPattern := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches := urlPattern.FindString(comment)
	if matches != "" {
		return matches
	}

	rutrackerPattern := regexp.MustCompile(`(?:rutracker\.org|rutracker\.cc)/[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches = rutrackerPattern.FindString(comment)
	if matches != "" {
		if !strings.HasPrefix(matches, "http://") && !strings.HasPrefix(matches, "https://") {
			return "https://" + matches
		}
		return matches
	}

	return ""
}
