package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/katalvlaran/telega-shess/internal/ai"
	"github.com/katalvlaran/telega-shess/internal/monitoring"
	"github.com/katalvlaran/telega-shess/internal/utils"
)

// GameHandler обрабатывает игровые события и управляет состоянием игры
type GameHandler struct {
	config   *utils.Config
	metrics  *monitoring.MetricsCollector
	games    map[string]*GameState
	mu       sync.RWMutex
	bot      *ai.ChessBot
	shutdown chan struct{}
}

// NewGameHandler создает новый обработчик игры
func NewGameHandler(config *utils.Config, metrics *monitoring.MetricsCollector) (*GameHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if metrics == nil {
		return nil, fmt.Errorf("metrics is nil")
	}

	difficulty, err := ai.ParseDifficulty(config.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("invalid difficulty: %w", err)
	}

	return &GameHandler{
		config:   config,
		metrics:  metrics,
		games:    make(map[string]*GameState),
		bot:      ai.NewChessBot(difficulty),
		shutdown: make(chan struct{}),
	}, nil
}

// Start запускает обработчик игры
func (gh *GameHandler) Start(ctx context.Context) error {
	// Запуск периодической очистки неактивных игр
	go gh.cleanupInactiveGames(ctx)

	// Ожидание завершения контекста
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown выполняет корректное завершение работы
func (gh *GameHandler) Shutdown(ctx context.Context) error {
	// Сигнал о начале завершения работы
	close(gh.shutdown)

	// Ожидание завершения всех активных игр
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("shutdown timeout")
		case <-ticker.C:
			gh.mu.RLock()
			if len(gh.games) == 0 {
				gh.mu.RUnlock()
				return nil
			}
			gh.mu.RUnlock()
		}
	}
}

// cleanupInactiveGames периодически очищает неактивные игры
func (gh *GameHandler) cleanupInactiveGames(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gh.mu.Lock()
			now := time.Now()
			for id, game := range gh.games {
				if now.Sub(game.LastActivity) > 30*time.Minute {
					delete(gh.games, id)
					gh.metrics.UpdateActiveGames(len(gh.games))
				}
			}
			gh.mu.Unlock()
		}
	}
}
