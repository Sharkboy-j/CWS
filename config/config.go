package config

type Config struct {
	RutrackerApiToken string `config:"rutracker_api_token,required" json:"rutracker_api_token"`
	TelegramToken     string `config:"telegram_token,required" json:"telegram_token"`
	ChatId            int64  `config:"telegram_chat_id,required" json:"telegram_chat_id"`
	DurationSeconds   int    `config:"duration_seconds" json:"duration_seconds"`
	ManualCheckOnly   bool   `config:"only_manual_check" json:"only_manual_check"`
	RutrackerHost     string `config:"rutracker_host" json:"rutracker_host"`
	// PostgreSQL configuration
	DBHost     string `config:"db_host" json:"db_host"`
	DBPort     int    `config:"db_port" json:"db_port"`
	DBUser     string `config:"db_user" json:"db_user"`
	DBPassword string `config:"db_password" json:"db_password"`
	DBName     string `config:"db_name" json:"db_name"`
	LogLevel   string `config:"log_level" json:"log_level"`
}
