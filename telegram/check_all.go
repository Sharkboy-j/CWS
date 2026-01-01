package telegram

import (
	"context"
	"cws/database"
	"cws/logger"
	"cws/qBit"
	"cws/rutracker_api"
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
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	if len(clients) == 0 {
		errorText := "📋 *Проверка активных торрентов*\n\nКлиенты не найдены. Добавьте клиента для проверки."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, &keyboard)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	startTime := time.Now()

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
	resultText := ch.formatAllClientsResult(results, elapsed, &now)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Повторить проверку", "check_torrents"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)

	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, resultText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)

	logger.Debugf("Пользователь %d получил результат проверки всех клиентов: %d клиентов, время выполнения: %v", chatId, len(clients), elapsed)
}

func (ch *ClientHandler) CheckAllClientsAuto(chatId int64) {
	ctx := context.Background()
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	clients, err := ch.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для пользователя %d: %v", chatId, err)
		return
	}

	if len(clients) == 0 {
		logger.Debug("Нет клиентов для пользователя %d", chatId)
		return
	}

	startTime := time.Now()

	var results []ClientCheckResult
	for _, client := range clients {
		result := ch.checkSingleClientSilent(ctx, client)
		results = append(results, result)
	}

	elapsed := time.Since(startTime)
	now := time.Now()
	resultText := ch.formatAllClientsResult(results, elapsed, &now)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Повторить проверку", "check_torrents"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)

	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, resultText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении/отправке сообщения для пользователя %d: %v", chatId, err)
		return
	}

	ch.stateMgr.SetMenuMessage(chatId, newMessageID)

	logger.Debugf("Автоматическая проверка завершена для пользователя %d: %d клиентов, время выполнения: %v", chatId, len(clients), elapsed)
}

func (ch *ClientHandler) checkSingleClientSilent(ctx context.Context, client *database.Client) ClientCheckResult {
	startTime := time.Now()
	result := ClientCheckResult{
		ClientName: client.Name,
	}

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка подключения: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	activeTorrents, err := qBit.GetTorrents(ctx, qbClient)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка получения торрентов: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	result.ActiveTorrents = len(activeTorrents)

	torrents, err := qBit.FilterTorrentsByRutrackerComment(ctx, qbClient, activeTorrents)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка фильтрации: %v", err)
		result.Duration = time.Since(startTime)
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

	hashBatches := qBit.GetHashStrings(torrents)

	rutrackerResults, err := rutracker_api.GetIdByHashes(hashBatches, ch.cfg)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка API рутрекера: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	foundCount := 0
	for _, topicID := range rutrackerResults {
		if topicID != nil {
			foundCount++
		}
	}
	result.FoundInRutracker = foundCount

	if rutrackerResults != nil {
		for hash, torrent := range torrentByHash {
			topicID, exists := rutrackerResults[hash]
			if !exists || topicID == nil {
				props, err := qBit.GetTorrentPropertiesCached(ctx, qbClient, torrent.Hash)
				url := ""
				if err == nil && props != nil {
					url = extractURLFromComment(props.Comment)
				}
				result.MissingTorrents = append(result.MissingTorrents, missingTorrentInfo{
					name: torrent.Name,
					hash: hash,
					url:  url,
				})
			}
		}
	} else {
		for hash, torrent := range torrentByHash {
			props, err := qBit.GetTorrentPropertiesCached(ctx, qbClient, torrent.Hash)
			url := ""
			if err == nil && props != nil {
				url = extractURLFromComment(props.Comment)
			}
			result.MissingTorrents = append(result.MissingTorrents, missingTorrentInfo{
				name: torrent.Name,
				hash: hash,
				url:  url,
			})
		}
	}

	result.Duration = time.Since(startTime)
	return result
}

func (ch *ClientHandler) checkSingleClient(ctx context.Context, client *database.Client, chatId int64, messageID int) ClientCheckResult {
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

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка подключения: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при подключении к клиенту *%s*:\n`%v`", client.Name, err)
		ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		return result
	}

	connectingText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получение списка активных торрентов...", client.Name)
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, connectingText, nil)
	if err == nil {
		messageID = newMessageID
	}

	activeTorrents, err := qBit.GetTorrents(ctx, qbClient)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка получения торрентов: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при получении торрентов от клиента *%s*:\n`%v`", client.Name, err)
		ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		return result
	}
	result.ActiveTorrents = len(activeTorrents)

	filteringText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n🔍 Фильтрация по комментарию (rutracker)...", client.Name, len(activeTorrents))
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, filteringText, nil)
	if err == nil {
		messageID = newMessageID
	}

	torrents, err := qBit.FilterTorrentsByRutrackerComment(ctx, qbClient, activeTorrents)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка фильтрации: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при фильтрации торрентов от клиента *%s*:\n`%v`", client.Name, err)
		ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
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

	hashBatches := qBit.GetHashStrings(torrents)

	rutrackerResults, err := rutracker_api.GetIdByHashes(hashBatches, ch.cfg)
	if err != nil {
		result.Error = fmt.Sprintf("Ошибка API рутрекера: %v", err)
		result.Duration = time.Since(startTime)
		errorText := fmt.Sprintf("❌ Ошибка при проверке хешей в API рутрекера от клиента *%s*:\n`%v`", client.Name, err)
		ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		return result
	}

	foundCount := 0
	for _, topicID := range rutrackerResults {
		if topicID != nil {
			foundCount++
		}
	}
	result.FoundInRutracker = foundCount

	if rutrackerResults != nil {
		for hash, torrent := range torrentByHash {
			topicID, exists := rutrackerResults[hash]
			if !exists || topicID == nil {
				props, err := qBit.GetTorrentPropertiesCached(ctx, qbClient, torrent.Hash)
				url := ""
				if err == nil && props != nil {
					url = extractURLFromComment(props.Comment)
				}
				result.MissingTorrents = append(result.MissingTorrents, missingTorrentInfo{
					name: torrent.Name,
					hash: hash,
					url:  url,
				})
			}
		}
	} else {
		for hash, torrent := range torrentByHash {
			props, err := qBit.GetTorrentPropertiesCached(ctx, qbClient, torrent.Hash)
			url := ""
			if err == nil && props != nil {
				url = extractURLFromComment(props.Comment)
			}
			result.MissingTorrents = append(result.MissingTorrents, missingTorrentInfo{
				name: torrent.Name,
				hash: hash,
				url:  url,
			})
		}
	}

	result.Duration = time.Since(startTime)
	return result
}

func (ch *ClientHandler) formatAllClientsResult(results []ClientCheckResult, totalDuration time.Duration, lastCheckTime *time.Time) string {
	var text strings.Builder
	text.WriteString("📊 *Результаты проверки всех клиентов*\n\n")
	text.WriteString(fmt.Sprintf("⏱ Общее время: *%s*\n", formatDuration(totalDuration)))
	if lastCheckTime != nil {
		text.WriteString(fmt.Sprintf("🕐 Последняя проверка: *%s*\n", lastCheckTime.Format("02.01.2006 15:04:05")))
	}
	text.WriteString("\n---\n\n")

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
				text.WriteString(fmt.Sprintf("   ⚠️ Не найдено: *%d*\n", len(result.MissingTorrents)))
			}
			text.WriteString(fmt.Sprintf("   ⏱ *%s*\n\n", formatDuration(result.Duration)))
		}
	}

	return text.String()
}
