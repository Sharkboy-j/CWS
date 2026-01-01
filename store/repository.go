package store

import (
	"context"
	"cws/logger"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

// NewRepository создает новый экземпляр репозитория
func NewRepository(host string, port int, user, password, dbname string) (*Repository, error) {
	logger.Info("Проверка существования БД: %s", dbname)
	if err := ensureDatabase(host, port, user, password, dbname); err != nil {
		logger.Error("Ошибка при создании БД %s: %v", dbname, err)

		return nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}
	logger.Info("БД %s готова к использованию", dbname)

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	logger.Debug("Подключение к БД: host=%s, port=%d, dbname=%s", host, port, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Ошибка при открытии соединения с БД: %v", err)

		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		logger.Error("Ошибка при проверке подключения к БД: %v", err)

		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("Успешное подключение к БД %s", dbname)

	repo := &Repository{db: db}

	logger.Info("Применение миграций БД...")
	if err := repo.runMigrations(context.Background()); err != nil {
		logger.Error("Ошибка при применении миграций: %v", err)

		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Info("Миграции БД применены успешно")

	return repo, nil
}

// Close закрывает соединение с БД
func (r *Repository) Close() error {
	return r.db.Close()
}

// GetAllClients возвращает список всех клиентов для конкретного пользователя
func (r *Repository) GetAllClients(ctx context.Context, userID int64) ([]*Client, error) {
	logger.Debug("Запрос списка клиентов для пользователя %d", userID)
	query := `SELECT id, user_id, name, host, port, username, password, ssl, created_at, updated_at 
	          FROM clients WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {

		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var clients []*Client
	for rows.Next() {
		var c Client
		err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Host, &c.Port, &c.Username, &c.Password, &c.SSL, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {

			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, &c)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Ошибка при итерации строк для пользователя %d: %v", userID, err)

		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	logger.Debug("Найдено %d клиентов для пользователя %d", len(clients), userID)

	return clients, nil
}

// GetClientByID возвращает клиента по ID, только если он принадлежит пользователю
func (r *Repository) GetClientByID(ctx context.Context, id int64, userID int64) (*Client, error) {
	logger.Debug("Запрос клиента ID=%d для пользователя %d", id, userID)
	query := `SELECT id, user_id, name, host, port, username, password, ssl, created_at, updated_at 
	          FROM clients WHERE id = $1 AND user_id = $2`

	var c Client
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Host, &c.Port, &c.Username, &c.Password, &c.SSL, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		logger.Debug("Клиент ID=%d не найден для пользователя %d", id, userID)

		return nil, nil
	}
	if err != nil {
		logger.Error("Ошибка при получении клиента ID=%d для пользователя %d: %v", id, userID, err)

		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	logger.Debug("Клиент ID=%d найден для пользователя %d: Name=%s", id, userID, c.Name)

	return &c, nil
}

// CreateClient создает нового клиента
func (r *Repository) CreateClient(ctx context.Context, client *Client) (*Client, error) {
	logger.Debugf("Создание клиента для пользователя %d: Name=%s, Host=%s:%d", client.UserID, client.Name, client.Host, client.Port)
	query := `INSERT INTO clients (user_id, name, host, port, username, password, ssl, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          RETURNING id, created_at, updated_at`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		client.UserID, client.Name, client.Host, client.Port, client.Username, client.Password, client.SSL, now, now,
	).Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)

	if err != nil {
		logger.Error("Ошибка при создании клиента для пользователя %d: %v", client.UserID, err)

		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	logger.Debugf("Клиент успешно создан: ID=%d, UserID=%d, Name=%s", client.ID, client.UserID, client.Name)

	return client, nil
}

// UpdateClient обновляет данные клиента, только если он принадлежит пользователю
func (r *Repository) UpdateClient(ctx context.Context, client *Client, userID int64) error {
	logger.Debugf("Обновление клиента ID=%d для пользователя %d: Name=%s, Host=%s:%d", client.ID, userID, client.Name, client.Host, client.Port)
	query := `UPDATE clients 
	          SET name = $1, host = $2, port = $3, username = $4, password = $5, ssl = $6, updated_at = $7
	          WHERE id = $8 AND user_id = $9`

	result, err := r.db.ExecContext(ctx, query,
		client.Name, client.Host, client.Port, client.Username, client.Password, client.SSL, time.Now(), client.ID, userID)
	if err != nil {
		logger.Error("Ошибка при обновлении клиента ID=%d для пользователя %d: %v", client.ID, userID, err)

		return fmt.Errorf("failed to update client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Ошибка при получении количества обновленных строк: %v", err)

		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Клиент ID=%d не найден или нет доступа для пользователя %d", client.ID, userID)

		return fmt.Errorf("client not found or access denied")
	}

	logger.Debugf("Клиент ID=%d успешно обновлен для пользователя %d", client.ID, userID)

	return nil
}

// DeleteClient удаляет клиента, только если он принадлежит пользователю
func (r *Repository) DeleteClient(ctx context.Context, id int64, userID int64) error {
	logger.Debugf("Удаление клиента ID=%d для пользователя %d", id, userID)
	query := `DELETE FROM clients WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		logger.Error("Ошибка при удалении клиента ID=%d для пользователя %d: %v", id, userID, err)

		return fmt.Errorf("failed to delete client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Ошибка при получении количества удаленных строк: %v", err)

		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Клиент ID=%d не найден или нет доступа для пользователя %d", id, userID)

		return fmt.Errorf("client not found or access denied")
	}

	logger.Debugf("Клиент ID=%d успешно удален для пользователя %d", id, userID)

	return nil
}

// SetUserState сохраняет состояние пользователя
func (r *Repository) SetUserState(ctx context.Context, userID int64, state string) error {
	logger.Debug("Сохранение состояния для пользователя %d: %s", userID, state)
	query := `INSERT INTO user_states (user_id, state, updated_at) 
	          VALUES ($1, $2, NOW())
	          ON CONFLICT (user_id) 
	          DO UPDATE SET state = $2, updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query, userID, state)
	if err != nil {
		logger.Error("Ошибка при сохранении состояния для пользователя %d: %v", userID, err)

		return fmt.Errorf("failed to set user state: %w", err)
	}

	return nil
}

// GetUserState получает состояние пользователя
func (r *Repository) GetUserState(ctx context.Context, userID int64) (string, error) {
	logger.Debug("Получение состояния для пользователя %d", userID)
	query := `SELECT state FROM user_states WHERE user_id = $1`

	var state string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&state)
	if err == sql.ErrNoRows {
		logger.Debug("Состояние не найдено для пользователя %d", userID)

		return "", nil
	}
	if err != nil {
		logger.Error("Ошибка при получении состояния для пользователя %d: %v", userID, err)

		return "", fmt.Errorf("failed to get user state: %w", err)
	}

	logger.Debug("Состояние найдено для пользователя %d: %s", userID, state)

	return state, nil
}

// DeleteUserState удаляет состояние пользователя
func (r *Repository) DeleteUserState(ctx context.Context, userID int64) error {
	logger.Debug("Удаление состояния для пользователя %d", userID)
	query := `DELETE FROM user_states WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		logger.Error("Ошибка при удалении состояния для пользователя %d: %v", userID, err)

		return fmt.Errorf("failed to delete user state: %w", err)
	}

	return nil
}

// GetAllUserStates возвращает все состояния пользователей (для загрузки при старте)
func (r *Repository) GetAllUserStates(ctx context.Context) (map[int64]string, error) {
	logger.Debug("Загрузка всех состояний пользователей")
	query := `SELECT user_id, state FROM user_states`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Ошибка при загрузке состояний: %v", err)

		return nil, fmt.Errorf("failed to query user states: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	states := make(map[int64]string)
	for rows.Next() {
		var userID int64
		var state string
		if err := rows.Scan(&userID, &state); err != nil {
			logger.Error("Ошибка при сканировании состояния: %v", err)

			return nil, fmt.Errorf("failed to scan user state: %w", err)
		}
		states[userID] = state
	}

	if err := rows.Err(); err != nil {
		logger.Error("Ошибка при итерации состояний: %v", err)

		return nil, fmt.Errorf("error iterating user states: %w", err)
	}

	logger.Debugf("Загружено %d состояний пользователей", len(states))

	return states, nil
}

// SetMenuMessageID сохраняет ID последнего сообщения меню для пользователя
func (r *Repository) SetMenuMessageID(ctx context.Context, userID int64, messageID int) error {
	logger.Debugf("Сохранение menu_message_id для пользователя %d: %d", userID, messageID)
	query := `UPDATE user_states 
	          SET menu_message_id = $1, updated_at = NOW()
	          WHERE user_id = $2`

	result, err := r.db.ExecContext(ctx, query, messageID, userID)
	if err != nil {
		logger.Error("Ошибка при сохранении menu_message_id для пользователя %d: %v", userID, err)

		return fmt.Errorf("failed to set menu message id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Ошибка при получении количества обновленных строк: %v", err)

		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		query = `INSERT INTO user_states (user_id, state, menu_message_id, updated_at)
		         VALUES ($1, '', $2, NOW())
		         ON CONFLICT (user_id)
		         DO UPDATE SET menu_message_id = $2, updated_at = NOW()`
		_, err = r.db.ExecContext(ctx, query, userID, messageID)
		if err != nil {
			logger.Error("Ошибка при создании/обновлении записи с menu_message_id для пользователя %d: %v", userID, err)

			return fmt.Errorf("failed to insert/update menu message id: %w", err)
		}
	}

	return nil
}

// GetMenuMessageID получает ID последнего сообщения меню для пользователя
func (r *Repository) GetMenuMessageID(ctx context.Context, userID int64) (int, error) {
	logger.Debug("Получение menu_message_id для пользователя %d", userID)
	query := `SELECT menu_message_id FROM user_states WHERE user_id = $1`

	var messageID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&messageID)
	if err == sql.ErrNoRows {
		logger.Debug("menu_message_id не найден для пользователя %d", userID)

		return 0, nil
	}
	if err != nil {
		logger.Error("Ошибка при получении menu_message_id для пользователя %d: %v", userID, err)

		return 0, fmt.Errorf("failed to get menu message id: %w", err)
	}

	if !messageID.Valid {

		return 0, nil
	}

	logger.Debug("menu_message_id найден для пользователя %d: %d", userID, messageID.Int64)

	return int(messageID.Int64), nil
}

// GetAllMenuMessageIDs возвращает все menu_message_id пользователей (для загрузки при старте)
func (r *Repository) GetAllMenuMessageIDs(ctx context.Context) (map[int64]int, error) {
	logger.Debug("Загрузка всех menu_message_id пользователей")
	query := `SELECT user_id, menu_message_id FROM user_states WHERE menu_message_id IS NOT NULL`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Ошибка при загрузке menu_message_id: %v", err)

		return nil, fmt.Errorf("failed to query menu message ids: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	messageIDs := make(map[int64]int)
	for rows.Next() {
		var userID int64
		var messageID sql.NullInt64
		if err := rows.Scan(&userID, &messageID); err != nil {
			logger.Error("Ошибка при сканировании menu_message_id: %v", err)

			return nil, fmt.Errorf("failed to scan menu message id: %w", err)
		}
		if messageID.Valid {
			messageIDs[userID] = int(messageID.Int64)
		}
	}

	if err := rows.Err(); err != nil {
		logger.Error("Ошибка при итерации menu_message_id: %v", err)

		return nil, fmt.Errorf("error iterating menu message ids: %w", err)
	}

	logger.Debugf("Загружено %d menu_message_id пользователей", len(messageIDs))

	return messageIDs, nil
}

// GetAllUserIDs возвращает список всех уникальных user_id из таблицы clients
func (r *Repository) GetAllUserIDs(ctx context.Context) ([]int64, error) {
	logger.Debug("Запрос всех уникальных user_id из таблицы clients")
	query := `SELECT DISTINCT user_id FROM clients ORDER BY user_id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Ошибка при запросе user_id: %v", err)

		return nil, fmt.Errorf("failed to query user ids: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			logger.Error("Ошибка при сканировании user_id: %v", err)

			return nil, fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Ошибка при итерации user_id: %v", err)

		return nil, fmt.Errorf("error iterating user ids: %w", err)
	}

	logger.Debug("Найдено %d уникальных пользователей", len(userIDs))

	return userIDs, nil
}

// SetUserTimezone сохраняет часовой пояс пользователя
func (r *Repository) SetUserTimezone(ctx context.Context, userID int64, timezone string) error {
	logger.Debug("Сохранение часового пояса для пользователя %d: %s", userID, timezone)
	query := `UPDATE user_states 
	          SET timezone = $1, updated_at = NOW()
	          WHERE user_id = $2`

	result, err := r.db.ExecContext(ctx, query, timezone, userID)
	if err != nil {
		logger.Error("Ошибка при сохранении часового пояса для пользователя %d: %v", userID, err)

		return fmt.Errorf("failed to set user timezone: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Ошибка при получении количества обновленных строк: %v", err)

		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		query = `INSERT INTO user_states (user_id, state, timezone, updated_at)
		         VALUES ($1, '', $2, NOW())
		         ON CONFLICT (user_id)
		         DO UPDATE SET timezone = $2, updated_at = NOW()`
		_, err = r.db.ExecContext(ctx, query, userID, timezone)
		if err != nil {
			logger.Error("Ошибка при создании/обновлении записи с timezone для пользователя %d: %v", userID, err)

			return fmt.Errorf("failed to insert/update timezone: %w", err)
		}
	}

	return nil
}

// GetUserTimezone получает часовой пояс пользователя
func (r *Repository) GetUserTimezone(ctx context.Context, userID int64) (string, error) {
	logger.Debug("Получение часового пояса для пользователя %d", userID)
	query := `SELECT timezone FROM user_states WHERE user_id = $1`

	var timezone sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&timezone)
	if err == sql.ErrNoRows {
		logger.Debug("Часовой пояс не найден для пользователя %d, используем Europe/Minsk", userID)

		return "Europe/Minsk", nil
	}
	if err != nil {
		logger.Error("Ошибка при получении часового пояса для пользователя %d: %v", userID, err)

		return "Europe/Minsk", fmt.Errorf("failed to get user timezone: %w", err)
	}

	if !timezone.Valid || timezone.String == "" {

		return "Europe/Minsk", nil
	}

	logger.Debug("Часовой пояс найден для пользователя %d: %s", userID, timezone.String)

	return timezone.String, nil
}
