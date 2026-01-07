package monitoring

import (
	"context"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
)

type TorrentMonitorService interface {
	StartTorrentMonitoring(ctx context.Context, chatId int64, clientID int64, hash string)
	StopTorrentMonitoring(chatId int64)
	IsTorrentMonitoringActive(chatId int64) bool
}

type TorrentMonitor struct {
	ChatID    int64
	ClientID  int64
	Hash      string
	Stop      chan bool
	MessageID int
}

type torrentMonitorService struct {
	repo              *storage.Repository
	msgSender         messaging.MessageSender
	getMenuMessage    func(chatId int64) int
	setMenuMessage    func(chatId int64, messageID int)
	torrentMonitoring map[int64]*TorrentMonitor
}

func NewTorrentMonitorService(repo *storage.Repository, msgSender messaging.MessageSender, getMenuMessage func(chatId int64) int, setMenuMessage func(chatId int64, messageID int)) TorrentMonitorService {
	return &torrentMonitorService{
		repo:              repo,
		msgSender:         msgSender,
		getMenuMessage:    getMenuMessage,
		setMenuMessage:    setMenuMessage,
		torrentMonitoring: make(map[int64]*TorrentMonitor),
	}
}
