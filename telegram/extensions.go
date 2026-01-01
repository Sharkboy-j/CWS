package telegram

import (
	"context"
	"cws/logger"
	"cws/qBit"
	"cws/telegram/messaging"

	"regexp"
	"strings"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func processRutrackerResults(ctx context.Context, qbClient *qbittorrent.Client, torrentByHash map[string]qbittorrent.Torrent, rutrackerResults map[string]*int) (int, []missingTorrentInfo) {
	foundCount := 0
	for _, topicID := range rutrackerResults {
		if topicID != nil {
			foundCount++
		}
	}

	var missingTorrents []missingTorrentInfo

	if rutrackerResults != nil {
		for hash, torrent := range torrentByHash {
			topicID, exists := rutrackerResults[hash]
			if !exists || topicID == nil {
				props, err := qBit.GetTorrentPropertiesCached(ctx, qbClient, torrent.Hash)
				url := ""
				if err == nil && props != nil {
					url = extractURLFromComment(props.Comment)
				}
				missingTorrents = append(missingTorrents, missingTorrentInfo{
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
			missingTorrents = append(missingTorrents, missingTorrentInfo{
				name: torrent.Name,
				hash: hash,
				url:  url,
			})
		}
	}

	return foundCount, missingTorrents
}

func sendNoClientsMessage(msgSender messaging.MessageSender, stateMgr *StateManager, chatId int64, text string) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
		),
	)
	messageID := stateMgr.GetMenuMessage(chatId)
	newMessageID, err := msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return err
	}
	stateMgr.SetMenuMessage(chatId, newMessageID)

	return nil
}
