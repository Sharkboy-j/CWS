package monitoring

func (tms *torrentMonitorService) StopTorrentMonitoring(chatId int64) {
	if monitor, exists := tms.torrentMonitoring[chatId]; exists {
		close(monitor.Stop)
		delete(tms.torrentMonitoring, chatId)
	}
}
