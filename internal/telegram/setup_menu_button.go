package telegram

import (
	"bytes"
	"cws/logger"
	"encoding/json"
	"fmt"
	"net/http"
)

func (bs *BotService) setupMenuButton() error {
	menuButton := map[string]interface{}{
		"type": "commands",
	}

	requestBody := map[string]interface{}{
		"chat_id":     bs.chatId,
		"menu_button": menuButton,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setChatMenuButton", bs.token)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	logger.Info("Menu Button установлен успешно")

	return nil
}
