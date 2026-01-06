package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/rutracker"
	"cws/internal/storage"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
	"fmt"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type missingTorrentInfo struct {
	name string
	hash string
	url  string
}

type ClientCheckResult struct {
	ClientName       string
	ActiveTorrents   int
	FilteredTorrents int
	FoundInRutracker int
	MissingTorrents  []missingTorrentInfo
	Error            string
	Duration         time.Duration
}

func (ch *ClientHandler) CheckAllClients(chatId int64) {
	ctx := context.Background()
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		errorText := "❌ Ошибка при получении списка клиентов"
		newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if sendErr == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}

		return
	}

	if len(clients) == 0 {
		errorText := "📋 *Проверка активных торрентов*\n\nКлиенты не найдены. Добавьте клиента для проверки."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.AddClient),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		newMessageID, sendErr := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if sendErr == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}

		return
	}

	startTime := time.Now()

	delete(ch.checkResultsCache, chatId)

	checkingText := fmt.Sprintf("🔍 Проверка активных торрентов для *%d* клиентов...", len(clients))
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, checkingText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	messageID = newMessageID

	var results []ClientCheckResult
	for i, client := range clients {
		progressText := fmt.Sprintf("🔍 Проверка клиентов...\n\n*%d* из *%d*\n\nПроверка: *%s*", i+1, len(clients), client.Name)
		newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, progressText, nil)
		if err == nil {
			messageID = newMessageID
		}

		result := ch.checkSingleClient(ctx, client, chatId, messageID)
		results = append(results, result)
		if result.Error == "" {
			messageID = ch.stateMgr.GetMenuMessage(chatId)
		}
	}

	elapsed := time.Since(startTime)
	now := time.Now()
	_, err = ch.sendAllClientsCheckResult(ctx, chatId, messageID, results, elapsed, now)
	if err != nil {
		return
	}

	logger.Debugf("Пользователь %d получил результат проверки всех клиентов: %d клиентов, время выполнения: %v", chatId, len(clients), elapsed)

	go ch.sendCheckUpdatesNotification(ctx, chatId, results)
}

func (ch *ClientHandler) sendAllClientsCheckResult(ctx context.Context, chatId int64, messageID int,
	results []ClientCheckResult, elapsed time.Duration, now time.Time) (int, error) {
	allMissingTorrents := ch.collectAllMissingTorrents(results)
	nowCopy := now
	ch.checkResultsCache[chatId] = &CheckResultsCache{
		Results:            results,
		TotalDuration:      elapsed,
		LastCheckTime:      &nowCopy,
		AllMissingTorrents: allMissingTorrents,
	}

	resultText, resultKeyboard := ch.formatAllClientsResult(ctx, chatId, results, elapsed, &nowCopy, 0)

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	if resultKeyboard != nil {
		keyboardRows = resultKeyboard.InlineKeyboard
	}
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.RepeatCheck),
		ui.Button(ui.MainMenu),
	))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, resultText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении/отправке сообщения для пользователя %d: %v", chatId, err)

		return 0, err
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)

	return newMessageID, nil
}

func (ch *ClientHandler) ShowMissingTorrentsPage(chatId int64, page int) {
	ctx := context.Background()
	cache, exists := ch.checkResultsCache[chatId]
	if !exists || cache == nil {
		logger.Warn("Пользователь %d запросил страницу %d, но кэш результатов отсутствует", chatId, page)
		_, _ = ch.msgSender.SendOrEdit(chatId, 0, "Результаты проверки устарели. Запустите проверку заново.", nil)

		return
	}

	messageID := ch.stateMgr.GetMenuMessage(chatId)
	resultText, resultKeyboard := ch.formatAllClientsResult(ctx, chatId, cache.Results, cache.TotalDuration, cache.LastCheckTime, page)

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	if resultKeyboard != nil {
		keyboardRows = resultKeyboard.InlineKeyboard
	}
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.RepeatCheck),
		ui.Button(ui.MainMenu),
	))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, resultText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
}

func (ch *ClientHandler) checkSingleClient(ctx context.Context, client *storage.Client, chatId int64, messageID int) ClientCheckResult {
	startTime := time.Now()
	result := ClientCheckResult{
		ClientName: client.Name,
	}

	checkingText := fmt.Sprintf("🔍 Проверка активных торрентов для клиента *%s*...", client.Name)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, checkingText, nil)
	if err == nil {
		ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		messageID = newMessageID
	}

	qbClient, err := qbit.New(ctx, client)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка подключения: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при подключении к клиенту *%s*:\n`%v`", client.Name, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)

		return result
	}

	connectingText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получение списка активных торрентов...", client.Name)
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, connectingText, nil)
	if err == nil {
		messageID = newMessageID
	}

	activeTorrents, err := qbClient.GetTorrents(ctx)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка получения торрентов: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при получении торрентов от клиента *%s*:\n`%v`", client.Name, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)

		return result
	}
	result.ActiveTorrents = len(activeTorrents)

	filteringText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n🔍 Фильтрация по комментарию (rutracker)...", client.Name, len(activeTorrents))
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, filteringText, nil)
	if err == nil {
		messageID = newMessageID
	}

	torrents, err := qbClient.FilterTorrentsByRutrackerComment(ctx, activeTorrents)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка фильтрации: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при фильтрации торрентов от клиента *%s*:\n`%v`", client.Name, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)

		return result
	}
	result.FilteredTorrents = len(torrents)

	if len(torrents) == 0 {
		result.Duration = time.Since(startTime)

		return result
	}

	torrentByHash := make(map[string]qbittorrent.Torrent)
	for _, torrent := range torrents {
		torrentByHash[torrent.InfohashV1] = torrent
	}

	checkingRutrackerText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n✅ Отфильтровано по rutracker: *%d*\n\n🔍 Проверка хешей в API рутрекера...", client.Name, len(activeTorrents), len(torrents))
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, checkingRutrackerText, nil)
	if err == nil {
		messageID = newMessageID
	}

	hashBatches := qbit.GetHashStrings(torrents)

	rutrackerResults, err := rutracker.GetIdByHashes(hashBatches, ch.cfg)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка API рутрекера: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при проверке хешей в API рутрекера от клиента *%s*:\n`%v`", client.Name, err)
		_, _ = ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)

		return result
	}

	result.FoundInRutracker, result.MissingTorrents = processRutrackerResults(ctx, qbClient, torrentByHash, rutrackerResults)
	result.Duration = time.Since(startTime)

	return result
}

func (ch *ClientHandler) collectAllMissingTorrents(results []ClientCheckResult) []missingTorrentInfo {
	urlMap := make(map[string]missingTorrentInfo)

	for _, result := range results {
		for _, torrent := range result.MissingTorrents {
			if torrent.url != "" {
				if _, exists := urlMap[torrent.url]; !exists {
					urlMap[torrent.url] = torrent
				}
			}
		}
	}

	allTorrents := make([]missingTorrentInfo, 0, len(urlMap))
	for _, torrent := range urlMap {
		allTorrents = append(allTorrents, torrent)
	}

	return allTorrents
}

func (ch *ClientHandler) formatAllClientsResult(ctx context.Context, chatId int64, results []ClientCheckResult, totalDuration time.Duration, lastCheckTime *time.Time, page int) (string, *tgbotapi.InlineKeyboardMarkup) {
	var text strings.Builder
	text.WriteString("📊 *Результаты проверки всех клиентов*\n\n")
	text.WriteString(fmt.Sprintf("⏱ Общее время: *%s*\n", formatDuration(totalDuration)))
	if lastCheckTime != nil {
		formattedTime := ch.formatTimeInUserTimezone(ctx, chatId, *lastCheckTime)
		text.WriteString(fmt.Sprintf("🕐 Последняя проверка: *%s*\n", formattedTime))
	}
	text.WriteString("\n---\n\n")

	var keyboardRows [][]tgbotapi.InlineKeyboardButton

	for _, result := range results {
		if result.Error != "" {
			text.WriteString(fmt.Sprintf("❌ *%s*\n", result.ClientName))
			text.WriteString(fmt.Sprintf("   `%s`\n\n", result.Error))
		} else {
			text.WriteString(fmt.Sprintf("💻 *%s*\n", result.ClientName))
			text.WriteString(fmt.Sprintf("   📊 Активных: *%d*\n", result.ActiveTorrents))
			text.WriteString(fmt.Sprintf("   🔍 Отфильтровано: *%d*\n", result.FilteredTorrents))
			text.WriteString(fmt.Sprintf("   ✅ Актуальных: *%d*/*%d*\n", result.FoundInRutracker, result.FilteredTorrents))
			if len(result.MissingTorrents) > 0 {
				text.WriteString(fmt.Sprintf("   ⚠️ Не найдено: *%d*\n\n", len(result.MissingTorrents)))

				maxDisplay := 20
				displayCount := len(result.MissingTorrents)
				if displayCount > maxDisplay {
					displayCount = maxDisplay
					text.WriteString(fmt.Sprintf("   _Показано первых %d из %d:_\n\n", maxDisplay, len(result.MissingTorrents)))
				}

				for i := 0; i < displayCount; i++ {
					info := result.MissingTorrents[i]
					text.WriteString(fmt.Sprintf("   • `%s`\n     `%s`\n", info.name, info.hash))
				}
				text.WriteString("\n")
			}
			text.WriteString(fmt.Sprintf("   ⏱ *%s*\n\n", formatDuration(result.Duration)))
		}
	}

	allMissingTorrents := ch.collectAllMissingTorrents(results)

	const buttonsPerPage = 5

	if len(allMissingTorrents) > 0 {
		pageOut, totalPages, startIdx, endIdx := paginateRange(len(allMissingTorrents), buttonsPerPage, page)

		if len(allMissingTorrents) > buttonsPerPage {
			for i := startIdx; i < endIdx; i++ {
				info := allMissingTorrents[i]
				keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(buildMissingTorrentRowButtons(info)...))
			}

			if totalPages > 1 {
				var navButtons []tgbotapi.InlineKeyboardButton
				if pageOut > 0 {
					navButtons = append(navButtons, ui.ButtonWithData(ui.PrevPage, fmt.Sprintf("page_missing_%d", pageOut-1)))
				}
				navButtons = append(navButtons, ui.Data(fmt.Sprintf("%d/%d", pageOut+1, totalPages), "page_info"))
				if pageOut < totalPages-1 {
					navButtons = append(navButtons, ui.ButtonWithData(ui.NextPage, fmt.Sprintf("page_missing_%d", pageOut+1)))
				}
				if len(navButtons) > 0 {
					keyboardRows = append(keyboardRows, navButtons)
				}
			}
		} else {
			for i := 0; i < len(allMissingTorrents); i++ {
				info := allMissingTorrents[i]
				keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(buildMissingTorrentRowButtons(info)...))
			}
		}
	}

	var keyboard *tgbotapi.InlineKeyboardMarkup
	if len(keyboardRows) > 0 {
		keyboard = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboardRows}
	}

	return text.String(), keyboard
}
