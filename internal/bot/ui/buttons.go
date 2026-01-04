package ui

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Locale string

const (
	LocaleRU Locale = "ru"
	LocaleEN Locale = "en"
)

const DefaultLocale = LocaleRU

type ButtonID string

const (
	MainMenu ButtonID = "main_menu"
	Cancel   ButtonID = "cancel"
	Back     ButtonID = "back"

	AddClient         ButtonID = "add_client"
	RepeatCheck       ButtonID = "repeat_check"
	NewSearch         ButtonID = "new_search"
	PrevPage          ButtonID = "prev_page"
	NextPage          ButtonID = "next_page"
	Yes               ButtonID = "yes"
	No                ButtonID = "no"
	CheckTorrents     ButtonID = "check_torrents"
	QuickActionsMenu  ButtonID = "quick_actions_menu"
	ClientsMenu       ButtonID = "clients_menu"
	AddTorrentFile    ButtonID = "add_torrent_file"
	SearchTorrent     ButtonID = "search_torrent"
	MonitorTorrent    ButtonID = "monitor_torrent"
	PauseAllTorrents  ButtonID = "pause_all_torrents"
	ResumeAllTorrents ButtonID = "resume_all_torrents"
	SpeedLimitMenu    ButtonID = "speed_limit_menu"

	Speed10       ButtonID = "speed_10"
	Speed100      ButtonID = "speed_100"
	Speed500      ButtonID = "speed_500"
	Speed1000     ButtonID = "speed_1000"
	Speed2000     ButtonID = "speed_2000"
	Speed5000     ButtonID = "speed_5000"
	Speed10000    ButtonID = "speed_10000"
	Speed50000    ButtonID = "speed_50000"
	SpeedCustom   ButtonID = "speed_custom"
	SpeedRemove   ButtonID = "speed_remove"
	Edit          ButtonID = "edit"
	Delete        ButtonID = "delete"
	BackToList    ButtonID = "back_to_list"
	ConfirmDelete ButtonID = "confirm_delete"
	KeepExisting  ButtonID = "keep_existing"

	ManualHashInput ButtonID = "manual_hash_input"
	ManualPathInput ButtonID = "manual_path_input"
	SkipHashYes     ButtonID = "skip_hash_yes"
	SkipHashNo      ButtonID = "skip_hash_no"
	DeleteFilesYes  ButtonID = "delete_files_yes"
	DeleteFilesNo   ButtonID = "delete_files_no"
)

type buttonResource struct {
	text map[Locale]string
	data string
}

var buttonResources = map[ButtonID]buttonResource{
	MainMenu: {
		data: "main_menu",
		text: map[Locale]string{
			LocaleRU: "🏠 В главное меню",
			LocaleEN: "🏠 Main menu",
		},
	},
	Cancel: {
		text: map[Locale]string{
			LocaleRU: "❌ Отмена",
			LocaleEN: "❌ Cancel",
		},
	},
	Back: {
		text: map[Locale]string{
			LocaleRU: "🔙 Назад",
			LocaleEN: "🔙 Back",
		},
	},
	AddClient: {
		data: "add_client",
		text: map[Locale]string{
			LocaleRU: "➕ Добавить клиента",
			LocaleEN: "➕ Add client",
		},
	},
	RepeatCheck: {
		data: "check_torrents",
		text: map[Locale]string{
			LocaleRU: "🔄 Повторить проверку",
			LocaleEN: "🔄 Re-run check",
		},
	},
	NewSearch: {
		data: "search_torrent",
		text: map[Locale]string{
			LocaleRU: "🔎 Новый поиск",
			LocaleEN: "🔎 New search",
		},
	},
	PrevPage: {
		text: map[Locale]string{
			LocaleRU: "◀️ Назад",
			LocaleEN: "◀️ Back",
		},
	},
	NextPage: {
		text: map[Locale]string{
			LocaleRU: "Вперёд ▶️",
			LocaleEN: "Next ▶️",
		},
	},
	Yes: {
		text: map[Locale]string{
			LocaleRU: "✅ Да",
			LocaleEN: "✅ Yes",
		},
	},
	No: {
		text: map[Locale]string{
			LocaleRU: "❌ Нет",
			LocaleEN: "❌ No",
		},
	},
	CheckTorrents: {
		data: "check_torrents",
		text: map[Locale]string{
			LocaleRU: "🔍 Првоерить обновления",
			LocaleEN: "🔍 Check updates",
		},
	},
	QuickActionsMenu: {
		data: "quick_actions",
		text: map[Locale]string{
			LocaleRU: "⚡ Быстрые действия",
			LocaleEN: "⚡ Quick actions",
		},
	},
	ClientsMenu: {
		data: "clients",
		text: map[Locale]string{
			LocaleRU: "📋 Клиенты",
			LocaleEN: "📋 Clients",
		},
	},
	AddTorrentFile: {
		data: "add_torrent_file",
		text: map[Locale]string{
			LocaleRU: "📥 Добавить торрент файл",
			LocaleEN: "📥 Add torrent file",
		},
	},
	SearchTorrent: {
		data: "search_torrent",
		text: map[Locale]string{
			LocaleRU: "🔎 Поиск торрента",
			LocaleEN: "🔎 Search torrent",
		},
	},
	MonitorTorrent: {
		data: "monitor_torrent",
		text: map[Locale]string{
			LocaleRU: "📊 Мониторинг торрента",
			LocaleEN: "📊 Torrent monitoring",
		},
	},
	PauseAllTorrents: {
		data: "quick_action_pause_all",
		text: map[Locale]string{
			LocaleRU: "⏸ Остановить все раздачи",
			LocaleEN: "⏸ Pause all torrents",
		},
	},
	ResumeAllTorrents: {
		data: "quick_action_resume_all",
		text: map[Locale]string{
			LocaleRU: "▶ Запустить все раздачи",
			LocaleEN: "▶ Resume all torrents",
		},
	},
	SpeedLimitMenu: {
		data: "quick_action_limit_speed_menu",
		text: map[Locale]string{
			LocaleRU: "🚦 Ограничение скорости",
			LocaleEN: "🚦 Speed limit",
		},
	},
	Speed10:         {data: "quick_action_limit_speed_10", text: map[Locale]string{LocaleRU: "0.10 МБ/с", LocaleEN: "0.10 MB/s"}},
	Speed100:        {data: "quick_action_limit_speed_100", text: map[Locale]string{LocaleRU: "1.00 МБ/с", LocaleEN: "1.00 MB/s"}},
	Speed500:        {data: "quick_action_limit_speed_500", text: map[Locale]string{LocaleRU: "5.00 МБ/с", LocaleEN: "5.00 MB/s"}},
	Speed1000:       {data: "quick_action_limit_speed_1000", text: map[Locale]string{LocaleRU: "10.00 МБ/с", LocaleEN: "10.00 MB/s"}},
	Speed2000:       {data: "quick_action_limit_speed_2000", text: map[Locale]string{LocaleRU: "20.00 МБ/с", LocaleEN: "20.00 MB/s"}},
	Speed5000:       {data: "quick_action_limit_speed_5000", text: map[Locale]string{LocaleRU: "50.00 МБ/с", LocaleEN: "50.00 MB/s"}},
	Speed10000:      {data: "quick_action_limit_speed_10000", text: map[Locale]string{LocaleRU: "100.00 МБ/с", LocaleEN: "100.00 MB/s"}},
	Speed50000:      {data: "quick_action_limit_speed_50000", text: map[Locale]string{LocaleRU: "500.00 МБ/с", LocaleEN: "500.00 MB/s"}},
	SpeedCustom:     {data: "quick_action_limit_speed_custom", text: map[Locale]string{LocaleRU: "✏ Ввести вручную", LocaleEN: "✏ Enter manually"}},
	SpeedRemove:     {data: "quick_action_remove_speed_limits", text: map[Locale]string{LocaleRU: "🚫 Убрать все ограничения", LocaleEN: "🚫 Remove limits"}},
	Edit:            {text: map[Locale]string{LocaleRU: "✏️ Изменить", LocaleEN: "✏️ Edit"}},
	Delete:          {text: map[Locale]string{LocaleRU: "🗑 Удалить", LocaleEN: "🗑 Delete"}},
	BackToList:      {data: "clients", text: map[Locale]string{LocaleRU: "🔙 Назад к списку", LocaleEN: "🔙 Back to list"}},
	ConfirmDelete:   {text: map[Locale]string{LocaleRU: "✅ Да, удалить", LocaleEN: "✅ Yes, delete"}},
	KeepExisting:    {text: map[Locale]string{LocaleRU: "❌ Нет, оставить", LocaleEN: "❌ No, keep"}},
	ManualHashInput: {text: map[Locale]string{LocaleRU: "✏️ Ввести хеш вручную", LocaleEN: "✏️ Enter hash manually"}},
	ManualPathInput: {text: map[Locale]string{LocaleRU: "✏️ Ввести путь вручную", LocaleEN: "✏️ Enter path manually"}},
	SkipHashYes:     {text: map[Locale]string{LocaleRU: "✅ Да, пропустить", LocaleEN: "✅ Yes, skip"}},
	SkipHashNo:      {text: map[Locale]string{LocaleRU: "❌ Нет, проверить", LocaleEN: "❌ No, check"}},
	DeleteFilesYes:  {text: map[Locale]string{LocaleRU: "✅ Да, удалить файлы", LocaleEN: "✅ Yes, delete files"}},
	DeleteFilesNo:   {text: map[Locale]string{LocaleRU: "❌ Нет, только торрент", LocaleEN: "❌ No, torrent only"}},
}

func Text(id ButtonID) string {
	return TextForLocale(id, DefaultLocale)
}

func TextForLocale(id ButtonID, locale Locale) string {
	res, ok := buttonResources[id]
	if !ok {
		return string(id)
	}
	if txt, hasText := res.text[locale]; hasText && txt != "" {
		return txt
	}
	if txt, hasText := res.text[DefaultLocale]; hasText && txt != "" {
		return txt
	}

	return string(id)
}

func Button(id ButtonID) tgbotapi.InlineKeyboardButton {
	return ButtonForLocale(id, DefaultLocale)
}

func ButtonForLocale(id ButtonID, locale Locale) tgbotapi.InlineKeyboardButton {
	res, ok := buttonResources[id]
	if !ok {
		return tgbotapi.NewInlineKeyboardButtonData(string(id), string(id))
	}
	data := res.data
	if data == "" {
		data = string(id)
	}

	return tgbotapi.NewInlineKeyboardButtonData(TextForLocale(id, locale), data)
}

func ButtonWithData(id ButtonID, callbackData string) tgbotapi.InlineKeyboardButton {
	return ButtonWithDataForLocale(id, callbackData, DefaultLocale)
}

func ButtonWithDataForLocale(id ButtonID, callbackData string, locale Locale) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(TextForLocale(id, locale), callbackData)
}

func Data(text, callbackData string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, callbackData)
}
