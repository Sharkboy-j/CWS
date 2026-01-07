package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"net/http"
	"net/url"
)

func (s *service) ResumeAllTorrents(ctx context.Context) error {
	logger.Debug("Resuming all torrents")
	data := url.Values{}
	data.Set("hashes", "all")

	body, status, err := s.doForm(ctx, "/api/v2/torrents/start", data)
	if err != nil {
		logger.Error("Error resuming all torrents: %v", err)

		return fmt.Errorf("failed to resume all torrents: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error resuming all torrents: status %d, response: %s", status, string(body))

		return fmt.Errorf("resume failed with status %d: %s", status, string(body))
	}

	logger.Info("All torrents resumed successfully")

	return nil
}
