package cache

import (
	"testing"
	"time"

	"github.com/notnil/chess"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	t.Run("Basic cache operations", func(t *testing.T) {
		cache := NewCache()
		game := chess.NewGame()
		pos := game.Position()

		// Проверяем отсутствие значения
		score, exists := cache.Get(pos)
		assert.False(t, exists)
		assert.Equal(t, float64(0), score)

		// Добавляем значение
		cache.Set(pos, 1.5)

		// Проверяем наличие значения
		score, exists = cache.Get(pos)
		assert.True(t, exists)
		assert.Equal(t, 1.5, score)
	})

	t.Run("Cache cleanup", func(t *testing.T) {
		cache := NewCache()
		game := chess.NewGame()
		pos := game.Position()

		cache.Set(pos, 1.5)

		// Модифицируем время последнего использования
		cache.positions[pos.String()] = CacheEntry{
			Score:     1.5,
			Timestamp: time.Now().Add(-2 * time.Hour),
		}

		// Запускаем очистку
		cache.cleanup()

		// Проверяем, что устаревшая запись удалена
		_, exists := cache.Get(pos)
		assert.False(t, exists)
	})
}
