package qbit

import (
	"context"
	"cws/logger"
	"fmt"
)

func (s *service) AddTorrentFile(ctx context.Context, torrentFile []byte, savePath string, category string, skipHashCheck bool) error {
	logger.Debug("Добавление торрент файла в qBittorrent, savePath: %s, category: %s, skipHashCheck: %v", savePath, category, skipHashCheck)

	options := make(map[string]string)
	if savePath != "" {
		options["savepath"] = savePath
	}
	if category != "" {
		options["category"] = category
	}
	if skipHashCheck {
		options["skip_checking"] = "true"
	}

	err := s.client.AddTorrentFromMemoryCtx(ctx, torrentFile, options)
	if err != nil {
		logger.Error("Ошибка при добавлении торрент файла: %v", err)

		return fmt.Errorf("failed to add torrent file: %w", err)
	}

	logger.Info("Торрент файл успешно добавлен в qBittorrent")

	return nil
}
