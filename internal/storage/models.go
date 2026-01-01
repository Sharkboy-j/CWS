package storage

import "time"

type Client struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"` // Telegram chat_id пользователя
	Name      string    `db:"name"`
	Host      string    `db:"host"`
	Port      int32     `db:"port"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	SSL       bool      `db:"ssl"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
