package monitoring

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	moveLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chess_move_latency_seconds",
			Help:    "Latency of chess moves",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"difficulty", "phase"},
	)

	cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chess_cache_hits_total",
			Help: "Number of cache hits",
		},
		[]string{"type"},
	)

	activeGames = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "chess_active_games",
			Help: "Number of active games",
		},
	)
)

func init() {
	prometheus.MustRegister(moveLatency)
	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(activeGames)
}

// MetricsCollector собирает метрики производительности
type MetricsCollector struct {
	mu sync.RWMutex
	// Внутренние метрики
	moveCount     int
	averageDepth  float64
	cacheHitRate  float64
	lastCollected time.Time
}

// NewMetricsCollector создает новый коллектор метрик
func NewMetricsCollector() *MetricsCollector {
	mc := &MetricsCollector{
		lastCollected: time.Now(),
	}
	go mc.periodicCollection()
	return mc
}

// RecordMove записывает метрики хода
func (mc *MetricsCollector) RecordMove(duration time.Duration, difficulty string, phase string) {
	moveLatency.WithLabelValues(difficulty, phase).Observe(duration.Seconds())

	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.moveCount++
}

// RecordCacheHit записывает попадание в кэш
func (mc *MetricsCollector) RecordCacheHit(hitType string) {
	cacheHits.WithLabelValues(hitType).Inc()
}

// UpdateActiveGames обновляет количество активных игр
func (mc *MetricsCollector) UpdateActiveGames(count int) {
	activeGames.Set(float64(count))
}

// periodicCollection периодически собирает метрики
func (mc *MetricsCollector) periodicCollection() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		mc.collectMetrics()
	}
}

// collectMetrics собирает текущие метрики
func (mc *MetricsCollector) collectMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Сброс метрик
	mc.moveCount = 0
	mc.averageDepth = 0
	mc.cacheHitRate = 0
	mc.lastCollected = time.Now()
}
