package bot

import (
	"context"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/logger"
	"encoding/json"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartNotifyBot(ctx context.Context, bot *tgbotapi.BotAPI, repo *storage.Repository) {
	if bot == nil || repo == nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	sender := messaging.NewMessageSender(bot)

	logger.Info("Notification bot started and waiting for updates...")

	for {
		select {
		case <-ctx.Done():
			bot.StopReceivingUpdates()

			return
		case update := <-updates:
			if update.Message != nil {
				handleNotifyMessage(ctx, sender, repo, update.Message)

				continue
			}

			if update.CallbackQuery != nil {
				handleNotifyCallback(ctx, bot, sender, repo, update.CallbackQuery)
			}
		}
	}
}

func handleNotifyMessage(ctx context.Context, sender messaging.MessageSender, repo *storage.Repository, msg *tgbotapi.Message) {
	if msg == nil || msg.Chat == nil {
		return
	}

	if !msg.IsCommand() || msg.Command() != "start" {
		return
	}

	chatId := msg.Chat.ID
	if err := repo.SetNotifyBotSubscribed(ctx, chatId, true); err != nil {
		logger.Warn("Failed to set notify bot subscribed for user %d: %v", chatId, err)

		return
	}

	_, _ = sender.SendOrEdit(chatId, 0, "✅ Notifications bot activated.", nil)
}

func handleNotifyCallback(ctx context.Context, bot *tgbotapi.BotAPI, sender messaging.MessageSender, repo *storage.Repository, query *tgbotapi.CallbackQuery) {
	callback := tgbotapi.NewCallback(query.ID, "")
	_, _ = bot.Request(callback)

	if query.Message == nil {
		return
	}

	chatId := query.Message.Chat.ID
	messageID := query.Message.MessageID
	data := query.Data

	if data == "notify_noop" {
		return
	}

	if !strings.HasPrefix(data, "notify_missing_page_") {
		return
	}

	pageStr := strings.TrimPrefix(data, "notify_missing_page_")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return
	}

	state, err := repo.GetCheckUpdatesNotifyState(ctx, chatId)
	if err != nil {
		logger.Warn("Failed to load notify state for user %d: %v", chatId, err)

		return
	}

	if state == nil || state.ItemsJSON == "" {
		return
	}

	var items []notifyTorrentItem
	unmarshalErr := json.Unmarshal([]byte(state.ItemsJSON), &items)
	if unmarshalErr != nil {
		logger.Warn("Failed to decode notify items for user %d: %v", chatId, unmarshalErr)

		return
	}

	text, keyboard := buildCheckUpdatesNotificationMessage(items, page, "")
	if text == "" {
		return
	}
	_, _ = sender.SendOrEdit(chatId, messageID, text, keyboard)
}
