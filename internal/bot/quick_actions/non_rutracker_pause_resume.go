package quick_actions

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"

	"github.com/autobrr/go-qbittorrent"
)

func (h *Handler) HandlePauseNonRutrackerTorrents(chatId int64) {
	h.handleNonRutrackerTorrentsAction(
		chatId,
		"pause non-rutracker torrents",
		ui.Msg(ui.MsgPauseNonRutrackerHeaderText),
		true,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.PauseTorrents(ctx, hashes)
		},
		ui.MsgPauseNonRutrackerClientSuccessFmt,
	)
}

func (h *Handler) HandleResumeNonRutrackerTorrents(chatId int64) {
	h.handleNonRutrackerTorrentsAction(
		chatId,
		"resume non-rutracker torrents",
		ui.Msg(ui.MsgResumeNonRutrackerHeaderText),
		false,
		func(ctx context.Context, qbClient qbit.Service, hashes []string) error {
			return qbClient.ResumeTorrents(ctx, hashes)
		},
		ui.MsgResumeNonRutrackerClientSuccessFmt,
	)
}

func (h *Handler) handleNonRutrackerTorrentsAction(
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

		allTorrents, err := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
			Filter: qbittorrent.TorrentFilter("all"),
		})
		if err != nil {
			logger.Error("Error getting torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		rutrackerTorrents, err := qbClient.FilterTorrentsByRutrackerComment(ctx, allTorrents)
		if err != nil {
			logger.Error("Error filtering rutracker torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		rutrackerHashes := make(map[string]struct{}, len(rutrackerTorrents))
		for _, t := range rutrackerTorrents {
			if t.Hash != "" {
				rutrackerHashes[t.Hash] = struct{}{}
			} else if t.InfohashV1 != "" {
				rutrackerHashes[t.InfohashV1] = struct{}{}
			}
		}

		hashes := make([]string, 0, len(allTorrents))
		for _, t := range allTorrents {
			hash := t.Hash
			if hash == "" {
				hash = t.InfohashV1
			}
			if hash == "" {
				continue
			}

			if _, isRutracker := rutrackerHashes[hash]; isRutracker {
				continue
			}
			hashes = append(hashes, hash)
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
