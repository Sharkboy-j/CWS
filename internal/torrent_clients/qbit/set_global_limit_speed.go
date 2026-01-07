package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

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
