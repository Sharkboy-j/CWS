package quick_actions

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowSpeedLimitMenu(chatId int64) {
	logger.Debugf("Showing speed limit menu for user %d", chatId)
	ctx := context.Background()
	clients, err := h.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Error getting clients for user %d: %v", chatId, err)
		_, _ = h.msgSender.SendOrEdit(chatId, 0, ui.Msg(ui.MsgClientsListError), nil)

		return
	}

	messageID := h.stateMgr.GetMenuMessage(chatId)

	if len(clients) == 0 {
		text := ui.Msg(ui.MsgSpeedLimitMenuNoClientsText)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		newMessageID, sendErr := h.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		if sendErr != nil {
			logger.Error("Error sending message for user %d: %v", chatId, sendErr)

			return
		}
		h.stateMgr.SetMenuMessage(chatId, newMessageID)

		return
	}

	text := ui.Msg(ui.MsgSpeedLimitMenuText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.Speed10),
			ui.Button(ui.Speed100),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.Speed500),
			ui.Button(ui.Speed1000),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.Speed2000),
			ui.Button(ui.Speed5000),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.Speed10000),
			ui.Button(ui.Speed50000),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SpeedCustom),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.Button(ui.SpeedRemove),
		),
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Back, "quick_actions"),
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
	text := ui.Msg(ui.MsgSpeedLimitCustomPromptText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "quick_actions"),
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
	ctx, clients, _, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
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
		text := ui.Msgf(ui.MsgSpeedLimitAppliedHeaderFmt, limitMBPerSec)
		_ = h.sendOrEditResultWithMainMenu(chatId, messageID, text, successCount, failCount, failedClients)

		return
	}

	if h.cmdHdlr != nil {
		h.cmdHdlr.ShowMainMenu(chatId)
	}
}

func (h *Handler) HandleRemoveSpeedLimits(chatId int64) {
	logger.Debugf("Handling remove speed limits for user %d", chatId)
	ctx, clients, _, ok := h.getClientsAndMenuMessageOrReply(chatId)
	if !ok {
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
		text := ui.Msg(ui.MsgSpeedLimitsRemovedHeader)
		_ = h.sendOrEditResultWithMainMenu(chatId, messageID, text, successCount, failCount, failedClients)

		return
	}

	if h.cmdHdlr != nil {
		h.cmdHdlr.ShowMainMenu(chatId)
	}
}
