package qbit

import (
	"context"
	"cws/logger"
	"fmt"

	"github.com/autobrr/go-qbittorrent"
)

func (s *service) GetTorrentSavePaths(ctx context.Context) ([]string, error) {
	logger.Debug("Получение путей сохранения из существующих торрентов")

	torrents, err := s.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
	if err != nil {
		logger.Error("Ошибка при получении торрентов: %v", err)

		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	pathMap := make(map[string]bool)
	for _, torrent := range torrents {
		if torrent.SavePath != "" {
			pathMap[torrent.SavePath] = true
		}
	}

	var paths []string
	for path := range pathMap {
		paths = append(paths, path)
	}

	logger.Debug("Найдено %d уникальных путей сохранения", len(paths))

	return paths, nil
}
