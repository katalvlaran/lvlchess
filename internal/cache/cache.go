package cache

import (
	"sync"
	"time"

	"github.com/notnil/chess"
)

// Cache представляет кэш для оценок позиций
type Cache struct {
	positions map[string]CacheEntry
	mu        sync.RWMutex
}

// CacheEntry представляет запись в кэше
type CacheEntry struct {
	Score     float64
	Timestamp time.Time
}

// NewCache создает новый кэш
func NewCache() *Cache {
	cache := &Cache{
		positions: make(map[string]CacheEntry),
	}
	go cache.cleanup()
	return cache
}

// Get получает оценку позиции из кэша
func (c *Cache) Get(pos *chess.Position) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := pos.String()
	if entry, exists := c.positions[key]; exists {
		if time.Since(entry.Timestamp) < time.Hour {
			return entry.Score, true
		}
	}
	return 0, false
}

// Set сохраняет оценку позиции в кэш
func (c *Cache) Set(pos *chess.Position, score float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := pos.String()
	c.positions[key] = CacheEntry{
		Score:     score,
		Timestamp: time.Now(),
	}
}

// cleanup периодически очищает устаревшие записи
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.positions {
			if now.Sub(entry.Timestamp) > time.Hour {
				delete(c.positions, key)
			}
		}
		c.mu.Unlock()
	}
}
