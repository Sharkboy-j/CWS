package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"net/http"
	"net/url"
)

func (s *service) PauseAllTorrents(ctx context.Context) error {
	logger.Debug("Pausing all torrents")
	data := url.Values{}
	data.Set("hashes", "all")

	body, status, err := s.doForm(ctx, "/api/v2/torrents/stop", data)
	if err != nil {
		logger.Error("Error pausing all torrents: %v", err)

		return fmt.Errorf("failed to pause all torrents: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error pausing all torrents: status %d, response: %s", status, string(body))

		return fmt.Errorf("pause failed with status %d: %s", status, string(body))
	}

	logger.Info("All torrents paused successfully")

	return nil
}
