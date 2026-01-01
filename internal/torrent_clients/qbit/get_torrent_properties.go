package qbit

import (
	"context"

	"github.com/autobrr/go-qbittorrent"
)

func (s *service) GetTorrentPropertiesCtx(ctx context.Context, hash string) (qbittorrent.TorrentProperties, error) {
	return s.client.GetTorrentPropertiesCtx(ctx, hash)
}
