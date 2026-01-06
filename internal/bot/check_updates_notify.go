package bot

import (
	"context"
	"crypto/sha256"
	"cws/internal/bot/ui"
	"cws/internal/storage"
	"cws/internal/textutil"
	"cws/logger"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type notifyTorrentItem struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

func (ch *ClientHandler) sendCheckUpdatesNotification(ctx context.Context, chatId int64, results []ClientCheckResult) {
	if ch.notifySender == nil {
		return
	}

	enabled, err := ch.repo.GetNotificationsEnabled(ctx, chatId)
	if err != nil {
		logger.Warn("Failed to read notifications_enabled for user %d: %v", chatId, err)
		enabled = true
	}
	if !enabled {
		return
	}

	missing := collectUniqueMissingByHash(results)
	if len(missing) == 0 {
		prev, getErr := ch.repo.GetCheckUpdatesNotifyState(ctx, chatId)
		if getErr != nil {
			logger.Warn("Failed to load check updates notify state for user %d: %v", chatId, getErr)

			return
		}

		if prev != nil && prev.MessageID != 0 {
			ch.notifySender.DeleteMessage(chatId, prev.MessageID)
		}

		clearErr := ch.repo.SetCheckUpdatesNotifyState(ctx, chatId, storage.CheckUpdatesNotifyState{})
		if clearErr != nil {
			logger.Warn("Failed to clear check updates notify state for user %d: %v", chatId, clearErr)
		}

		return
	}
	missingHashes := joinMissingHashes(missing)
	currentCount := len(missing)

	payloadHash := hashString(missingHashes)

	items := toNotifyItems(missing)
	itemsJSON, marshalErr := json.Marshal(items)
	if marshalErr != nil {
		logger.Warn("Failed to marshal notification items for user %d: %v", chatId, marshalErr)

		return
	}

	text, keyboard := buildCheckUpdatesNotificationMessage(items, 0, ch.mainBotUsername)
	if text == "" {
		return
	}

	prev, err := ch.repo.GetCheckUpdatesNotifyState(ctx, chatId)
	if err != nil {
		logger.Warn("Failed to load check updates notify state for user %d: %v", chatId, err)

		return
	}

	if prev != nil && prev.PayloadHash != "" && prev.PayloadHash == payloadHash {
		return
	}

	newFound := hasNewMissing(missingHashes, prev)
	prevCount := countMissingHashes(prev)
	countChanged := prevCount != currentCount

	var messageID int
	if newFound || countChanged {
		if prev != nil && prev.MessageID != 0 {
			ch.notifySender.DeleteMessage(chatId, prev.MessageID)
		}
		messageID, err = ch.notifySender.SendOrEdit(chatId, 0, text, keyboard)
	} else {
		prevID := 0
		if prev != nil {
			prevID = prev.MessageID
		}
		messageID, err = ch.notifySender.SendOrEdit(chatId, prevID, text, keyboard)
	}
	if err != nil {
		logger.Warn("Failed to send check updates notification to user %d: %v", chatId, err)

		return
	}

	saveErr := ch.repo.SetCheckUpdatesNotifyState(ctx, chatId, storage.CheckUpdatesNotifyState{
		MessageID:     messageID,
		PayloadHash:   payloadHash,
		MissingHashes: missingHashes,
		ItemsJSON:     string(itemsJSON),
	})
	if saveErr != nil {
		logger.Warn("Failed to save check updates notify state for user %d: %v", chatId, saveErr)
	}
}

func countMissingHashes(prev *storage.CheckUpdatesNotifyState) int {
	if prev == nil || prev.MissingHashes == "" {
		return 0
	}

	n := 0
	for _, h := range strings.Split(prev.MissingHashes, "\n") {
		if strings.TrimSpace(h) == "" {
			continue
		}
		n++
	}

	return n
}

func toNotifyItems(missing []missingTorrentInfo) []notifyTorrentItem {
	items := make([]notifyTorrentItem, 0, len(missing))
	for _, mt := range missing {
		items = append(items, notifyTorrentItem{
			Name: mt.name,
			Hash: mt.hash,
			URL:  mt.url,
		})
	}

	return items
}

func collectUniqueMissingByHash(results []ClientCheckResult) []missingTorrentInfo {
	byHash := make(map[string]missingTorrentInfo)
	for _, r := range results {
		for _, mt := range r.MissingTorrents {
			if mt.hash == "" {
				continue
			}
			if _, ok := byHash[mt.hash]; ok {
				continue
			}
			byHash[mt.hash] = mt
		}
	}

	out := make([]missingTorrentInfo, 0, len(byHash))
	for _, v := range byHash {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].hash < out[j].hash })

	return out
}

func joinMissingHashes(missing []missingTorrentInfo) string {
	if len(missing) == 0 {
		return ""
	}

	hashes := make([]string, 0, len(missing))
	for _, mt := range missing {
		if mt.hash != "" {
			hashes = append(hashes, mt.hash)
		}
	}

	return strings.Join(hashes, "\n")
}

func hasNewMissing(currentMissingHashes string, prev *storage.CheckUpdatesNotifyState) bool {
	if currentMissingHashes == "" {
		return false
	}
	if prev == nil || prev.MissingHashes == "" {
		return true
	}

	prevSet := make(map[string]struct{})
	for _, h := range strings.Split(prev.MissingHashes, "\n") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		prevSet[h] = struct{}{}
	}

	for _, h := range strings.Split(currentMissingHashes, "\n") {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if _, ok := prevSet[h]; !ok {
			return true
		}
	}

	return false
}

func buildCheckUpdatesNotificationMessage(items []notifyTorrentItem, page int, mainBotUsername string) (string, *tgbotapi.InlineKeyboardMarkup) {
	if len(items) == 0 {
		return "", nil
	}

	const pageSize = 6
	page, totalPages, start, end := paginateRange(len(items), pageSize, page)

	var b strings.Builder
	if totalPages > 1 {
		b.WriteString("Page: *")
		b.WriteString(intToString(page + 1))
		b.WriteString("/")
		b.WriteString(intToString(totalPages))
		b.WriteString("*\n")
	}
	b.WriteString("\n")

	for i := start; i < end; i++ {
		it := items[i]
		name := textutil.EscapeMarkdown(it.Name)
		hash := textutil.EscapeMarkdown(it.Hash)
		b.WriteString("• `")
		b.WriteString(name)
		b.WriteString("`\n")
		if hash != "" {
			b.WriteString("  `")
			b.WriteString(hash)
			b.WriteString("`\n")
		}
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := start; i < end; i++ {
		it := items[i]
		if it.URL == "" {
			continue
		}
		btnText := truncateButtonLabel(it.Name, 60)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonURL(btnText, it.URL),
		}
		if mainBotUsername != "" && it.Hash != "" {
			monitorURL := "https://t.me/" + mainBotUsername + "?start=monitor_" + it.Hash
			row = append(row, tgbotapi.NewInlineKeyboardButtonURL(ui.IconChart, monitorURL))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
	}

	if totalPages > 1 {
		var nav []tgbotapi.InlineKeyboardButton
		if page > 0 {
			nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(ui.IconArrowLeft, "notify_missing_page_"+intToString(page-1)))
		}
		nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(intToString(page+1)+"/"+intToString(totalPages), "notify_noop"))
		if page < totalPages-1 {
			nav = append(nav, tgbotapi.NewInlineKeyboardButtonData(ui.IconArrowRight, "notify_missing_page_"+intToString(page+1)))
		}
		rows = append(rows, nav)
	}

	if len(rows) == 0 {
		return b.String(), nil
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return b.String(), &kb
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))

	return hex.EncodeToString(sum[:])
}

func intToString(v int) string {
	return strconv.Itoa(v)
}
