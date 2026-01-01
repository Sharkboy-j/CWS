package telegram

import (
	"bytes"
	"context"
	"cws/config"
	"cws/database"
	"cws/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot          *tgbotapi.BotAPI
	chatId       int64
	token        string
	repo         *database.Repository
	stateMgr     *StateManager
	msgSender    *MessageSender
	cmdHdlr      *CommandHandler
	callbackHdlr *CallbackHandler
	dialogHdlr   *DialogHandler
	clientHdlr   *ClientHandler
	autoChecker  *AutoChecker
}

func NewBotService(token string, chatId int64, repo *database.Repository, cfg *config.Config) (*BotService, error) {
	bot, err := InitBot(token)
	if err != nil {
		return nil, err
	}

	stateMgr, err := NewStateManager(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	msgSender := NewMessageSender(bot)

	service := &BotService{
		bot:       bot,
		chatId:    chatId,
		token:     token,
		repo:      repo,
		stateMgr:  stateMgr,
		msgSender: msgSender,
	}

	service.cmdHdlr = NewCommandHandler(bot, repo, stateMgr, msgSender)
	service.clientHdlr = NewClientHandler(repo, msgSender, stateMgr, cfg)
	service.dialogHdlr = NewDialogHandler(repo, msgSender, stateMgr, service.clientHdlr)
	service.callbackHdlr = NewCallbackHandler(bot, stateMgr, msgSender, service.dialogHdlr, service.clientHdlr)

	service.clientHdlr.SetCommandHandler(service.cmdHdlr)
	service.dialogHdlr.SetCommandHandler(service.cmdHdlr)
	service.callbackHdlr.SetCommandHandler(service.cmdHdlr)

	if !cfg.ManualCheckOnly && cfg.DurationSeconds > 0 {
		service.autoChecker = NewAutoChecker(service.clientHdlr, cfg.DurationSeconds)
	}

	err = service.setupCommands()
	if err != nil {
		logger.Warn("Не удалось установить команды бота: %v", err)
	}

	err = service.setupMenuButton()
	if err != nil {
		logger.Warn("Не удалось установить Menu Button: %v", err)
	}

	logger.Debugf("Бот инициализирован для пользователя %d", chatId)

	return service, nil
}

func (bs *BotService) StartAutoChecker(ctx context.Context) {
	if bs.autoChecker != nil {
		go bs.autoChecker.Start(ctx)
	}
}

func (bs *BotService) setupCommands() error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "menu",
			Description: "Главное меню",
		},
		{
			Command:     "check",
			Description: "Проверить статус",
		},
		{
			Command:     "clients",
			Description: "Управление клиентами qBittorrent",
		},
	}

	setCommands := tgbotapi.NewSetMyCommands(commands...)
	_, err := bs.bot.Request(setCommands)
	if err != nil {
		return err
	}

	logger.Info("Команды бота установлены успешно")
	return nil
}

func (bs *BotService) setupMenuButton() error {
	menuButton := map[string]interface{}{
		"type": "commands",
	}

	requestBody := map[string]interface{}{
		"chat_id":     bs.chatId,
		"menu_button": menuButton,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/setChatMenuButton", bs.token)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	logger.Info("Menu Button установлен успешно")
	return nil
}

func (bs *BotService) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bs.bot.GetUpdatesChan(u)
	logger.Info("Бот запущен и ожидает обновлений...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Получен сигнал остановки, завершаем работу бота...")
			bs.bot.StopReceivingUpdates()
			return nil
		case update := <-updates:
			logger.Debug("Получено обновление: UpdateID=%d", update.UpdateID)
			bs.handleUpdate(update)
		}
	}
}

func (bs *BotService) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		chatId := update.CallbackQuery.Message.Chat.ID
		username := update.CallbackQuery.From.UserName
		logger.Debugf("Пользователь %d (@%s) нажал на кнопку: %s", chatId, username, update.CallbackQuery.Data)
		bs.callbackHdlr.HandleCallbackQuery(update.CallbackQuery)
		return
	}

	if update.Message != nil && update.Message.IsCommand() {
		chatId := update.Message.Chat.ID
		username := update.Message.From.UserName
		command := update.Message.Command()
		messageID := update.Message.MessageID
		logger.Debugf("Пользователь %d (@%s) выполнил команду: /%s", chatId, username, command)

		bs.cmdHdlr.HandleCommand(update.Message)

		bs.msgSender.DeleteMessage(chatId, messageID)
		return
	}

	if update.Message != nil {
		chatId := update.Message.Chat.ID
		username := update.Message.From.UserName
		logger.Debugf("Пользователь %d (@%s) отправил сообщение: %s", chatId, username, update.Message.Text)

		// Обрабатываем локацию для определения часового пояса
		if update.Message.Location != nil {
			bs.handleUserLocation(chatId, update.Message.Location)
		}

		// Обрабатываем торрент файл
		if update.Message.Document != nil {
			bs.handleTorrentFile(chatId, update.Message)
			return
		}

		bs.dialogHdlr.HandleMessage(update.Message)
	}
}

// handleUserLocation обрабатывает локацию пользователя для определения часового пояса
func (bs *BotService) handleUserLocation(chatId int64, location *tgbotapi.Location) {
	ctx := context.Background()
	logger.Debugf("Пользователь %d отправил локацию: lat=%.6f, lon=%.6f", chatId, location.Latitude, location.Longitude)

	// Определяем часовой пояс по координатам
	// Используем простой подход: определяем часовой пояс по долготе
	// Более точный способ - использовать API для определения часового пояса
	timezone := determineTimezoneByCoordinates(location.Latitude, location.Longitude)

	// Сохраняем часовой пояс пользователя
	err := bs.repo.SetUserTimezone(ctx, chatId, timezone)
	if err != nil {
		logger.Error("Ошибка при сохранении часового пояса для пользователя %d: %v", chatId, err)
		return
	}

	logger.Info("Часовой пояс пользователя %d установлен: %s", chatId, timezone)
	msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("✅ Часовой пояс установлен: %s", timezone))
	bs.msgSender.Send(msg)
}

// determineTimezoneByCoordinates определяет часовой пояс по координатам
// Это упрощенная версия, для более точного определения можно использовать API
// Возвращает IANA timezone name или UTC по умолчанию
func determineTimezoneByCoordinates(lat, lon float64) string {
	// Упрощенный подход: определяем часовой пояс по долготе
	// Каждые 15 градусов долготы = 1 час разницы от UTC
	offset := int(lon / 15.0)

	// Ограничиваем диапазон от -12 до +14
	if offset < -12 {
		offset = -12
	}
	if offset > 14 {
		offset = 14
	}

	// Маппинг смещения на популярные IANA timezone
	// Это упрощенный подход, для точности лучше использовать API
	timezoneMap := map[int]string{
		-12: "Etc/GMT+12",
		-11: "Pacific/Midway",
		-10: "Pacific/Honolulu",
		-9:  "America/Anchorage",
		-8:  "America/Los_Angeles",
		-7:  "America/Denver",
		-6:  "America/Chicago",
		-5:  "America/New_York",
		-4:  "America/Caracas",
		-3:  "America/Sao_Paulo",
		-2:  "Atlantic/South_Georgia",
		-1:  "Atlantic/Azores",
		0:   "UTC",
		1:   "Europe/Paris",
		2:   "Europe/Berlin",
		3:   "Europe/Moscow",
		4:   "Asia/Dubai",
		5:   "Asia/Karachi",
		6:   "Asia/Dhaka",
		7:   "Asia/Bangkok",
		8:   "Asia/Shanghai",
		9:   "Asia/Tokyo",
		10:  "Australia/Sydney",
		11:  "Pacific/Guadalcanal",
		12:  "Pacific/Auckland",
		13:  "Pacific/Tongatapu",
		14:  "Pacific/Kiritimati",
	}

	if tz, ok := timezoneMap[offset]; ok {
		return tz
	}

	// Если смещение не найдено, используем UTC
	return "UTC"
}

// handleTorrentFile обрабатывает получение торрент файла от пользователя
func (bs *BotService) handleTorrentFile(chatId int64, message *tgbotapi.Message) {
	ctx := context.Background()
	state, exists := bs.stateMgr.GetUserState(chatId)
	if !exists || !strings.HasPrefix(state, "add_torrent_wait_file_") {
		logger.Debug("Пользователь %d отправил файл, но не в процессе добавления торрента", chatId)
		return
	}

	// Извлекаем ID клиента из состояния (формат: add_torrent_wait_file_{clientID})
	prefix := "add_torrent_wait_file_"
	clientIDStr := strings.TrimPrefix(state, prefix)
	if clientIDStr == state {
		logger.Warn("Неверный формат состояния для добавления торрента: %s", state)
		return
	}
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Error("Ошибка при парсинге ID клиента из состояния: %v", err)
		return
	}

	document := message.Document
	if document == nil {
		logger.Warn("Документ не найден в сообщении")
		return
	}

	// Проверяем, что это торрент файл
	if !strings.HasSuffix(strings.ToLower(document.FileName), ".torrent") {
		msg := tgbotapi.NewMessage(chatId, "❌ Пожалуйста, отправьте файл с расширением .torrent")
		bs.msgSender.Send(msg)
		return
	}

	logger.Debugf("Пользователь %d отправил торрент файл %s для клиента %d", chatId, document.FileName, clientID)

	// Сразу обновляем сообщение, показывая что файл получен и обрабатывается
	text := fmt.Sprintf("📥 *Обработка торрент файла*\n\n📎 Файл: `%s`\n\n⏳ Обрабатываю...", document.FileName)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_add_torrent"),
		),
	)
	messageID := bs.stateMgr.GetMenuMessage(chatId)
	_, err = bs.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Warn("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
	}

	// Скачиваем файл
	fileURL, err := bs.bot.GetFileDirectURL(document.FileID)
	if err != nil {
		logger.Error("Ошибка при получении URL файла: %v", err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при получении файла")
		bs.msgSender.Send(msg)
		return
	}

	// Скачиваем содержимое файла
	resp, err := http.Get(fileURL)
	if err != nil {
		logger.Error("Ошибка при скачивании файла: %v", err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при скачивании файла")
		bs.msgSender.Send(msg)
		return
	}
	defer resp.Body.Close()

	torrentData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Ошибка при чтении файла: %v", err)
		msg := tgbotapi.NewMessage(chatId, "❌ Ошибка при чтении файла")
		bs.msgSender.Send(msg)
		return
	}

	// Сохраняем файл во временное состояние и переходим к выбору пути
	bs.clientHdlr.HandleTorrentFileReceived(ctx, chatId, clientID, torrentData, document.FileName)
	
	// Удаляем сообщение с файлом после обработки
	bs.msgSender.DeleteMessage(chatId, message.MessageID)
}
