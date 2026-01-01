package qBit

import (
	"context"
	"strings"
	"sync"

	"github.com/autobrr/go-qbittorrent"
)

var (
	torrentPropertiesCache = make(map[string]*qbittorrent.TorrentProperties)
	cacheMutex             sync.RWMutex
)

func FilterTorrentsByRutrackerComment(ctx context.Context, client *qbittorrent.Client, torrents []qbittorrent.Torrent) ([]qbittorrent.Torrent, error) {
	if len(torrents) == 0 {
		return []qbittorrent.Torrent{}, nil
	}

	const (
		rutrackerKeyword = "rutracker"
		batchSize        = 5
	)

	var (
		filteredTorrents []qbittorrent.Torrent
		mu               sync.Mutex // Для безопасного доступа к filteredTorrents
	)

	totalBatches := (len(torrents) + batchSize - 1) / batchSize

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

				props, err := GetTorrentPropertiesCached(ctx, client, torrent.Hash)
				if err != nil {
					// Если не удалось получить свойства, пропускаем торрент
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

	return filteredTorrents, nil
}

func GetTorrentPropertiesCached(ctx context.Context, client *qbittorrent.Client, hash string) (*qbittorrent.TorrentProperties, error) {
	cacheMutex.RLock()
	cached, exists := torrentPropertiesCache[hash]
	cacheMutex.RUnlock()

	if exists && cached != nil {
		return cached, nil
	}

	props, err := client.GetTorrentPropertiesCtx(ctx, hash)
	if err != nil {
		return nil, err
	}

	cacheMutex.Lock()
	torrentPropertiesCache[hash] = &props
	cacheMutex.Unlock()

	return &props, nil
}
