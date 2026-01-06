package bot

import (
	"fmt"

	"cws/internal/bot/ui"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func paginateRange(total, pageSize, page int) (pageOut, totalPages, start, end int) {
	if pageSize <= 0 {
		pageSize = 1
	}
	if total < 0 {
		total = 0
	}

	totalPages = (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start = page * pageSize
	end = start + pageSize
	if end > total {
		end = total
	}

	return page, totalPages, start, end
}

func truncateButtonLabel(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}

	return s[:max-3] + "..."
}

func buildMissingTorrentRowButtons(info missingTorrentInfo) []tgbotapi.InlineKeyboardButton {
	btnText := truncateButtonLabel(info.name, 60)
	row := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonURL(btnText, info.url),
	}
	if info.hash != "" {
		row = append(row, ui.Data(ui.IconChart, fmt.Sprintf("monitor_from_missing_%s", info.hash)))
	}

	return row
}
