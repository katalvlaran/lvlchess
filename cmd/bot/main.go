package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/katalvlaran/telega-shess/internal/game"
	"github.com/katalvlaran/telega-shess/internal/monitoring"
	"github.com/katalvlaran/telega-shess/internal/utils"
)

func main() {
	// Инициализация логгера
	logger := utils.InitLogger()
	logger.Info("Starting chess bot...")

	// Загрузка конфигурации
	config := utils.LoadConfig()
	if config.BotToken == "" {
		logger.Fatal("BOT_TOKEN is not set")
	}

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация метрик
	metrics := monitoring.NewMetricsCollector()
	gameHandler, err := game.NewGameHandler(config, metrics)
	if err != nil {
		logger.Fatalf("Failed to create game handler: %v", err)
	}
	// Инициализация обработчика игры
	gameHandler := game.NewGameHandler()

	// Настройка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	// Запуск обработчика сигналов
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		// Создаем контекст с таймаутом для graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Запускаем процесс graceful shutdown
		if err := gameHandler.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("Error during shutdown: %v", err)
		}

		cancel() // Отменяем основной контекст
		logger.Infof("Received signal: %v", sig)
		cancel()
	}()

	if err != context.Canceled {
		logger.Fatalf("Failed to start server: %v", err)
	}
	logger.Info("Server stopped gracefully")
	if err := gameHandler.Start(ctx); err != nil {

		logger.Info("Chess bot shutdown complete")
		logger.Fatalf("Failed to start server: %v", err)
	}
}
