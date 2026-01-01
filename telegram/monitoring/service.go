package monitoring

import (
	"context"
	"cws/store"
	"cws/telegram/messaging"
	"regexp"
	"strings"
)

type TorrentMonitorService interface {
	StartTorrentMonitoring(ctx context.Context, chatId int64, clientID int64, hash string)
	StopTorrentMonitoring(chatId int64)
}

type TorrentMonitor struct {
	ChatID   int64
	ClientID int64
	Hash     string
	Stop     chan bool
}

type torrentMonitorService struct {
	repo              *store.Repository
	msgSender         messaging.MessageSender
	getMenuMessage    func(chatId int64) int
	setMenuMessage    func(chatId int64, messageID int)
	torrentMonitoring map[int64]*TorrentMonitor
}

func NewTorrentMonitorService(repo *store.Repository, msgSender messaging.MessageSender, getMenuMessage func(chatId int64) int, setMenuMessage func(chatId int64, messageID int)) TorrentMonitorService {
	return &torrentMonitorService{
		repo:              repo,
		msgSender:         msgSender,
		getMenuMessage:    getMenuMessage,
		setMenuMessage:    setMenuMessage,
		torrentMonitoring: make(map[int64]*TorrentMonitor),
	}
}

func extractURLFromComment(comment string) string {
	if comment == "" {
		return ""
	}

	urlPattern := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches := urlPattern.FindString(comment)
	if matches != "" {
		return matches
	}

	rutrackerPattern := regexp.MustCompile(`(?:rutracker\.org|rutracker\.cc)/[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches = rutrackerPattern.FindString(comment)
	if matches != "" {
		if !strings.HasPrefix(matches, "http://") && !strings.HasPrefix(matches, "https://") {
			return "https://" + matches
		}

		return matches
	}

	return ""
}
