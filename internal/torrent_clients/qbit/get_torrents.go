package qbit

import (
	"context"
	"fmt"
	"strings"

	"github.com/autobrr/go-qbittorrent"
)

func (s *service) GetTorrents(ctx context.Context) ([]qbittorrent.Torrent, error) {
	var (
		filter = "seeding"
		//filter   = "all"
		category string
		tag      string
		hashes   []string
	)
	req := qbittorrent.TorrentFilterOptions{
		Filter:   qbittorrent.TorrentFilter(strings.ToLower(filter)),
		Category: category,
		Tag:      tag,
		Hashes:   hashes,
	}

	torrents, err := s.client.GetTorrentsCtx(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not get torrents %v", err)
	}

	return torrents, nil
}

func (s *service) GetTorrentsCtx(ctx context.Context, options qbittorrent.TorrentFilterOptions) ([]qbittorrent.Torrent, error) {
	return s.client.GetTorrentsCtx(ctx, options)
}
