package rutracker

import (
	"cws/config"
	"cws/logger"
	"encoding/json"
	"fmt"
	"strings"
)

type hashAndNum struct {
	Result map[string]*int `json:"result"`
}

type apiErrorResponse struct {
	Error   json.RawMessage `json:"error"`
	Message string          `json:"message"`
}

type apiErrorDetails struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func GetIdByHashes(hashes []string, cfg *config.Config) (map[string]*int, error) {
	if len(hashes) == 0 {
		logger.Debug("RuTracker GetIdByHashes: no hash batches to check")

		return map[string]*int{}, nil
	}

	totalHashes := countHashesInBatches(hashes)
	logger.Info(
		"RuTracker GetIdByHashes: start batches=%d hashes=%d host=%s",
		len(hashes), totalHashes, cfg.RutrackerHost,
	)

	fullHashSet := make(map[string]*int)

	for i, hashBatch := range hashes {
		if hashBatch == "" {
			logger.Warn("RuTracker GetIdByHashes: empty batch %d/%d skipped", i+1, len(hashes))

			continue
		}

		batchHashCount := strings.Count(hashBatch, ",") + 1
		logger.Debug("RuTracker GetIdByHashes: batch %d/%d hashes=%d", i+1, len(hashes), batchHashCount)

		urlPart := fmt.Sprintf("v1/get_topic_id?by=hash&val=%s&api_key=%s", hashBatch, cfg.RutrackerApiToken)
		resp, err := makeRequest(cfg, urlPart)
		if err != nil {
			logger.Error("RuTracker GetIdByHashes: batch %d/%d request failed: %v", i+1, len(hashes), err)

			return nil, err
		}

		hashSet, err := parse(resp)
		if err != nil {
			logger.Error(
				"RuTracker GetIdByHashes: batch %d/%d parse failed: %v body=%s",
				i+1, len(hashes), err, truncateBody(resp, 500),
			)

			return nil, err
		}

		found, missing := countTopicIDs(hashSet)
		logger.Info(
			"RuTracker GetIdByHashes: batch %d/%d entries=%d found=%d missing=%d",
			i+1, len(hashes), len(hashSet), found, missing,
		)

		if len(hashSet) == 0 {
			logger.Warn(
				"RuTracker GetIdByHashes: batch %d/%d returned empty result map, body=%s",
				i+1, len(hashes), truncateBody(resp, 500),
			)
		}

		logMissingHashesInBatch(hashBatch, hashSet, i+1, len(hashes))

		for k, num := range hashSet {
			fullHashSet[k] = num
		}
	}

	totalFound, totalMissing := countTopicIDs(fullHashSet)
	logger.Info(
		"RuTracker GetIdByHashes: done unique=%d found=%d missing=%d requested=%d",
		len(fullHashSet), totalFound, totalMissing, totalHashes,
	)

	if totalHashes > 0 && len(fullHashSet) == 0 {
		logger.Warn("RuTracker GetIdByHashes: no results for %d requested hashes", totalHashes)
	}

	return fullHashSet, nil
}

func parse(data []byte) (map[string]*int, error) {
	if apiErr := parseAPIError(data); apiErr != nil {
		return nil, apiErr
	}

	var p hashAndNum
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}

	if p.Result == nil {
		logger.Warn("RuTracker API response has no result field: body=%s", truncateBody(data, 500))

		return map[string]*int{}, nil
	}

	return p.Result, nil
}

func parseAPIError(data []byte) error {
	var envelope apiErrorResponse
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil
	}

	if len(envelope.Error) == 0 {
		if envelope.Message != "" && !strings.Contains(string(data), `"result"`) {
			logger.Error("RuTracker API message response: message=%s body=%s", envelope.Message, truncateBody(data, 500))

			return fmt.Errorf("api message: %s", envelope.Message)
		}

		return nil
	}

	var details apiErrorDetails
	if err := json.Unmarshal(envelope.Error, &details); err == nil && details.Text != "" {
		logger.Error("RuTracker API error response: code=%d text=%s", details.Code, details.Text)

		return fmt.Errorf("api error: %s", details.Text)
	}

	var errorText string
	if err := json.Unmarshal(envelope.Error, &errorText); err == nil && errorText != "" {
		logger.Error("RuTracker API error response: error=%s", errorText)

		return fmt.Errorf("api error: %s", errorText)
	}

	logger.Error("RuTracker API error response: body=%s", truncateBody(data, 500))

	return fmt.Errorf("api error: %s", truncateBody(envelope.Error, 200))
}

func countHashesInBatches(batches []string) int {
	total := 0
	for _, batch := range batches {
		if batch == "" {
			continue
		}
		total += strings.Count(batch, ",") + 1
	}

	return total
}

func countTopicIDs(results map[string]*int) (found int, missing int) {
	for _, topicID := range results {
		if topicID != nil {
			found++

			continue
		}
		missing++
	}

	return found, missing
}

func logMissingHashesInBatch(hashBatch string, hashSet map[string]*int, batchNum, batchTotal int) {
	for _, hash := range strings.Split(hashBatch, ",") {
		hash = strings.TrimSpace(hash)
		if hash == "" {
			continue
		}

		topicID, exists := hashSet[hash]
		if !exists {
			logger.Warn(
				"RuTracker GetIdByHashes: batch %d/%d hash %s absent from API response",
				batchNum, batchTotal, hash,
			)

			continue
		}

		if topicID == nil {
			logger.Debug(
				"RuTracker GetIdByHashes: batch %d/%d hash %s present but topic id is null",
				batchNum, batchTotal, hash,
			)
		}
	}
}
