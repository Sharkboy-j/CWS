package telegram

import (
	"context"
	"cws/logger"
	"cws/store"
	"cws/telegram/messaging"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DialogHandler struct {
	repo       *store.Repository
	msgSender  messaging.MessageSender
	stateMgr   *StateManager
	clientHdlr *ClientHandler
	cmdHdlr    *CommandHandler
}

func NewDialogHandler(repo *store.Repository, msgSender messaging.MessageSender, stateMgr *StateManager, clientHdlr *ClientHandler) *DialogHandler {
	return &DialogHandler{
		repo:       repo,
		msgSender:  msgSender,
		stateMgr:   stateMgr,
		clientHdlr: clientHdlr,
	}
}

func (dh *DialogHandler) SetCommandHandler(cmdHdlr *CommandHandler) {
	dh.cmdHdlr = cmdHdlr
}

func (dh *DialogHandler) StartAddClientDialog(chatId int64) {
	dh.stateMgr.SetUserState(chatId, "add_client_name")
	text := "➕ *Добавление нового клиента*\n\n📝 Введите имя клиента:"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_add_client"),
		),
	)
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, text, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) HandleMessage(message *tgbotapi.Message) {
	chatId := message.Chat.ID
	messageID := message.MessageID
	state, exists := dh.stateMgr.GetUserState(chatId)
	if !exists {
		logger.Debug("Пользователь %d отправил сообщение, но не в процессе диалога", chatId)

		return
	}

	text := message.Text
	if text == "" {
		logger.Debug("Пользователь %d отправил пустое сообщение", chatId)

		return
	}

	logger.Debug("Обработка сообщения от пользователя %d в состоянии: %s, текст: %s", chatId, state, text)

	const separator = "|||"

	defer dh.msgSender.DeleteMessage(chatId, messageID)

	switch {
	case state == "add_client_name":
		dh.handleAddClientName(chatId, text, separator)
	case strings.HasPrefix(state, "add_client_host"):
		dh.handleAddClientHost(chatId, text, state, separator)
	case strings.HasPrefix(state, "add_client_port"):
		dh.handleAddClientPort(chatId, text, state, separator)
	case strings.HasPrefix(state, "add_client_username"):
		dh.handleAddClientUsername(chatId, text, state, separator)
	case strings.HasPrefix(state, "add_client_password"):
		dh.handleAddClientPassword(chatId, text, state, separator)
	case strings.HasPrefix(state, "edit_client_name"):
		dh.handleEditClientName(chatId, text, state, separator)
	case strings.HasPrefix(state, "edit_client_host"):
		dh.handleEditClientHost(chatId, text, state, separator)
	case strings.HasPrefix(state, "edit_client_port"):
		dh.handleEditClientPort(chatId, text, state, separator)
	case strings.HasPrefix(state, "edit_client_username"):
		dh.handleEditClientUsername(chatId, text, state, separator)
	case strings.HasPrefix(state, "edit_client_password"):
		dh.handleEditClientPassword(chatId, text, state, separator)
	case strings.HasPrefix(state, "add_torrent_custom_path_"):
		dh.handleAddTorrentCustomPath(chatId, text, state)
	case strings.HasPrefix(state, "monitor_torrent_hash_"):
		dh.handleMonitorTorrentHash(chatId, text, state)
	case state == "search_torrent_query":
		dh.handleTorrentSearchQuery(chatId, text)
	default:
		logger.Warn("Неизвестное состояние для пользователя %d: %s, текст: %s", chatId, state, text)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неизвестное состояние. Начните операцию заново.", nil)
	}
}

func (dh *DialogHandler) handleAddClientName(chatId int64, text, separator string) {
	logger.Debugf("Пользователь %d ввел имя клиента: %s", chatId, text)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("add_client_host%s%s", separator, text))
	messageText := "🌐 Введите host (например: 192.168.1.100):"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) handleAddClientHost(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел host: %s, состояние: %s", chatId, text, state)
	parts := strings.SplitN(state, separator, 3)
	if len(parts) < 2 {
		logger.Warn("Неверный формат состояния add_client_host для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните добавление заново.", nil)

		return
	}
	clientName := parts[1]
	logger.Debug("Извлечено имя клиента: %s", clientName)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("add_client_port%s%s%s%s", separator, clientName, separator, text))
	messageText := "🔌 Введите port (например: 8080):"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	logger.Debug("Обновление сообщения %d для пользователя %d", messageID, chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)
		_, sendErr := dh.msgSender.SendOrEdit(chatId, 0, messageText, nil)
		if sendErr != nil {
			logger.Error("Не удалось отправить новое сообщение пользователю %d: %v", chatId, sendErr)
		}

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
	logger.Debug("Сообщение успешно обновлено для пользователя %d, новый messageID: %d", chatId, newMessageID)
}

func (dh *DialogHandler) handleAddClientPort(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел port: %s, состояние: %s", chatId, text, state)
	parts := strings.SplitN(state, separator, 4)
	if len(parts) < 3 {
		logger.Warn("Неверный формат состояния add_client_port для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните добавление заново.", nil)

		return
	}
	clientName := parts[1]
	host := parts[2]
	port, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		messageText := "⚠️ Ошибка: порт должен быть числом. Попробуйте снова:"
		messageID := dh.stateMgr.GetDialogMessage(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)

		return
	}
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("add_client_username%s%s%s%s%s%d", separator, clientName, separator, host, separator, port))
	messageText := "👤 Введите username:"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) handleAddClientUsername(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел username: %s, состояние: %s", chatId, text, state)
	parts := strings.SplitN(state, separator, 5)
	if len(parts) < 4 {
		logger.Warn("Неверный формат состояния add_client_username для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните добавление заново.", nil)

		return
	}
	clientName := parts[1]
	host := parts[2]
	portStr := parts[3]
	port, _ := strconv.ParseInt(portStr, 10, 32)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("add_client_password%s%s%s%s%s%d%s%s", separator, clientName, separator, host, separator, port, separator, text))
	messageText := "🔑 Введите password:"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) handleAddClientPassword(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел password, состояние: %s", chatId, state)
	parts := strings.SplitN(state, separator, 6)
	if len(parts) < 5 {
		logger.Warn("Неверный формат состояния add_client_password для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните добавление заново.", nil)

		return
	}
	clientName := parts[1]
	host := parts[2]
	portStr := parts[3]
	port, _ := strconv.ParseInt(portStr, 10, 32)
	username := parts[4]
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("add_client_ssl%s%s%s%s%s%d%s%s%s%s", separator, clientName, separator, host, separator, port, separator, username, separator, text))
	messageText := "🔒 Использовать SSL?"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да", "set_ssl_true"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", "set_ssl_false"),
		),
	)
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) FinishAddClient(chatId int64, ssl bool) {
	const separator = "|||"

	state, exists := dh.stateMgr.GetUserState(chatId)
	if !exists || !strings.HasPrefix(state, "add_client_ssl") {
		logger.Warn("Состояние не найдено или неверный префикс для пользователя %d: exists=%v, state=%s", chatId, exists, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: сессия истекла. Начните заново.", nil)

		return
	}

	logger.Debug("Обработка завершения добавления клиента для пользователя %d, состояние: %s", chatId, state)
	parts := strings.SplitN(state, separator, 7)
	if len(parts) < 6 {
		logger.Warn("Неверный формат состояния add_client_ssl для пользователя %d: %s (частей: %d, ожидается 6)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверные данные. Начните заново.", nil)

		return
	}

	clientName := parts[1]
	host := parts[2]
	portStr := parts[3]
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверный порт. Начните заново.", nil)

		return
	}
	username := parts[4]
	password := parts[5]

	ctx := context.Background()
	client := &store.Client{
		UserID:   chatId,
		Name:     clientName,
		Host:     host,
		Port:     int32(port),
		Username: username,
		Password: password,
		SSL:      ssl,
	}

	createdClient, err := dh.repo.CreateClient(ctx, client)
	if err != nil {
		logger.Error("Ошибка при создании клиента для пользователя %d: %v", chatId, err)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка при создании клиента. Попробуйте снова.", nil)
		dh.stateMgr.DeleteUserState(chatId)

		return
	}

	logger.Debugf("Пользователь %d успешно создал клиента: ID=%d, Name=%s, Host=%s:%d",
		chatId, createdClient.ID, createdClient.Name, createdClient.Host, createdClient.Port)

	dialogMessageID := dh.stateMgr.GetDialogMessage(chatId)
	if dialogMessageID > 0 {
		dh.msgSender.DeleteMessage(chatId, dialogMessageID)
		dh.stateMgr.SetDialogMessage(chatId, 0)
	}

	dh.stateMgr.DeleteUserState(chatId)

	if dh.cmdHdlr != nil {
		dh.cmdHdlr.HandleClientsCommand(chatId)
	}
}

func (dh *DialogHandler) StartEditClientDialog(chatId int64, clientID int64) {
	ctx := context.Background()
	client, err := dh.repo.GetClientByID(ctx, clientID, chatId)
	if err != nil {
		logger.Error("Ошибка при получении клиента %d для редактирования пользователем %d: %v", clientID, chatId, err)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка при получении данных клиента", nil)

		return
	}

	if client == nil {
		logger.Warn("Пользователь %d попытался редактировать несуществующий клиент %d", chatId, clientID)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Клиент не найден или у вас нет доступа", nil)

		return
	}

	const separator = "|||"
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("edit_client_name%s%d%s%s", separator, clientID, separator, client.Name))
	messageText := fmt.Sprintf("✏️ *Редактирование клиента*\n\n📝 Текущее имя: `%s`\n\nВведите новое имя клиента (или отправьте текущее для сохранения):", client.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_edit_client"),
		),
	)
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при отправке/обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) handleEditClientName(chatId int64, text, state, separator string) {
	parts := strings.SplitN(state, separator, 3)
	if len(parts) < 3 {
		logger.Warn("Неверный формат состояния edit_client_name для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните редактирование заново.", nil)

		return
	}
	clientIDStr := parts[1]
	logger.Debugf("Пользователь %d ввел новое имя клиента: %s", chatId, text)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("edit_client_host%s%s%s%s", separator, clientIDStr, separator, text))
	messageText := "🌐 Введите host (например: 192.168.1.100):"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) handleEditClientHost(chatId int64, text, state, separator string) {
	parts := strings.SplitN(state, separator, 3)
	if len(parts) < 3 {
		logger.Warn("Неверный формат состояния edit_client_host для пользователя %d: %s", chatId, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните редактирование заново.", nil)

		return
	}
	clientIDStr := parts[1]
	clientName := parts[2]
	logger.Debugf("Пользователь %d ввел host: %s", chatId, text)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("edit_client_port%s%s%s%s%s%s", separator, clientIDStr, separator, clientName, separator, text))
	messageText := "🔌 Введите port (например: 8080):"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

// handleEditClientPort обрабатывает ввод нового порта клиента
func (dh *DialogHandler) handleEditClientPort(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел port: %s, состояние: %s", chatId, text, state)
	parts := strings.SplitN(state, separator, 4)
	if len(parts) < 4 {
		logger.Warn("Неверный формат состояния edit_client_port для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните редактирование заново.", nil)

		return
	}
	clientIDStr := parts[1]
	clientName := parts[2]
	host := parts[3]
	port, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		logger.Warn("Пользователь %d ввел неверный порт: %s", chatId, text)
		messageText := "⚠️ Ошибка: порт должен быть числом. Попробуйте снова:"
		messageID := dh.stateMgr.GetDialogMessage(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)

		return
	}
	logger.Debugf("Пользователь %d ввел port: %d", chatId, port)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("edit_client_username%s%s%s%s%s%s%s%d", separator, clientIDStr, separator, clientName, separator, host, separator, port))
	messageText := "👤 Введите username:"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

// handleEditClientUsername обрабатывает ввод нового username клиента
func (dh *DialogHandler) handleEditClientUsername(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел username: %s, состояние: %s", chatId, text, state)
	parts := strings.SplitN(state, separator, 5)
	if len(parts) < 5 {
		logger.Warn("Неверный формат состояния edit_client_username для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните редактирование заново.", nil)

		return
	}
	clientIDStr := parts[1]
	clientName := parts[2]
	host := parts[3]
	portStr := parts[4]
	port, _ := strconv.ParseInt(portStr, 10, 32)
	logger.Debugf("Пользователь %d ввел username: %s", chatId, text)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("edit_client_password%s%s%s%s%s%s%s%d%s%s", separator, clientIDStr, separator, clientName, separator, host, separator, port, separator, text))
	messageText := "🔑 Введите password:"
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, nil)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

// handleEditClientPassword обрабатывает ввод нового password клиента
func (dh *DialogHandler) handleEditClientPassword(chatId int64, text, state, separator string) {
	logger.Debugf("Пользователь %d ввел password, состояние: %s", chatId, state)
	parts := strings.SplitN(state, separator, 6)
	if len(parts) < 6 {
		logger.Warn("Неверный формат состояния edit_client_password для пользователя %d: %s (частей: %d)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните редактирование заново.", nil)

		return
	}
	clientIDStr := parts[1]
	clientName := parts[2]
	host := parts[3]
	portStr := parts[4]
	port, _ := strconv.ParseInt(portStr, 10, 32)
	username := parts[5]
	logger.Debugf("Пользователь %d ввел password", chatId)
	dh.stateMgr.SetUserState(chatId, fmt.Sprintf("edit_client_ssl%s%s%s%s%s%s%s%d%s%s%s%s", separator, clientIDStr, separator, clientName, separator, host, separator, port, separator, username, separator, text))
	messageText := "🔒 Использовать SSL?"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да", "set_edit_ssl_true"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", "set_edit_ssl_false"),
		),
	)
	messageID := dh.stateMgr.GetDialogMessage(chatId)
	newMessageID, err := dh.msgSender.SendOrEdit(chatId, messageID, messageText, &keyboard)
	if err != nil {
		logger.Error("Ошибка при обновлении сообщения для пользователя %d: %v", chatId, err)

		return
	}
	dh.stateMgr.SetDialogMessage(chatId, newMessageID)
}

func (dh *DialogHandler) FinishEditClient(chatId int64, ssl bool) {
	const separator = "|||"

	state, exists := dh.stateMgr.GetUserState(chatId)
	if !exists || !strings.HasPrefix(state, "edit_client_ssl") {
		logger.Warn("Состояние не найдено или неверный префикс для пользователя %d: exists=%v, state=%s", chatId, exists, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: сессия истекла. Начните заново.", nil)

		return
	}

	logger.Debug("Обработка завершения редактирования для пользователя %d, состояние: %s", chatId, state)
	parts := strings.SplitN(state, separator, 7)
	if len(parts) < 7 {
		logger.Warn("Неверный формат состояния edit_client_ssl для пользователя %d: %s (частей: %d, ожидается 7)", chatId, state, len(parts))
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверные данные. Начните заново.", nil)

		return
	}

	clientIDStr := parts[1]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверный ID клиента. Начните заново.", nil)

		return
	}

	clientName := parts[2]
	host := parts[3]
	portStr := parts[4]
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверный порт. Начните заново.", nil)

		return
	}
	username := parts[5]
	password := parts[6]

	ctx := context.Background()
	client := &store.Client{
		ID:       clientID,
		UserID:   chatId,
		Name:     clientName,
		Host:     host,
		Port:     int32(port),
		Username: username,
		Password: password,
		SSL:      ssl,
	}

	err = dh.repo.UpdateClient(ctx, client, chatId)
	if err != nil {
		logger.Error("Ошибка при обновлении клиента %d для пользователя %d: %v", clientID, chatId, err)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка при обновлении клиента. Попробуйте снова.", nil)
		dh.stateMgr.DeleteUserState(chatId)

		return
	}

	logger.Debugf("Пользователь %d успешно обновил клиента: ID=%d, Name=%s, Host=%s:%d",
		chatId, clientID, clientName, host, port)

	dialogMessageID := dh.stateMgr.GetDialogMessage(chatId)
	if dialogMessageID > 0 {
		dh.msgSender.DeleteMessage(chatId, dialogMessageID)
		dh.stateMgr.SetDialogMessage(chatId, 0)
	}

	menuMessageID := dh.stateMgr.GetMenuMessage(chatId)
	if menuMessageID > 0 {
		dh.msgSender.DeleteMessage(chatId, menuMessageID)
		dh.stateMgr.SetMenuMessage(chatId, 0)
	}

	dh.stateMgr.DeleteUserState(chatId)

	dh.clientHdlr.ShowClientDetails(chatId, clientID)
}

// handleAddTorrentCustomPath обрабатывает ввод пути сохранения вручную
func (dh *DialogHandler) handleAddTorrentCustomPath(chatId int64, text, state string) {
	logger.Debugf("Пользователь %d ввел путь сохранения: %s", chatId, text)

	parts := strings.Split(state, "_")
	if len(parts) < 5 {
		logger.Warn("Неверный формат состояния add_torrent_custom_path для пользователя %d: %s", chatId, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните заново.", nil)

		return
	}
	clientIDStr := parts[4]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Error("Ошибка при парсинге ID клиента: %v", err)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверный ID клиента.", nil)

		return
	}

	cache, exists := dh.clientHdlr.torrentFilesCache[chatId]
	if !exists || cache == nil || cache.ClientID != clientID {
		logger.Warn("Кэш торрент файла не найден для пользователя %d", chatId)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "❌ Ошибка: данные торрента не найдены. Начните заново.", nil)

		return
	}

	cache.SelectedPath = text
	dh.clientHdlr.torrentFilesCache[chatId] = cache

	dh.clientHdlr.ShowSkipHashCheckQuestion(chatId, clientID, text)
}

// handleMonitorTorrentHash обрабатывает ввод хеша торрента для мониторинга
func (dh *DialogHandler) handleMonitorTorrentHash(chatId int64, text, state string) {
	logger.Debugf("Пользователь %d ввел хеш торрента для мониторинга: %s", chatId, text)

	parts := strings.Split(state, "_")
	if len(parts) < 4 {
		logger.Warn("Неверный формат состояния monitor_torrent_hash для пользователя %d: %s", chatId, state)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверное состояние. Начните заново.", nil)

		return
	}
	clientIDStr := parts[3]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		logger.Warn("Неверный ID клиента в состоянии monitor_torrent_hash для пользователя %d: %s", chatId, clientIDStr)
		dh.stateMgr.DeleteUserState(chatId)
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "Ошибка: неверный ID клиента. Начните заново.", nil)

		return
	}

	hash := strings.TrimSpace(strings.ToUpper(text))

	if hash == "" {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "❌ Хеш не может быть пустым. Введите хеш торрента:", nil)

		return
	}

	if len(hash) != 40 {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "❌ Хеш должен содержать 40 символов. Введите правильный хеш:", nil)

		return
	}

	dh.stateMgr.DeleteUserState(chatId)

	ctx := context.Background()
	dh.clientHdlr.torrentMonitorSvc.StartTorrentMonitoring(ctx, chatId, clientID, hash)
}

func (dh *DialogHandler) handleTorrentSearchQuery(chatId int64, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		_, _ = dh.msgSender.SendOrEdit(chatId, 0, "❌ Поисковый запрос не может быть пустым. Введите хеш или название торрента:", nil)

		return
	}

	dh.stateMgr.DeleteUserState(chatId)

	dh.clientHdlr.torrentSearchSvc.SearchTorrents(chatId, query)
}
