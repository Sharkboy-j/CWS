package qbit

import (
	"strings"

	"github.com/autobrr/go-qbittorrent"
)

func FindTorrentByName(torrents []qbittorrent.Torrent, name string) *qbittorrent.Torrent {
	for _, torrent := range torrents {
		if strings.EqualFold(torrent.Name, name) {

			return &torrent
		}
	}

	return nil
}
