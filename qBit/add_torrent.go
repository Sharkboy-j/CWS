package qBit

import (
	"context"
	"cws/logger"
	"fmt"

	"github.com/autobrr/go-qbittorrent"
)

// AddTorrentFile добавляет торрент файл в qBittorrent
func AddTorrentFile(ctx context.Context, client *qbittorrent.Client, torrentFile []byte, savePath string, category string, skipHashCheck bool) error {
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

	err := client.AddTorrentFromMemoryCtx(ctx, torrentFile, options)
	if err != nil {
		logger.Error("Ошибка при добавлении торрент файла: %v", err)
		return fmt.Errorf("failed to add torrent file: %w", err)
	}

	logger.Info("Торрент файл успешно добавлен в qBittorrent")
	return nil
}

// GetCategories получает список категорий из qBittorrent
func GetCategories(ctx context.Context, client *qbittorrent.Client) ([]string, error) {
	logger.Debug("Получение списка категорий из qBittorrent")

	categories, err := client.GetCategoriesCtx(ctx)
	if err != nil {
		logger.Error("Ошибка при получении категорий: %v", err)
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	var categoryList []string
	for name := range categories {
		categoryList = append(categoryList, name)
	}

	logger.Debug("Получено %d категорий", len(categoryList))
	return categoryList, nil
}

// GetDefaultSavePath получает путь сохранения по умолчанию из qBittorrent
func GetDefaultSavePath(ctx context.Context, client *qbittorrent.Client) (string, error) {
	logger.Debug("Получение пути сохранения по умолчанию из qBittorrent")

	prefs, err := client.GetAppPreferencesCtx(ctx)
	if err != nil {
		logger.Error("Ошибка при получении настроек: %v", err)
		return "", fmt.Errorf("failed to get preferences: %w", err)
	}

	if prefs.SavePath != "" {
		return prefs.SavePath, nil
	}

	return "", fmt.Errorf("default save path not found")
}

// GetTorrentSavePaths получает уникальные пути сохранения из существующих торрентов
func GetTorrentSavePaths(ctx context.Context, client *qbittorrent.Client) ([]string, error) {
	logger.Debug("Получение путей сохранения из существующих торрентов")

	torrents, err := client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
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

// DeleteTorrent удаляет торрент из qBittorrent
func DeleteTorrent(ctx context.Context, client *qbittorrent.Client, hash string, deleteFiles bool) error {
	logger.Debug("Удаление торрента из qBittorrent, hash: %s, deleteFiles: %v", hash, deleteFiles)

	err := client.DeleteTorrentsCtx(ctx, []string{hash}, deleteFiles)
	if err != nil {
		logger.Error("Ошибка при удалении торрента: %v", err)
		return fmt.Errorf("failed to delete torrent: %w", err)
	}

	logger.Info("Торрент успешно удален из qBittorrent")
	return nil
}
