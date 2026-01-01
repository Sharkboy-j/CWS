package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	currentLevel LogLevel = INFO
	logger       *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", 0)
}

func SetLevel(level LogLevel) {
	currentLevel = level
}

func SetLevelFromString(level string) {
	switch level {
	case "DEBUG", "debug":
		currentLevel = DEBUG
	case "INFO", "info":
		currentLevel = INFO
	case "WARN", "warn", "WARNING", "warning":
		currentLevel = WARN
	case "ERROR", "error":
		currentLevel = ERROR
	default:
		currentLevel = INFO
	}
}

func getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func shouldLog(level LogLevel) bool {
	return level >= currentLevel
}

func formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	levelStr := getLevelString(level)
	message := fmt.Sprintf(format, args...)

	return fmt.Sprintf("[%s] [%s] %s", timestamp, levelStr, message)
}

func Debug(format string, args ...interface{}) {
	if shouldLog(DEBUG) {
		logger.Println(formatMessage(DEBUG, format, args...))
	}
}

func Info(format string, args ...interface{}) {
	if shouldLog(INFO) {
		logger.Println(formatMessage(INFO, format, args...))
	}
}

func Warn(format string, args ...interface{}) {
	if shouldLog(WARN) {
		logger.Println(formatMessage(WARN, format, args...))
	}
}

func Error(format string, args ...interface{}) {
	if shouldLog(ERROR) {
		logger.Println(formatMessage(ERROR, format, args...))
	}
}

func Debugf(format string, args ...interface{}) {
	Debug(format, args...)
}

func Infof(format string, args ...interface{}) {
	Info(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Warn(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Error(format, args...)
}
