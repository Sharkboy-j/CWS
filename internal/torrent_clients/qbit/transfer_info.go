package qbit

import (
	"context"
	"cws/logger"
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *service) GetTransferInfo(ctx context.Context) (*TransferInfo, error) {
	logger.Debug("Getting transfer info")

	body, status, err := s.doRequest(ctx, http.MethodGet, "/api/v2/transfer/info", nil, "")
	if err != nil {
		logger.Error("Error getting transfer info: %v", err)

		return nil, fmt.Errorf("failed to get transfer info: %w", err)
	}
	if status != http.StatusOK {
		logger.Error("Error getting transfer info: status %d, response: %s", status, string(body))

		return nil, fmt.Errorf("get transfer info failed with status %d: %s", status, string(body))
	}

	var transferInfo struct {
		DlInfoSpeed int64 `json:"dl_info_speed"`
		UpInfoSpeed int64 `json:"up_info_speed"`
		DlRateLimit int64 `json:"dl_rate_limit"`
		UpRateLimit int64 `json:"up_rate_limit"`
	}

	if err = json.Unmarshal(body, &transferInfo); err != nil {
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
