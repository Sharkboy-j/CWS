package qbit

import (
	"context"
	"cws/logger"
	"strings"
	"sync"

	"github.com/autobrr/go-qbittorrent"
)

var (
	torrentPropertiesCache = make(map[string]*qbittorrent.TorrentProperties)
	cacheMutex             sync.RWMutex
)

func (s *service) FilterTorrentsByRutrackerComment(ctx context.Context, torrents []qbittorrent.Torrent) ([]qbittorrent.Torrent, error) {
	if len(torrents) == 0 {
		return []qbittorrent.Torrent{}, nil
	}

	const (
		rutrackerKeyword = "rutracker"
		batchSize        = 5
	)

	var (
		filteredTorrents []qbittorrent.Torrent
		mu               sync.Mutex
		propertiesErrors int
	)

	totalBatches := (len(torrents) + batchSize - 1) / batchSize
	logger.Debug("RuTracker filter: checking %d torrents in %d batches", len(torrents), totalBatches)

	for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
		start := batchIndex * batchSize
		end := start + batchSize
		if end > len(torrents) {
			end = len(torrents)
		}

		batch := torrents[start:end]
		var wg sync.WaitGroup

		for i := range batch {
			wg.Add(1)
			go func(torrent qbittorrent.Torrent) {
				defer wg.Done()

				props, err := s.GetTorrentPropertiesCached(ctx, torrent.Hash)
				if err != nil {
					mu.Lock()
					propertiesErrors++
					mu.Unlock()
					logger.Warn("RuTracker filter: failed to get properties for torrent %s (%s): %v", torrent.Hash, torrent.Name, err)

					return
				}

				comment := strings.ToLower(props.Comment)
				if strings.Contains(comment, rutrackerKeyword) {
					mu.Lock()
					filteredTorrents = append(filteredTorrents, torrent)
					mu.Unlock()
				}
			}(batch[i])
		}

		wg.Wait()
	}

	logger.Info(
		"RuTracker filter: matched %d/%d torrents, properties errors=%d",
		len(filteredTorrents), len(torrents), propertiesErrors,
	)

	return filteredTorrents, nil
}
