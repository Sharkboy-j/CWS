package rutracker

import (
	"cws/config"
	"encoding/json"
	"fmt"
)

type hashAndNum struct {
	Result map[string]*int `json:"result"`
}

func GetIdByHashes(hashes []string, cfg *config.Config) (map[string]*int, error) {
	fullHashSet := make(map[string]*int)

	for _, hashBatch := range hashes {
		urlPart := fmt.Sprintf("v1/get_topic_id?by=hash&val=%s&api_key=%s", hashBatch, cfg.RutrackerApiToken)
		resp, err := makeRequest(cfg, urlPart)
		if err != nil {
			return nil, err
		}

		hashSet, err := parse(resp)
		if err != nil {
			return nil, err
		}

		for k, num := range hashSet {
			fullHashSet[k] = num
		}
	}

	return fullHashSet, nil
}

func parse(data []byte) (map[string]*int, error) {
	var p hashAndNum
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}

	return p.Result, nil
}
