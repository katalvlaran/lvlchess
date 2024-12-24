package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter реализует ограничение запросов
type RateLimiter struct {
	requests map[string]*RequestLimit
	mu       sync.RWMutex
}

type RequestLimit struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter создает новый лимитер
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*RequestLimit),
	}
}

// RateLimit middleware для ограничения запросов
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mu.Lock()
		limit, exists := rl.requests[ip]
		if !exists {
			limit = &RequestLimit{}
			rl.requests[ip] = limit
		}

		now := time.Now()
		if now.Sub(limit.lastSeen) > time.Minute {
			limit.count = 0
			limit.lastSeen = now
		}

		limit.count++
		if limit.count > 60 { // 60 запросов в минуту
			rl.mu.Unlock()
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders добавляет заголовки безопасности
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
