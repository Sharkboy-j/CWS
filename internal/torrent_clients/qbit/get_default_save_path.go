package qbit

import (
	"context"
	"cws/logger"
	"fmt"
)

func (s *service) GetDefaultSavePath(ctx context.Context) (string, error) {
	logger.Debug("Получение пути сохранения по умолчанию из qBittorrent")

	prefs, err := s.client.GetAppPreferencesCtx(ctx)
	if err != nil {
		logger.Error("Ошибка при получении настроек: %v", err)

		return "", fmt.Errorf("failed to get preferences: %w", err)
	}

	if prefs.SavePath != "" {

		return prefs.SavePath, nil
	}

	return "", fmt.Errorf("default save path not found")
}
