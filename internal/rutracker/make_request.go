package rutracker

import (
	"cws/config"
	"cws/logger"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func makeRequest(cfg *config.Config, urlPart string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", cfg.RutrackerHost, urlPart)
	safeURL := maskSensitiveURL(url)

	logger.Debug("RuTracker API request: %s", safeURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Error("RuTracker API failed to create request: url=%s err=%v", safeURL, err)

		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("RuTracker API request failed: url=%s err=%v", safeURL, err)

		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("RuTracker API failed to read response: url=%s status=%d err=%v", safeURL, resp.StatusCode, err)

		return nil, err
	}

	duration := time.Since(start)

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"RuTracker API bad status: url=%s status=%d bytes=%d duration=%v body=%s",
			safeURL, resp.StatusCode, len(body), duration, truncateBody(body, 500),
		)

		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, truncateBody(body, 200))
	}

	logger.Debug(
		"RuTracker API response: url=%s status=%d bytes=%d duration=%v",
		safeURL, resp.StatusCode, len(body), duration,
	)

	if len(body) == 0 {
		logger.Warn("RuTracker API returned empty body: url=%s", safeURL)
	}

	return body, nil
}

func maskSensitiveURL(rawURL string) string {
	keyPrefix := "api_key="
	idx := strings.Index(rawURL, keyPrefix)
	if idx == -1 {
		return rawURL
	}

	valueStart := idx + len(keyPrefix)
	valueEnd := strings.Index(rawURL[valueStart:], "&")
	if valueEnd == -1 {
		return rawURL[:valueStart] + "***"
	}

	return rawURL[:valueStart] + "***" + rawURL[valueStart+valueEnd:]
}

func truncateBody(body []byte, maxLen int) string {
	text := strings.TrimSpace(string(body))
	if len(text) <= maxLen {
		return text
	}

	return text[:maxLen] + "..."
}
