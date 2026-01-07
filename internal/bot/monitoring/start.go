package monitoring

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"cws/internal/bot/ui"
	"cws/internal/textutil"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
)

func (tms *torrentMonitorService) StartTorrentMonitoring(ctx context.Context, chatId int64, clientID int64, hash string) {
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
		text := ui.Msgs(ui.MsgTorrentMonitorClientProcessingFmt, textutil.EscapeMarkdown(client.Name))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := tms.getMenuMessage(chatId)
		newMessageID, _ := tms.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if newMessageID != 0 && newMessageID != messageID {
			tms.setMenuMessage(chatId, newMessageID)
			monitor.MessageID = newMessageID
		} else {
			monitor.MessageID = messageID
		}
	}

	go tms.monitorTorrentProgress(ctx, monitor)
}

func (tms *torrentMonitorService) monitorTorrentProgress(ctx context.Context, monitor *TorrentMonitor) {
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

func (tms *torrentMonitorService) updateTorrentProgress(ctx context.Context, monitor *TorrentMonitor) {
	client, err := tms.repo.GetClientByID(ctx, monitor.ClientID, monitor.ChatID)
	if err != nil || client == nil {
		logger.Error("Ошибка при получении клиента %d для мониторинга: %v", monitor.ClientID, err)
		tms.StopTorrentMonitoring(monitor.ChatID)

		return
	}

	qbClient, err := qbit.New(ctx, client)
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
		text := ui.Msgs(ui.MsgTorrentMonitorClientTorrentProcessingFmt, textutil.EscapeMarkdown(client.Name))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := tms.getMenuMessage(monitor.ChatID)
		newMessageID, _ := tms.msgSender.SendOrEdit(monitor.ChatID, messageID, text, &keyboard)
		if newMessageID != 0 && newMessageID != messageID {
			tms.setMenuMessage(monitor.ChatID, newMessageID)
		}

		return
	}

	numPeers := int(torrent.NumSeeds) + int(torrent.NumIncomplete)

	var torrentURL string
	props, err := qbClient.GetTorrentPropertiesCtx(ctx, monitor.Hash)
	if err == nil {
		torrentURL = textutil.ExtractURLFromComment(props.Comment)
	}

	text := tms.formatTorrentProgress(torrent, client.Name, numPeers)

	var rows [][]tgbotapi.InlineKeyboardButton
	if torrentURL != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(ui.Msg(ui.MsgTorrentMonitorOpenTorrentButtonText), torrentURL),
		))
	}
	stateStr := string(torrent.State)
	lowerState := strings.ToLower(stateStr)
	isActive := lowerState == "downloading" ||
		lowerState == "uploading" ||
		lowerState == "stalledup" ||
		lowerState == "stalleddl" ||
		lowerState == "metadl" ||
		strings.HasPrefix(lowerState, "forced")
	if isActive {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.PauseTorrent, fmt.Sprintf("monitor_pause_%d_%s", monitor.ClientID, monitor.Hash)),
			ui.ButtonWithData(ui.Delete, fmt.Sprintf("monitor_delete_%d_%s", monitor.ClientID, monitor.Hash)),
		))
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.ResumeTorrent, fmt.Sprintf("monitor_resume_%d_%s", monitor.ClientID, monitor.Hash)),
			ui.ButtonWithData(ui.Delete, fmt.Sprintf("monitor_delete_%d_%s", monitor.ClientID, monitor.Hash)),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.BackToTorrents),
		ui.Button(ui.MainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	currentMenuMsgID := tms.getMenuMessage(monitor.ChatID)
	if currentMenuMsgID != 0 && monitor.MessageID != 0 && currentMenuMsgID != monitor.MessageID {
		logger.Debug("User %d left monitoring menu (menu msg id changed from %d to %d), skipping update", monitor.ChatID, monitor.MessageID, currentMenuMsgID)

		return
	}

	messageID := tms.getMenuMessage(monitor.ChatID)
	newMessageID, err := tms.msgSender.SendOrEdit(monitor.ChatID, messageID, text, &keyboard)
	if err != nil {
		logger.Warn("Ошибка при обновлении прогресса торрента: %v", err)
	}
	if newMessageID != 0 && newMessageID != messageID {
		tms.setMenuMessage(monitor.ChatID, newMessageID)
		monitor.MessageID = newMessageID
	}
}

func (tms *torrentMonitorService) formatTorrentProgress(torrent *qbittorrent.Torrent, clientName string, numPeers int) string {
	var status string
	var progress float64
	switch torrent.State {
	case "downloading":
		status = ui.Msg(ui.MsgTorrentActivityDownloading)
		progress = torrent.Progress * 100
	case "uploading":
		status = ui.Msg(ui.MsgTorrentActivityUploading)
		progress = 100.0
	case "stalledUP":
		status = ui.Msg(ui.MsgTorrentActivityUploadStalled)
		progress = torrent.Progress * 100
	case "stalledDL":
		status = ui.Msg(ui.MsgTorrentActivityDownloadStalled)
		progress = torrent.Progress * 100
	case "checkingUP", "checkingDL", "checkingResumeData":
		status = ui.Msg(ui.MsgTorrentActivityChecking)
		progress = torrent.Progress * 100
	case "queuedUP", "queuedDL":
		status = ui.Msg(ui.MsgTorrentActivityQueued)
		progress = torrent.Progress * 100
	case "pausedUP", "pausedDL", "stoppedUP", "stoppedDL":
		status = ui.Msg(ui.MsgTorrentActivityPaused)
		progress = torrent.Progress * 100
	case "metaDL":
		status = ui.Msg(ui.MsgTorrentActivityFetchingMetadata)
		progress = torrent.Progress * 100
	case "error":
		status = ui.Msg(ui.MsgTorrentActivityError)
		progress = torrent.Progress * 100
	case "missingFiles":
		status = ui.Msg(ui.MsgTorrentActivityMissingFiles)
		progress = torrent.Progress * 100
	default:
		status = ui.Msgs(ui.MsgTorrentActivityOtherFmt, string(torrent.State))
		progress = torrent.Progress * 100
	}

	text := ui.Msg(ui.MsgTorrentProgressHeaderText)
	text += ui.Msgs(ui.MsgTorrentProgressNameFmt, textutil.EscapeMarkdown(torrent.Name))
	text += ui.Msgs(ui.MsgTorrentProgressPathFmt, torrent.SavePath)
	text += ui.Msgs(ui.MsgTorrentProgressStatusFmt, status)
	text += ui.Msgs(ui.MsgTorrentProgressPercentFmt, progress)
	text += ui.Msgs(ui.MsgTorrentProgressDownloadFmt, ui.FormatSpeedBytes(torrent.DlSpeed))
	text += ui.Msgs(ui.MsgTorrentProgressUploadFmt, ui.FormatSpeedBytes(torrent.UpSpeed))
	text += ui.Msgs(ui.MsgTorrentProgressUploadedFmt, ui.FormatBytes(torrent.Uploaded))
	text += ui.Msgs(ui.MsgTorrentProgressSeedsPeersFmt, torrent.NumSeeds, numPeers)
	text += ui.Msgs(ui.MsgTorrentProgressSizeFmt, ui.FormatBytes(torrent.Completed), ui.FormatBytes(torrent.Size))

	return text
}
