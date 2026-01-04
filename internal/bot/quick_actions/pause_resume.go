package quick_actions

import (
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
	"fmt"
)

func (h *Handler) HandlePauseAllTorrents(chatId int64) {
	logger.Debugf("Handling pause all torrents for user %d", chatId)
	ctx, clients, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}
	text := "⏸ *Остановка всех раздач*\n\n"
	var successCount, failCount int
	var failedClients []string

	for _, client := range clients {
		qbClient, err := qbit.New(ctx, client)
		if err != nil {
			logger.Error("Error connecting to qBit client %s for user %d: %v", client.Name, chatId, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		err = qbClient.PauseAllTorrents(ctx)
		if err != nil {
			logger.Error("Error pausing all torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
			text += fmt.Sprintf("✅ *%s* - остановлено\n", client.Name)
		}
	}

	if !h.sendOrEditResultWithMainMenu(chatId, messageID, text, successCount, failCount, failedClients) {
		return
	}
}

func (h *Handler) HandleResumeAllTorrents(chatId int64) {
	logger.Debugf("Handling resume all torrents for user %d", chatId)
	ctx, clients, messageID, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
		return
	}
	text := "▶ *Запуск всех раздач*\n\n"
	var successCount, failCount int
	var failedClients []string

	for _, client := range clients {
		qbClient, err := qbit.New(ctx, client)
		if err != nil {
			logger.Error("Error connecting to qBit client %s for user %d: %v", client.Name, chatId, err)
			failCount++
			failedClients = append(failedClients, client.Name)

			continue
		}

		err = qbClient.ResumeAllTorrents(ctx)
		if err != nil {
			logger.Error("Error resuming all torrents for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
			text += fmt.Sprintf("✅ *%s* - запущено\n", client.Name)
		}
	}

	if !h.sendOrEditResultWithMainMenu(chatId, messageID, text, successCount, failCount, failedClients) {
		return
	}
}
