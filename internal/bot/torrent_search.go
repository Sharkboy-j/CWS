package bot

import (
	"context"
	"cws/internal/bot/ui"
	"cws/internal/dialogstate"
	"cws/internal/storage"
	"cws/internal/telegram/messaging"
	"cws/internal/torrent_clients/qbit"
	"cws/logger"
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
	repo               *storage.Repository
	msgSender          messaging.MessageSender
	stateMgr           *StateManager
	torrentSearchCache map[int64]*TorrentSearchCache
}

func NewTorrentSearchService(repo *storage.Repository, msgSender messaging.MessageSender, stateMgr *StateManager) *TorrentSearchService {
	return &TorrentSearchService{
		repo:               repo,
		msgSender:          msgSender,
		stateMgr:           stateMgr,
		torrentSearchCache: make(map[int64]*TorrentSearchCache),
	}
}

func (tss *TorrentSearchService) StartTorrentSearchDialog(chatId int64) {
	tss.stateMgr.SetUserState(chatId, string(dialogstate.StateSearchTorrent))
	text := ui.Msg(ui.MsgSearchStartPromptText)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			ui.ButtonWithData(ui.Cancel, "main_menu"),
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

func (tss *TorrentSearchService) SearchTorrents(chatId int64, query string) {
	ctx := context.Background()

	clients, err := tss.repo.GetAllClients(ctx, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиентов для поиска: %v", err)
		text := ui.Msg(ui.MsgSearchClientsListError)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := tss.stateMgr.GetMenuMessage(chatId)
		_, _ = tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)

		return
	}

	if len(clients) == 0 {
		text := ui.Msg(ui.MsgSearchNoClientsText)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.AddClient),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := tss.stateMgr.GetMenuMessage(chatId)
		_, _ = tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)

		return
	}

	query = strings.TrimSpace(query)
	queryUpper := strings.ToUpper(query)

	var searchResults []TorrentSearchResult

	for _, client := range clients {
		qbClient, qbErr := qbit.New(ctx, client)
		if qbErr != nil {
			logger.Warn("Ошибка при подключении к клиенту %s для поиска: %v", client.Name, qbErr)

			continue
		}

		torrents, torrentsErr := qbClient.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{Filter: qbittorrent.TorrentFilterAll})
		if torrentsErr != nil {
			logger.Warn("Ошибка при получении торрентов от клиента %s: %v", client.Name, torrentsErr)

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

func (tss *TorrentSearchService) ShowSearchResultsPage(chatId int64, page int) {
	cache, exists := tss.torrentSearchCache[chatId]
	if !exists || cache == nil {
		logger.Warn("Пользователь %d запросил страницу %d, но кэш результатов поиска отсутствует", chatId, page)
		text := ui.Msg(ui.MsgSearchResultsStaleText)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.NewSearch),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		messageID := tss.stateMgr.GetMenuMessage(chatId)
		_, _ = tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)

		return
	}

	tss.showSearchResults(chatId, cache.Query, cache.Results, page)
}

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

func (tss *TorrentSearchService) ClearSearchCache(chatId int64) {
	delete(tss.torrentSearchCache, chatId)
}

func (tss *TorrentSearchService) showSearchResults(chatId int64, query string, results []TorrentSearchResult, page int) {
	messageID := tss.stateMgr.GetMenuMessage(chatId)

	if len(results) == 0 {
		text := ui.Msgf(ui.MsgSearchNoResultsFmt, query)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.NewSearch),
			),
			tgbotapi.NewInlineKeyboardRow(
				ui.Button(ui.MainMenu),
			),
		)
		_, _ = tss.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)

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
		text.WriteString(ui.Msgf(ui.MsgSearchResultsHeaderWithQueryFmt, query))
	} else {
		text.WriteString(ui.Msg(ui.MsgSearchResultsHeader))
	}
	text.WriteString(ui.Msgf(ui.MsgSearchResultsFoundCountFmt, len(results)))

	var rows [][]tgbotapi.InlineKeyboardButton
	for i, result := range pageResults {
		displayName := result.Name
		if len(displayName) > 50 {
			displayName = displayName[:47] + "..."
		}

		globalIdx := startIdx + i
		text.WriteString(ui.Msgf(ui.MsgSearchResultsItemLineFmt, globalIdx+1, displayName))
		text.WriteString(ui.Msgf(ui.MsgSearchResultsItemHashFmt, result.Hash))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			ui.Data(
				ui.Msgf(ui.MsgSearchResultsButtonItemFmt, globalIdx+1, truncatePath(displayName, 999)),
				fmt.Sprintf("search_torrent_select_%d", globalIdx),
			),
		))
	}

	if totalPages > 1 {
		var navButtons []tgbotapi.InlineKeyboardButton
		if page > 0 {
			navButtons = append(navButtons, ui.ButtonWithData(ui.PrevPage, fmt.Sprintf("page_search_%d", page-1)))
		}
		navButtons = append(navButtons, ui.Data(fmt.Sprintf("%d/%d", page+1, totalPages), "page_info"))
		if page < totalPages-1 {
			navButtons = append(navButtons, ui.ButtonWithData(ui.NextPage, fmt.Sprintf("page_search_%d", page+1)))
		}
		if len(navButtons) > 0 {
			rows = append(rows, navButtons)
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		ui.Button(ui.NewSearch),
		ui.Button(ui.MainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	_, _ = tss.msgSender.SendOrEdit(chatId, messageID, text.String(), &keyboard)
}
