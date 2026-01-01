package telegram

import (
	"context"
	"cws/logger"
	"cws/qBit"
	"cws/rutracker_api"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (ch *ClientHandler) CheckClientTorrents(chatId int64, clientID int64) {
	ctx := context.Background()
	messageID := ch.stateMgr.GetMenuMessage(chatId)

	client, err := ch.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для пользователя %d: %v", clientID, chatId, err)
		errorText := "❌ Ошибка при получении данных клиента"
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался проверить несуществующий клиент %d", chatId, clientID)
		errorText := "❌ Клиент не найден или у вас нет доступа"
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	startTime := time.Now()

	checkingText := fmt.Sprintf("🔍 Проверка активных торрентов для клиента *%s*...", client.Name)
	newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, checkingText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)
	messageID = newMessageID

	qbClient, err := qBit.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при подключении к qBit клиенту %s для пользователя %d: %v", client.Name, chatId, err)
		errorText := fmt.Sprintf("❌ Ошибка при подключении к клиенту *%s*:\n`%v`", client.Name, err)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	connectingText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получение списка активных торрентов...", client.Name)
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, connectingText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	messageID = newMessageID

	activeTorrents, err := qBit.GetTorrents(ctx, qbClient)
	if err != nil {
		logger.Error("Ошибка при получении торрентов от клиента %s для пользователя %d: %v", client.Name, chatId, err)
		errorText := fmt.Sprintf("❌ Ошибка при получении торрентов от клиента *%s*:\n`%v`", client.Name, err)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	filteringText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n🔍 Фильтрация по комментарию (rutracker)...", client.Name, len(activeTorrents))
	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, filteringText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	messageID = newMessageID

	torrents, err := qBit.FilterTorrentsByRutrackerComment(ctx, qbClient, activeTorrents)
	if err != nil {
		logger.Error("Ошибка при фильтрации торрентов от клиента %s для пользователя %d: %v", client.Name, chatId, err)
		errorText := fmt.Sprintf("❌ Ошибка при фильтрации торрентов от клиента *%s*:\n`%v`", client.Name, err)
		newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
		if err == nil {
			ch.stateMgr.SetMenuMessage(chatId, newMessageID)
		}
		return
	}

	var rutrackerResults map[string]*int
	torrentByHash := make(map[string]qbittorrent.Torrent)
	if len(torrents) > 0 {
		for _, torrent := range torrents {
			torrentByHash[torrent.InfohashV1] = torrent
		}

		checkingRutrackerText := fmt.Sprintf("✅ Подключение к *%s* успешно\n\n🔍 Получено активных торрентов: *%d*\n\n✅ Отфильтровано по rutracker: *%d*\n\n🔍 Проверка хешей в API рутрекера...", client.Name, len(activeTorrents), len(torrents))
		newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, checkingRutrackerText, nil)
		if err != nil {
			logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
			return
		}
		messageID = newMessageID

		hashBatches := qBit.GetHashStrings(torrents)

		rutrackerResults, err = rutracker_api.GetIdByHashes(hashBatches, ch.cfg)
		if err != nil {
			logger.Error("Ошибка при проверке хешей в API рутрекера для клиента %s для пользователя %d: %v", client.Name, chatId, err)
			errorText := fmt.Sprintf("❌ Ошибка при проверке хешей в API рутрекера от клиента *%s*:\n`%v`", client.Name, err)
			newMessageID, err := ch.msgSender.SendOrEdit(chatId, messageID, errorText, nil)
			if err == nil {
				ch.stateMgr.SetMenuMessage(chatId, newMessageID)
			}
			return
		}
	}

	elapsed := time.Since(startTime)
	durationText := formatDuration(elapsed)

	count := len(torrents)
	var resultText string
	var keyboardRows [][]tgbotapi.InlineKeyboardButton

	if count == 0 {
		resultText = fmt.Sprintf("✅ *%s*\n\n📊 Активных торрентов: *%d*\n\n⏱ Время выполнения: *%s*\n\nНет активных торрентов", client.Name, count, durationText)
	} else {
		foundCount := 0
		for _, topicID := range rutrackerResults {
			if topicID != nil {
				foundCount++
			}
		}

		type missingTorrentInfo struct {
			name string
			hash string
			url  string
		}
		var missingTorrentsInfo []missingTorrentInfo
		if rutrackerResults != nil {
			for hash, torrent := range torrentByHash {
				topicID, exists := rutrackerResults[hash]
				if !exists || topicID == nil {
					props, err := qBit.GetTorrentPropertiesCached(ctx, qbClient, torrent.Hash)
					url := ""
					if err == nil && props != nil {
						url = extractURLFromComment(props.Comment)
					}
					missingTorrentsInfo = append(missingTorrentsInfo, missingTorrentInfo{
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
				missingTorrentsInfo = append(missingTorrentsInfo, missingTorrentInfo{
					name: torrent.Name,
					hash: hash,
					url:  url,
				})
			}
		}

		resultText = fmt.Sprintf("💻 *%s*\n\n📊 Активных торрентов: *%d*\n\n🔍 Отфильтровано по rutracker: *%d*\n\n✅ Актуальных: *%d*/*%d*\n\n", client.Name, len(activeTorrents), count, foundCount, count)

		if len(missingTorrentsInfo) > 0 {
			resultText += fmt.Sprintf("\n\n⚠️ Не найдено в рутрекере: *%d*\n\n", len(missingTorrentsInfo))
			maxDisplay := 20
			displayCount := len(missingTorrentsInfo)
			if displayCount > maxDisplay {
				displayCount = maxDisplay
				resultText += fmt.Sprintf("_Показано первых %d из %d:_\n\n", maxDisplay, len(missingTorrentsInfo))
			}

			for i := 0; i < displayCount; i++ {
				info := missingTorrentsInfo[i]
				resultText += fmt.Sprintf("• %s\n  `%s`\n", info.name, info.hash)

				if info.url != "" {
					buttonText := info.name
					if len(buttonText) > 60 {
						buttonText = buttonText[:57] + "..."
					}
					keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL(buttonText, info.url),
					))
				}
			}
		}

		resultText += fmt.Sprintf("\n⏱ Время выполнения: *%s*", durationText)
	}

	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Повторить проверку", fmt.Sprintf("recheck_client_%d", clientID)),
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	newMessageID, err = ch.msgSender.SendOrEdit(chatId, messageID, resultText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	ch.stateMgr.SetMenuMessage(chatId, newMessageID)

	logger.Debugf("Пользователь %d получил результат проверки клиента %s: %d активных торрентов, время выполнения: %v", chatId, client.Name, count, elapsed)
}

func extractURLFromComment(comment string) string {
	if comment == "" {
		return ""
	}

	urlPattern := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches := urlPattern.FindString(comment)
	if matches != "" {
		return matches
	}

	rutrackerPattern := regexp.MustCompile(`(?:rutracker\.org|rutracker\.cc)/[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches = rutrackerPattern.FindString(comment)
	if matches != "" {
		if !strings.HasPrefix(matches, "http://") && !strings.HasPrefix(matches, "https://") {
			return "https://" + matches
		}
		return matches
	}

	return ""
}
