package monitoring

func (tms *torrentMonitorService) IsTorrentMonitoringActive(chatId int64) bool {
	_, exists := tms.torrentMonitoring[chatId]

	return exists
}
