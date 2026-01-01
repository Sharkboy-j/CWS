package telegram

import (
	"context"
	"cws/database"
	"cws/logger"
	"cws/qBit"
	"fmt"
	"strings"

	"github.com/autobrr/go-qbittorrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TorrentSearchCache struct {
	Results []TorrentSearchResult
	Query   string
}

type TorrentSearchResult struct {
	ClientID   int64
	ClientName string
	Hash       string
	Name       string
}

type TorrentSearchService struct {
	repo               *database.Repository
	msgSender          *MessageSender
	stateMgr           *StateManager
	torrentSearchCache map[int64]*TorrentSearchCache
}

func NewTorrentSearchService(repo *database.Repository, msgSender *MessageSender, stateMgr *StateManager) *TorrentSearchService {
	return &TorrentSearchService{
		repo:               repo,
		msgSender:          msgSender,
		stateMgr:           stateMgr,
		torrentSearchCache: make(map[int64]*TorrentSearchCache),
	}
}

// StartTorrentSearchDialog запускает диалог для поиска торрента
func (tss *TorrentSearchService) StartTorrentSearchDialog(chatId int64) {
	tss.stateMgr.SetUserState(chatId, "search_torrent_query")
	text := "🔎 *Поиск торрента*\n\nВведите хеш или название торрента (частичное или полное):"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "main_menu"),
		),
	)
	messageID := tss.stateMgr.GetMenuMessage(chatId)
	newMessageID, err := tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)
		return
	}
	tss.stateMgr.SetMenuMessage(chatId, newMessageID)
}

// SearchTorrents выполняет поиск торрентов по всем клиентам пользователя
func (tss *TorrentSearchService) SearchTorrents(chatId int64, query string) {
	ctx := context.Background()

	clients, err := tss.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для поиска: %v", err)
		text := "❌ Ошибка при получении списка клиентов"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := tss.stateMgr.GetMenuMessage(chatId)
		tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		return
	}

	if len(clients) == 0 {
		text := "🔎 *Поиск торрента*\n\nКлиенты не найдены. Добавьте клиента для поиска."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить клиента", "add_client"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := tss.stateMgr.GetMenuMessage(chatId)
		tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		return
	}

	query = strings.TrimSpace(query)
	queryUpper := strings.ToUpper(query)

	var searchResults []TorrentSearchResult

	for _, client := range clients {
		qbClient, err := qBit.CreateClient(ctx, client)
		if err != nil {
			logger.Warn("Ошибка при подключении к клиенту %s для поиска: %v", client.Name, err)
			continue
		}

		torrents, err := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
		if err != nil {
			logger.Warn("Ошибка при получении торрентов от клиента %s: %v", client.Name, err)
			continue
		}

		for _, torrent := range torrents {
			hashV1 := strings.ToUpper(torrent.InfohashV1)
			hashV2 := strings.ToUpper(torrent.InfohashV2)
			torrentName := torrent.Name

			matched := false
			if hashV1 == queryUpper || hashV2 == queryUpper {
				matched = true
			} else if strings.Contains(strings.ToLower(torrentName), strings.ToLower(query)) {
				matched = true
			}

			if matched {
				hash := hashV1
				if hash == "" {
					hash = hashV2
				}
				searchResults = append(searchResults, TorrentSearchResult{
					ClientID:   client.ID,
					ClientName: client.Name,
					Hash:       hash,
					Name:       torrentName,
				})
			}
		}
	}

	tss.torrentSearchCache[chatId] = &TorrentSearchCache{
		Results: searchResults,
		Query:   query,
	}

	tss.showSearchResults(chatId, query, searchResults, 0)
}

// ShowSearchResultsPage показывает страницу результатов поиска из кэша
func (tss *TorrentSearchService) ShowSearchResultsPage(chatId int64, page int) {
	cache, exists := tss.torrentSearchCache[chatId]
	if !exists || cache == nil {
		logger.Warn("Пользователь %d запросил страницу %d, но кэш результатов поиска отсутствует", chatId, page)
		text := "Результаты поиска устарели. Выполните поиск заново."
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔎 Новый поиск", "search_torrent"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		messageID := tss.stateMgr.GetMenuMessage(chatId)
		tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		return
	}

	tss.showSearchResults(chatId, cache.Query, cache.Results, page)
}

// GetSearchResult возвращает результат поиска по индексу из кэша
func (tss *TorrentSearchService) GetSearchResult(chatId int64, index int) (*TorrentSearchResult, error) {
	cache, exists := tss.torrentSearchCache[chatId]
	if !exists || cache == nil {
		return nil, fmt.Errorf("кэш результатов поиска не найден")
	}

	if index < 0 || index >= len(cache.Results) {
		return nil, fmt.Errorf("неверный индекс результата: %d", index)
	}

	return &cache.Results[index], nil
}

// ClearSearchCache очищает кэш результатов поиска для пользователя
func (tss *TorrentSearchService) ClearSearchCache(chatId int64) {
	delete(tss.torrentSearchCache, chatId)
}

// showSearchResults отображает результаты поиска
func (tss *TorrentSearchService) showSearchResults(chatId int64, query string, results []TorrentSearchResult, page int) {
	messageID := tss.stateMgr.GetMenuMessage(chatId)

	if len(results) == 0 {
		text := fmt.Sprintf("🔎 *Поиск торрента*\n\n❌ По запросу `%s` ничего не найдено", query)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔎 Новый поиск", "search_torrent"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
			),
		)
		tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
		return
	}

	resultsPerPage := 5
	totalPages := (len(results) + resultsPerPage - 1) / resultsPerPage

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	startIdx := page * resultsPerPage
	endIdx := startIdx + resultsPerPage
	if endIdx > len(results) {
		endIdx = len(results)
	}

	pageResults := results[startIdx:endIdx]

	var text strings.Builder
	if query != "" {
		text.WriteString(fmt.Sprintf("🔎 *Результаты поиска*\n\nЗапрос: `%s`\n\n", query))
	} else {
		text.WriteString("🔎 *Результаты поиска*\n\n")
	}
	text.WriteString(fmt.Sprintf("Найдено: *%d*\n\n", len(results)))

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, result := range pageResults {
		displayName := result.Name
		if len(displayName) > 50 {
			displayName = displayName[:47] + "..."
		}

		// Глобальный индекс в общем списке результатов
		globalIdx := startIdx + i
		text.WriteString(fmt.Sprintf("%d. *%s*\n", globalIdx+1, displayName))
		text.WriteString(fmt.Sprintf("   Hash: `%s`\n\n", result.Hash))

		// Создаем кнопку для каждого результата (используем глобальный индекс)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%d. %s", globalIdx+1, truncatePath(displayName, 999)),
				fmt.Sprintf("search_torrent_select_%d", globalIdx),
			),
		))
	}

	if totalPages > 1 {
		var navButtons []tgbotapi.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", fmt.Sprintf("page_search_%d", page-1)))
		}
		navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d/%d", page+1, totalPages), "page_info"))
		if page < totalPages-1 {
			navButtons = append(navButtons, tgbotapi.NewInlineKeyboardButtonData("Вперёд ▶️", fmt.Sprintf("page_search_%d", page+1)))
		}
		if len(navButtons) > 0 {
			rows = append(rows, navButtons)
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔎 Новый поиск", "search_torrent"),
		tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	tss.msgSender.SendOrEdit(chatId, messageID, text.String(), &keyboard)
}
