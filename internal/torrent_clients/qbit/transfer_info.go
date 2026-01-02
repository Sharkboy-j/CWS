package qbit

import (
	"context"
	"cws/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *service) GetTransferInfo(ctx context.Context) (*TransferInfo, error) {
	logger.Debug("Getting transfer info")

	apiURL := fmt.Sprintf("%s/api/v2/transfer/info", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		logger.Error("Error creating transfer info request: %v", err)

		return nil, fmt.Errorf("failed to create transfer info request: %w", err)
	}

	req.Header.Set("Referer", s.baseURL+"/")
	req.Header.Set("Origin", s.baseURL)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("Error getting transfer info: %v", err)

		return nil, fmt.Errorf("failed to get transfer info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading transfer info response: %v", err)

		return nil, fmt.Errorf("failed to read transfer info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Error getting transfer info: status %d, response: %s", resp.StatusCode, string(body))

		return nil, fmt.Errorf("get transfer info failed with status %d: %s", resp.StatusCode, string(body))
	}

	var transferInfo struct {
		DlInfoSpeed int64 `json:"dl_info_speed"`
		UpInfoSpeed int64 `json:"up_info_speed"`
		DlRateLimit int64 `json:"dl_rate_limit"`
		UpRateLimit int64 `json:"up_rate_limit"`
	}

	if err := json.Unmarshal(body, &transferInfo); err != nil {
		logger.Error("Error parsing transfer info response: %v", err)

		return nil, fmt.Errorf("failed to parse transfer info response: %w", err)
	}

	return &TransferInfo{
		DownloadSpeed: transferInfo.DlInfoSpeed,
		UploadSpeed:   transferInfo.UpInfoSpeed,
		DownloadLimit: transferInfo.DlRateLimit,
		UploadLimit:   transferInfo.UpRateLimit,
	}, nil
}
