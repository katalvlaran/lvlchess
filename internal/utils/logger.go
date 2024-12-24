package utils

import (
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger

// InitLogger инициализирует логгер
func InitLogger() *log.Logger {
	if logger != nil {
		return logger
	}

	// Создаем директорию для логов если её нет
	logDir := filepath.Dir(LogFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	// Открываем файл для логов
	file, err := os.OpenFile(LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Создаем логгер
	logger = log.New(file, "", log.LstdFlags)
	return logger
}
