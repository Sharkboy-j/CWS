package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"net/http"
	"net/url"
)

func (s *service) PauseTorrent(ctx context.Context, hash string) error {
	logger.Debugf("Pausing torrent %s", hash)
	data := url.Values{}
	data.Set("hashes", hash)

	body, status, err := s.doForm(ctx, "/api/v2/torrents/stop", data)
	if err != nil {
		logger.Error("Error pausing torrent: %v", err)

		return fmt.Errorf("failed to pause torrent: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error pausing torrent %s: status %d, response: %s", hash, status, string(body))

		return fmt.Errorf("pause failed with status %d: %s", status, string(body))
	}

	logger.Infof("Torrent %s paused successfully", hash)

	return nil
}
