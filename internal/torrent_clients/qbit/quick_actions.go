package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (s *service) PauseAllTorrents(ctx context.Context) error {
	logger.Debug("Pausing all torrents")

	apiURL := fmt.Sprintf("%s/api/v2/torrents/stop", s.baseURL)
	data := url.Values{}
	data.Set("hashes", "all")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Error creating pause request: %v", err)

		return fmt.Errorf("failed to create pause request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.baseURL+"/")
	req.Header.Set("Origin", s.baseURL)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Error pausing all torrents: %v", err)

		return fmt.Errorf("failed to pause all torrents: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading pause response: %v", err)

		return fmt.Errorf("failed to read pause response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error pausing all torrents: status %d, response: %s", resp.StatusCode, string(body))

		return fmt.Errorf("pause failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("All torrents paused successfully")

	return nil
}

func (s *service) ResumeAllTorrents(ctx context.Context) error {
	logger.Debug("Resuming all torrents")

	apiURL := fmt.Sprintf("%s/api/v2/torrents/start", s.baseURL)
	data := url.Values{}
	data.Set("hashes", "all")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Error creating resume request: %v", err)

		return fmt.Errorf("failed to create resume request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.baseURL+"/")
	req.Header.Set("Origin", s.baseURL)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Error resuming all torrents: %v", err)

		return fmt.Errorf("failed to resume all torrents: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading resume response: %v", err)

		return fmt.Errorf("failed to read resume response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error resuming all torrents: status %d, response: %s", resp.StatusCode, string(body))

		return fmt.Errorf("resume failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("All torrents resumed successfully")

	return nil
}

func (s *service) SetGlobalSpeedLimits(ctx context.Context, downloadLimit, uploadLimit int64) error {
	logger.Debug("Setting global speed limits: download=%d bytes/s, upload=%d bytes/s", downloadLimit, uploadLimit)

	downloadURL := fmt.Sprintf("%s/api/v2/transfer/setDownloadLimit", s.baseURL)
	data := url.Values{}
	data.Set("limit", strconv.FormatInt(downloadLimit, 10))

	req, err := http.NewRequestWithContext(ctx, "POST", downloadURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Error creating download limit request: %v", err)

		return fmt.Errorf("failed to create download limit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.baseURL+"/")
	req.Header.Set("Origin", s.baseURL)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Error setting global download limit: %v", err)

		return fmt.Errorf("failed to set global download limit: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading download limit response: %v", err)

		return fmt.Errorf("failed to read download limit response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error setting global download limit: status %d, response: %s", resp.StatusCode, string(body))

		return fmt.Errorf("set download limit failed with status %d: %s", resp.StatusCode, string(body))
	}

	uploadURL := fmt.Sprintf("%s/api/v2/transfer/setUploadLimit", s.baseURL)
	data = url.Values{}
	data.Set("limit", strconv.FormatInt(uploadLimit, 10))

	req, err = http.NewRequestWithContext(ctx, "POST", uploadURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Error creating upload limit request: %v", err)

		return fmt.Errorf("failed to create upload limit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", s.baseURL+"/")
	req.Header.Set("Origin", s.baseURL)

	resp, err = s.httpClient.Do(req)
	if err != nil {
		logger.Error("Error setting global upload limit: %v", err)

		return fmt.Errorf("failed to set global upload limit: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading upload limit response: %v", err)

		return fmt.Errorf("failed to read upload limit response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error setting global upload limit: status %d, response: %s", resp.StatusCode, string(body))

		return fmt.Errorf("set upload limit failed with status %d: %s", resp.StatusCode, string(body))
	}

	logger.Info("Global speed limits set successfully: download=%d bytes/s, upload=%d bytes/s", downloadLimit, uploadLimit)

	return nil
}
