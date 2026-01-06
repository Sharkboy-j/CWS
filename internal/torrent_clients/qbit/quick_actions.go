package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

func (s *service) SetGlobalSpeedLimits(ctx context.Context, downloadLimit, uploadLimit int64) error {
	logger.Debug("Setting global speed limits: download=%d bytes/s, upload=%d bytes/s", downloadLimit, uploadLimit)

	data := url.Values{}
	data.Set("limit", strconv.FormatInt(downloadLimit, 10))

	body, status, err := s.doForm(ctx, "/api/v2/transfer/setDownloadLimit", data)
	if err != nil {
		logger.Error("Error setting global download limit: %v", err)

		return fmt.Errorf("failed to set global download limit: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error setting global download limit: status %d, response: %s", status, string(body))

		return fmt.Errorf("set download limit failed with status %d: %s", status, string(body))
	}

	data = url.Values{}
	data.Set("limit", strconv.FormatInt(uploadLimit, 10))
	body, status, err = s.doForm(ctx, "/api/v2/transfer/setUploadLimit", data)
	if err != nil {
		logger.Error("Error setting global upload limit: %v", err)

		return fmt.Errorf("failed to set global upload limit: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error setting global upload limit: status %d, response: %s", status, string(body))

		return fmt.Errorf("set upload limit failed with status %d: %s", status, string(body))
	}

	logger.Info("Global speed limits set successfully: download=%d bytes/s, upload=%d bytes/s", downloadLimit, uploadLimit)

	return nil
}
