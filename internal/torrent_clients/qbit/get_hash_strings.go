package qbit

import (
	"strings"

	"github.com/autobrr/go-qbittorrent"
)

func GetHashStrings(torrents []qbittorrent.Torrent) []string {
	hashes := make([]string, 0)
	hashStrings := make([]string, 0)
	counter := 0

	for _, v := range torrents {
		hashes = append(hashes, v.InfohashV1)
		counter++
		if counter == 99 {
			hashStrings = append(hashStrings, strings.Join(hashes, ","))
			hashes = make([]string, 0)
			counter = 0
		}
	}

	if len(hashes) > 0 {
		hashStrings = append(hashStrings, strings.Join(hashes, ","))
	}

	return hashStrings
}
