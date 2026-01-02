package quick_actions

import (
	"context"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowSpeedLimitMenu(chatId int64) {
	logger.Debugf("Showing speed limit menu for user %d", chatId)
	ctx := context.Background()
	clients, err := h.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Error getting clients for user %d: %v", chatId, err)
		_, _ = h.msgSender.SendOrEdit(chatId, 0, "Ошибка при получении списка клиентов", nil)

		return
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := "🚦 *Ограничение скорости*\n\nКлиенты не найдены. Добавьте клиента для использования ограничения скорости."
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

		return
	}

	text := "🚦 *Ограничение скорости*\n\nВыберите скорость:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("0.10 МБ/с", "quick_action_limit_speed_10"),
			tgbotapi.NewInlineKeyboardButtonData("1.00 МБ/с", "quick_action_limit_speed_100"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("5.00 МБ/с", "quick_action_limit_speed_500"),
			tgbotapi.NewInlineKeyboardButtonData("10.00 МБ/с", "quick_action_limit_speed_1000"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("20.00 МБ/с", "quick_action_limit_speed_2000"),
			tgbotapi.NewInlineKeyboardButtonData("50.00 МБ/с", "quick_action_limit_speed_5000"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("100.00 МБ/с", "quick_action_limit_speed_10000"),
			tgbotapi.NewInlineKeyboardButtonData("500.00 МБ/с", "quick_action_limit_speed_50000"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏ Ввести вручную", "quick_action_limit_speed_custom"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚫 Убрать все ограничения", "quick_action_remove_speed_limits"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "quick_actions"),
		),
	)

	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (h *Handler) StartCustomSpeedLimitDialog(chatId int64) {
	logger.Debugf("Starting custom speed limit dialog for user %d", chatId)
	if h.stateSetter != nil {
		h.stateSetter.SetUserState(chatId, "custom_speed_limit")
	}
	text := "🚦 *Ограничение скорости*\n\nВведите скорость в МБ/с (например: 2.5):"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "quick_actions"),
		),
	)
	messageID := h.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Error sending message for user %d: %v", chatId, err)

		return
	}
	h.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (h *Handler) HandleLimitSpeedBytes(chatId int64, limitBytesPerSec int64) {
	logger.Debugf("Handling limit speed for user %d: %d bytes/s", chatId, limitBytesPerSec)
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

	limitMBPerSec := float64(limitBytesPerSec) / (1024 * 1024)

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

		err = qbClient.SetGlobalSpeedLimits(ctx, limitBytesPerSec, limitBytesPerSec)
		if err != nil {
			logger.Error("Error setting speed limits for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
		}
	}

	if failCount > 0 {
		messageID := h.stateMgr.GetMenuMessage(chatId)
		text := fmt.Sprintf("🚦 *Ограничение скорости до %.2f МБ/с*\n\n", limitMBPerSec)
		text += fmt.Sprintf("❌ Ошибки (%d):\n", failCount)
		for _, name := range failedClients {
			text += fmt.Sprintf("  • %s\n", name)
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

		return
	}

	if h.cmdHdlr != nil {
		h.cmdHdlr.ShowMainMenu(chatId)
	}
}

func (h *Handler) HandleRemoveSpeedLimits(chatId int64) {
	logger.Debugf("Handling remove speed limits for user %d", chatId)
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

	limitBytesPerSec := int64(0)

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

		err = qbClient.SetGlobalSpeedLimits(ctx, limitBytesPerSec, limitBytesPerSec)
		if err != nil {
			logger.Error("Error removing speed limits for client %s: %v", client.Name, err)
			failCount++
			failedClients = append(failedClients, client.Name)
		} else {
			successCount++
		}
	}

	if failCount > 0 {
		messageID := h.stateMgr.GetMenuMessage(chatId)
		text := "🚫 *Снятие ограничений скорости*\n\n"
		text += fmt.Sprintf("❌ Ошибки (%d):\n", failCount)
		for _, name := range failedClients {
			text += fmt.Sprintf("  • %s\n", name)
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

		return
	}

	if h.cmdHdlr != nil {
		h.cmdHdlr.ShowMainMenu(chatId)
	}
}
