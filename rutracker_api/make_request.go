package rutracker_api

import (
	"cws/config"
	"fmt"
	"io"
	"net/http"
)

func makeRequest(cfg *config.Config, urlPart string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", cfg.RutrackerHost, urlPart)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
