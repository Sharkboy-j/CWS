package qbit

import (
	"context"
	"cws/logger"
	"fmt"
	"sort"

	"github.com/autobrr/go-qbittorrent"
)

func (s *service) GetTorrentSavePaths(ctx context.Context) ([]string, error) {
	logger.Debug("Fetching save paths from existing torrents")

	torrents, err := s.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
	if err != nil {
		logger.Error("Failed to fetch torrents: %v", err)

		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	pathCounts := make(map[string]int)
	for _, torrent := range torrents {
		if torrent.SavePath != "" {
			pathCounts[torrent.SavePath]++
		}
	}

	type savePathStat struct {
		path  string
		count int
	}

	stats := make([]savePathStat, 0, len(pathCounts))
	for path, count := range pathCounts {
		stats = append(stats, savePathStat{path: path, count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].count != stats[j].count {
			return stats[i].count > stats[j].count
		}

		return stats[i].path < stats[j].path
	})

	paths := make([]string, 0, len(stats))
	for _, st := range stats {
		paths = append(paths, st.path)
	}

	logger.Debug("Found %d unique save paths", len(paths))

	return paths, nil
}
