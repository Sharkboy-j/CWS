package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"net/http"
	"net/url"
)

func (s *service) ResumeTorrent(ctx context.Context, hash string) error {
	logger.Debugf("Resuming torrent %s", hash)
	data := url.Values{}
	data.Set("hashes", hash)

	body, status, err := s.doForm(ctx, "/api/v2/torrents/start", data)
	if err != nil {
		logger.Error("Error resuming torrent: %v", err)

		return fmt.Errorf("failed to resume torrent: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error resuming torrent %s: status %d, response: %s", hash, status, string(body))

		return fmt.Errorf("resume failed with status %d: %s", status, string(body))
	}

	logger.Infof("Torrent %s resumed successfully", hash)

	return nil
}
