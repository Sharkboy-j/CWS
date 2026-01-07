package quick_actions

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
)

func (h *Handler) HandlePauseAllTorrents(chatId int64) {
	logger.Debugf("Handling pause all torrents for user %d", chatId)
	ctx, clients, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}
	text := ui.Msg(ui.MsgPauseAllHeaderText)
	var successCount, failCount int
	var failedClients []string

	for _, client := range clients {
		qbClient, err := qbit.New(ctx, client)
		if err != nil {
			logger.Error("Error connecting to qBit client %s for user %d: %v", client.Name, chatId, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		err = qbClient.PauseAllTorrents(ctx)
		if err != nil {
			logger.Error("Error pausing all torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
			text += ui.Msgs(ui.MsgPauseAllClientSuccessFmt, client.Name)
		}
	}

	keyboard := h.pauseTorrentsKeyboard()
	_ = h.sendOrEditResult(chatId, messageID, text, successCount, failCount, failedClients, &keyboard)
}

func (h *Handler) HandleResumeAllTorrents(chatId int64) {
	logger.Debugf("Handling resume all torrents for user %d", chatId)
	ctx, clients, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}
	text := ui.Msg(ui.MsgResumeAllHeaderText)
	var successCount, failCount int
	var failedClients []string

	for _, client := range clients {
		qbClient, err := qbit.New(ctx, client)
		if err != nil {
			logger.Error("Error connecting to qBit client %s for user %d: %v", client.Name, chatId, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		err = qbClient.ResumeAllTorrents(ctx)
		if err != nil {
			logger.Error("Error resuming all torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
			text += ui.Msgs(ui.MsgResumeAllClientSuccessFmt, client.Name)
		}
	}

	keyboard := h.resumeTorrentsKeyboard()
	_ = h.sendOrEditResult(chatId, messageID, text, successCount, failCount, failedClients, &keyboard)
}

func (h *Handler) HandlePauseNonRutrackerTorrents(chatId int64) {
	h.handleTorrentSelectionAction(
		chatId,
		"pause non-rutracker torrents",
		ui.Msg(ui.MsgPauseNonRutrackerHeaderText),
		true,
		torrentSelectionNonRutracker,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.PauseTorrents(ctx, hashes)
		},
		ui.MsgPauseNonRutrackerClientSuccessFmt,
	)
}

func (h *Handler) HandleResumeNonRutrackerTorrents(chatId int64) {
	h.handleTorrentSelectionAction(
		chatId,
		"resume non-rutracker torrents",
		ui.Msg(ui.MsgResumeNonRutrackerHeaderText),
		false,
		torrentSelectionNonRutracker,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.ResumeTorrents(ctx, hashes)
		},
		ui.MsgResumeNonRutrackerClientSuccessFmt,
	)
}

func (h *Handler) HandlePauseRutrackerTorrents(chatId int64) {
	h.handleTorrentSelectionAction(
		chatId,
		"pause rutracker torrents",
		ui.Msg(ui.MsgPauseRutrackerHeaderText),
		true,
		torrentSelectionRutracker,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.PauseTorrents(ctx, hashes)
		},
		ui.MsgPauseRutrackerClientSuccessFmt,
	)
}

func (h *Handler) HandleResumeRutrackerTorrents(chatId int64) {
	h.handleTorrentSelectionAction(
		chatId,
		"resume rutracker torrents",
		ui.Msg(ui.MsgResumeRutrackerHeaderText),
		false,
		torrentSelectionRutracker,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.ResumeTorrents(ctx, hashes)
		},
		ui.MsgResumeRutrackerClientSuccessFmt,
	)
}

func (h *Handler) HandleLimitSpeedBytes(chatId int64, limitBytesPerSec int64) {
	logger.Debugf("Handling limit speed for user %d: %d bytes/s", chatId, limitBytesPerSec)
	ctx, clients, _, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}

	limitMBPerSec := float64(limitBytesPerSec) / (1024 * 1024)

	var successCount, failCount int
	var failedClients []string

	for _, client := range clients {
		qbClient, err := qbit.New(ctx, client)
		if err != nil {
			logger.Error("Error connecting to qBit client %s for user %d: %v", client.Name, chatId, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		err = qbClient.SetGlobalSpeedLimits(ctx, limitBytesPerSec, limitBytesPerSec)
		if err != nil {
			logger.Error("Error setting speed limits for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
		}
	}

	if failCount > 0 {
		messageID := h.stateMgr.GetMenuMessage(chatId)
		text := ui.Msgs(ui.MsgSpeedLimitAppliedHeaderFmt, limitMBPerSec)
		_ = h.sendOrEditResultWithMainMenu(chatId, messageID, text, successCount, failCount, failedClients)

		return
	}

	if h.cmdHdlr != nil {
		h.cmdHdlr.ShowMainMenu(chatId)
	}
}

func (h *Handler) HandleRemoveSpeedLimits(chatId int64) {
	logger.Debugf("Handling remove speed limits for user %d", chatId)
	ctx, clients, _, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}

	limitBytesPerSec := int64(0)

	var successCount, failCount int
	var failedClients []string

	for _, client := range clients {
		qbClient, err := qbit.New(ctx, client)
		if err != nil {
			logger.Error("Error connecting to qBit client %s for user %d: %v", client.Name, chatId, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		err = qbClient.SetGlobalSpeedLimits(ctx, limitBytesPerSec, limitBytesPerSec)
		if err != nil {
			logger.Error("Error removing speed limits for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
		}
	}

	if failCount > 0 {
		messageID := h.stateMgr.GetMenuMessage(chatId)
		text := ui.Msg(ui.MsgSpeedLimitsRemovedHeader)
		_ = h.sendOrEditResultWithMainMenu(chatId, messageID, text, successCount, failCount, failedClients)

		return
	}

	if h.cmdHdlr != nil {
		h.cmdHdlr.ShowMainMenu(chatId)
	}
}
