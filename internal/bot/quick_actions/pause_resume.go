package quick_actions

import (
	"context"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) HandlePauseAllTorrents(chatId int64) {
	logger.Debugf("Handling pause all torrents for user %d", chatId)
	ctx := context.Background()
	clients, err := h.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Error getting clients for user %d: %v", chatId, err)
		_, _ = h.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при получении списка клиентов", nil)

		return
	}

	if len(clients) == 0 {
		_, _ = h.msgSender.SendOrEdit(chatId, 0, "❌ Клиенты не найдены", nil)

		return
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)
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

	if failCount > 0 {
		text += fmt.Sprintf("\n❌ Ошибки (%d):\n", failCount)
		for _, name := range failedClients {
			text += fmt.Sprintf("  • %s\n", name)
		}
	}

	text += fmt.Sprintf("\nВсего обработано: %d успешно, %d с ошибками", successCount, failCount)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (h *Handler) HandleResumeAllTorrents(chatId int64) {
	logger.Debugf("Handling resume all torrents for user %d", chatId)
	ctx := context.Background()
	clients, err := h.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Error getting clients for user %d: %v", chatId, err)
		_, _ = h.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка при получении списка клиентов", nil)

		return
	}

	if len(clients) == 0 {
		_, _ = h.msgSender.SendOrEdit(chatId, 0, "❌ Клиенты не найдены", nil)

		return
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)
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

	if failCount > 0 {
		text += fmt.Sprintf("\n❌ Ошибки (%d):\n", failCount)
		for _, name := range failedClients {
			text += fmt.Sprintf("  • %s\n", name)
		}
	}

	text += fmt.Sprintf("\nВсего обработано: %d успешно, %d с ошибками", successCount, failCount)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}
