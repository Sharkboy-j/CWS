package sweek

import (
	"context"
	"cws/config"
	"cws/qBit"
	"cws/rutracker_api"
	"cws/telegram"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/autobrr/go-qbittorrent"
)

const rutrackerHostStr = "rutracker"
const rutrackerHostStr2 = "t-ru.org"

func doCheck(ctx context.Context, config *config.Config) error {
	torrents, err := qBit.GetTorrents(ctx, qBittorrentClient)

	//log.Println(fmt.Sprintf("found %d hashes on client...", len(torrents)))

	rutrackerTorrents := make([]qbittorrent.Torrent, 0)

	for _, torrent := range torrents {
		trackers, trErr := qBit.GetTrackers(ctx, qBittorrentClient, torrent.Hash)
		if trErr != nil {
			log.Println(trErr)

			return trErr
		}

		if slices.ContainsFunc(trackers, isRutracker) {
			rutrackerTorrents = append(rutrackerTorrents, torrent)
		} else {
			log.Print(fmt.Sprintf("torrent %s not found on rutracker", torrent.Name))
		}

		//if torrent.TrackersCount > 1 {
		//	trackers, trErr := qBit.GetTrackers(ctx, qBittorrentClient, torrent.Hash)
		//	if trErr != nil {
		//		log.Println(trErr)
		//
		//		return trErr
		//	}
		//
		//	if slices.ContainsFunc(trackers, isRutracker) {
		//		rutrackerTorrents = append(rutrackerTorrents, torrent)
		//	} else {
		//		log.Print(fmt.Sprintf("torrent %s not found on rutracker", torrent.Name))
		//	}
		//} else {
		//	if strings.Contains(torrent.Tracker, rutrackerHostStr) || strings.Contains(torrent.Tracker, rutrackerHostStr2) {
		//		rutrackerTorrents = append(rutrackerTorrents, torrent)
		//	} else {
		//		log.Print(fmt.Sprintf("torrent %s not found on rutracker", torrent.Name))
		//	}
		//}
	}

	hashStrings := qBit.GetHashStrings(rutrackerTorrents)
	result, err := rutracker_api.GetIdByHashes(hashStrings, config)
	if err != nil {
		log.Println(err)

		return err
	}

	err = validateHash(ctx, config, result)
	if err != nil {
		return err
	}

	return nil
}

func isRutracker(tracker qbittorrent.TorrentTracker) bool {
	return strings.Contains(tracker.Url, rutrackerHostStr) || strings.Contains(tracker.Url, rutrackerHostStr2)
}

func validateHash(ctx context.Context, config *config.Config, result map[string]*int) error {
	notFoundOnTrackerHashes := make([]string, 0)
	for key, val := range result {
		if val == nil {
			notFoundOnTrackerHashes = append(notFoundOnTrackerHashes, key)
		}
	}

	for _, hashV1 := range notFoundOnTrackerHashes {
		err, props := qBit.GetProperties(ctx, qBittorrentClient, hashV1)
		if err != nil {
			return err
		}

		comment := strings.Replace(props.Comment, ".org", ".net", 1)
		fmt.Println(fmt.Sprintf("%s|%s|%s", props.Name, hashV1, comment))

		if telegramBotInited {
			err = telegram.SendMsg(
				telegramBotClient,
				fmt.Sprintf("`%s` \n `%s` \n %s", props.Name, hashV1, comment),
				config.ChatId)
			if err != nil {
				log.Printf(err.Error())
			}
		}
	}

	return nil
}
