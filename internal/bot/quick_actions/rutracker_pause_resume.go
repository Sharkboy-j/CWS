package quick_actions

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"

	"github.com/autobrr/go-qbittorrent"
)

func (h *Handler) HandlePauseRutrackerTorrents(chatId int64) {
	h.handleRutrackerTorrentsAction(
		chatId,
		"pause rutracker torrents",
		ui.Msg(ui.MsgPauseRutrackerHeaderText),
		true,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.PauseTorrents(ctx, hashes)
		},
		ui.MsgPauseRutrackerClientSuccessFmt,
	)
}

func (h *Handler) HandleResumeRutrackerTorrents(chatId int64) {
	h.handleRutrackerTorrentsAction(
		chatId,
		"resume rutracker torrents",
		ui.Msg(ui.MsgResumeRutrackerHeaderText),
		false,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.ResumeTorrents(ctx, hashes)
		},
		ui.MsgResumeRutrackerClientSuccessFmt,
	)
}

func (h *Handler) handleRutrackerTorrentsAction(
	chatId int64,
	actionLog string,
	headerText string,
	isPause bool,
	apply func(ctx context.Context, qbClient qbit.Service, hashes []string) error,
	successFmt ui.MsgID,
) {
	logger.Debugf("Handling %s for user %d", actionLog, chatId)
	ctx, clients, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}

	text := headerText
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

		torrents, err := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
			Filter: qbittorrent.TorrentFilter("all"),
		})
		if err != nil {
			logger.Error("Error getting torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		rutrackerTorrents, err := qbClient.FilterTorrentsByRutrackerComment(ctx, torrents)
		if err != nil {
			logger.Error("Error filtering rutracker torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		hashes := make([]string, 0, len(rutrackerTorrents))
		for _, t := range rutrackerTorrents {
			if t.Hash != "" {
				hashes = append(hashes, t.Hash)
			} else if t.InfohashV1 != "" {
				hashes = append(hashes, t.InfohashV1)
			}
		}

		applyErr := apply(ctx, qbClient, hashes)
		if applyErr != nil {
			logger.Error("Error during %s for client %s: %v", actionLog, client.Name, applyErr)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		successCount++
		text += ui.Msgf(successFmt, client.Name, len(hashes))
	}

	if isPause {
		keyboard := h.pauseTorrentsKeyboard()
		_ = h.sendOrEditResult(chatId, messageID, text, successCount, failCount, failedClients, &keyboard)
	} else {
		keyboard := h.resumeTorrentsKeyboard()
		_ = h.sendOrEditResult(chatId, messageID, text, successCount, failCount, failedClients, &keyboard)
	}
}
