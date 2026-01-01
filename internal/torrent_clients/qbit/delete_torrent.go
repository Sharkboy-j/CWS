package qbit

import (
	"context"
	"cws/logger"
	"fmt"
)

func (s *service) DeleteTorrent(ctx context.Context, hash string, deleteFiles bool) error {
	logger.Debug("Удаление торрента из qBittorrent, hash: %s, deleteFiles: %v", hash, deleteFiles)

	err := s.client.DeleteTorrentsCtx(ctx, []string{hash}, deleteFiles)
	if err != nil {
		logger.Error("Ошибка при удалении торрента: %v", err)

		return fmt.Errorf("failed to delete torrent: %w", err)
	}

	logger.Info("Торрент успешно удален из qBittorrent")

	return nil
}
