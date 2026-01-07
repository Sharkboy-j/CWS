package quick_actions

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type torrentSelection int

const (
	torrentSelectionRutracker torrentSelection = iota
	torrentSelectionNonRutracker
)

func (h *Handler) handleTorrentSelectionAction(
	chatId int64,
	actionLog string,
	headerText string,
	isPause bool,
	selection torrentSelection,
	apply func(ctx context.Context, qbClient qbit.Service, hashes []string) error,
	successFmt ui.MsgID,
) {
	logger.Debugf("Handling %s for user %d", actionLog, chatId)
	ctx, clients, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if ok {
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

			hashes := collectTorrentHashes(allTorrents, rutrackerTorrents, selection)

			applyErr := apply(ctx, qbClient, hashes)
			if applyErr != nil {
				logger.Error("Error during %s for client %s: %v", actionLog, client.Name, applyErr)
				failCount++
				failedClients = append(failedClients, client.Name)

				continue
			}

			successCount++
			text += ui.Msgs(successFmt, client.Name, len(hashes))
		}

		var keyboard *tgbotapi.InlineKeyboardMarkup
		if isPause {
			kb := h.pauseTorrentsKeyboard()
			keyboard = &kb
		} else {
			kb := h.resumeTorrentsKeyboard()
			keyboard = &kb
		}

		_ = h.sendOrEditResult(chatId, messageID, text, successCount, failCount, failedClients, keyboard)
	} else {
		return
	}
}

func collectTorrentHashes(
	allTorrents []qbittorrent.Torrent,
	rutrackerTorrents []qbittorrent.Torrent,
	selection torrentSelection,
) []string {
	switch selection {
	case torrentSelectionRutracker:
		return rutrackerHashesOnly(rutrackerTorrents)
	case torrentSelectionNonRutracker:
		return nonRutrackerHashes(allTorrents, rutrackerTorrents)
	default:
		return nil
	}
}

func rutrackerHashesOnly(rutrackerTorrents []qbittorrent.Torrent) []string {
	hashes := make([]string, 0, len(rutrackerTorrents))
	for _, t := range rutrackerTorrents {
		hash := torrentHash(t)
		if hash != "" {
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

func nonRutrackerHashes(
	allTorrents []qbittorrent.Torrent,
	rutrackerTorrents []qbittorrent.Torrent,
) []string {
	rutrackerHashes := make(map[string]struct{}, len(rutrackerTorrents))
	for _, t := range rutrackerTorrents {
		hash := torrentHash(t)
		if hash != "" {
			rutrackerHashes[hash] = struct{}{}
		}
	}

	hashes := make([]string, 0, len(allTorrents))
	for _, t := range allTorrents {
		hash := torrentHash(t)
		if hash == "" {
			continue
		}

		_, exists := rutrackerHashes[hash]
		if exists {
			continue
		}

		hashes = append(hashes, hash)
	}

	return hashes
}

func torrentHash(t qbittorrent.Torrent) string {
	if t.Hash != "" {
		return t.Hash
	}

	return t.InfohashV1
}
