package qbit

import (
	"context"

	"github.com/autobrr/go-qbittorrent"
)

func (s *service) GetTorrentPropertiesCached(ctx context.Context, hash string) (*qbittorrent.TorrentProperties, error) {
	cacheMutex.RLock()
	cached, exists := torrentPropertiesCache[hash]
	cacheMutex.RUnlock()

	if exists && cached != nil {
		return cached, nil
	}

	props, err := s.client.GetTorrentPropertiesCtx(ctx, hash)
	if err != nil {
		return nil, err
	}

	cacheMutex.Lock()
	torrentPropertiesCache[hash] = &props
	cacheMutex.Unlock()

	return &props, nil
}
