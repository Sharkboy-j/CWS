package telegram

import (
	"context"
	"cws/logger"
	"cws/store"
)

type StateManager struct {
	repo          *store.Repository
	userState     map[int64]string
	dialogMessage map[int64]int
	menuMessage   map[int64]int
}

func NewStateManager(repo *store.Repository) (*StateManager, error) {
	sm := &StateManager{
		repo:          repo,
		userState:     make(map[int64]string),
		dialogMessage: make(map[int64]int),
		menuMessage:   make(map[int64]int),
	}

	ctx := context.Background()

	states, err := repo.GetAllUserStates(ctx)
	if err != nil {
		logger.Warn("Не удалось загрузить состояния пользователей: %v", err)
	} else {
		sm.userState = states
		logger.Debugf("Загружено %d состояний пользователей из БД", len(states))
	}

	menuMessageIDs, err := repo.GetAllMenuMessageIDs(ctx)
	if err != nil {
		logger.Warn("Не удалось загрузить menu_message_id пользователей: %v", err)
	} else {
		sm.menuMessage = menuMessageIDs
		logger.Debugf("Загружено %d menu_message_id пользователей из БД", len(menuMessageIDs))
	}

	return sm, nil
}

func (sm *StateManager) SetUserState(chatId int64, state string) {
	ctx := context.Background()
	sm.userState[chatId] = state
	if err := sm.repo.SetUserState(ctx, chatId, state); err != nil {
		logger.Error("Ошибка при сохранении состояния пользователя %d: %v", chatId, err)
	}
}

func (sm *StateManager) GetUserState(chatId int64) (string, bool) {
	state, exists := sm.userState[chatId]

	return state, exists
}

func (sm *StateManager) DeleteUserState(chatId int64) {
	ctx := context.Background()
	delete(sm.userState, chatId)
	delete(sm.dialogMessage, chatId) // Удаляем и message_id
	if err := sm.repo.DeleteUserState(ctx, chatId); err != nil {
		logger.Error("Ошибка при удалении состояния пользователя %d: %v", chatId, err)
	}
}

func (sm *StateManager) SetDialogMessage(chatId int64, messageID int) {
	sm.dialogMessage[chatId] = messageID
}

func (sm *StateManager) GetDialogMessage(chatId int64) int {

	return sm.dialogMessage[chatId]
}

func (sm *StateManager) SetMenuMessage(chatId int64, messageID int) {
	ctx := context.Background()
	sm.menuMessage[chatId] = messageID
	if err := sm.repo.SetMenuMessageID(ctx, chatId, messageID); err != nil {
		logger.Error("Ошибка при сохранении menu_message_id пользователя %d: %v", chatId, err)
	}
}

func (sm *StateManager) GetMenuMessage(chatId int64) int {

	return sm.menuMessage[chatId]
}
