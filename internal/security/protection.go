package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"sync"
	"time"
)

// TokenManager управляет токенами доступа
type TokenManager struct {
	tokens    map[string]TokenInfo
	mu        sync.RWMutex
	maxTokens int
}

type TokenInfo struct {
	UserID    int64
	ExpiresAt time.Time
	LastUsed  time.Time
}

// NewTokenManager создает новый менеджер токенов
func NewTokenManager(maxTokens int) *TokenManager {
	tm := &TokenManager{
		tokens:    make(map[string]TokenInfo),
		maxTokens: maxTokens,
	}
	go tm.cleanup()
	return tm
}

// ValidateToken проверяет токен и возвращает ID пользователя
func (tm *TokenManager) ValidateToken(token string) (int64, bool) {
	if token == "" {
		return 0, false
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.tokens[token]
	if !exists || time.Now().After(info.ExpiresAt) {
		return 0, false
	}

	return info.UserID, true
}

// CreateToken создает новый токен для пользователя
func (tm *TokenManager) CreateToken(userID int64) (string, error) {
	token := generateSecureToken()
	if token == "" {
		return "", ErrTokenGeneration
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if len(tm.tokens) >= tm.maxTokens {
		tm.removeOldestToken()
	}

	tm.tokens[token] = TokenInfo{
		UserID:    userID,
		ExpiresAt: time.Now().Add(TokenExpiration),
		LastUsed:  time.Now(),
	}

	return token, nil
}

// generateSecureToken генерирует криптографически безопасный токен
func generateSecureToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// cleanup периодически очищает устаревшие токены
func (tm *TokenManager) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		tm.mu.Lock()
		now := time.Now()
		for token, info := range tm.tokens {
			if now.After(info.ExpiresAt) {
				delete(tm.tokens, token)
			}
		}
		tm.mu.Unlock()
	}
}

// removeOldestToken удаляет самый старый токен
func (tm *TokenManager) removeOldestToken() {
	var oldestToken string
	var oldestTime time.Time

	for token, info := range tm.tokens {
		if oldestToken == "" || info.LastUsed.Before(oldestTime) {
			oldestToken = token
			oldestTime = info.LastUsed
		}
	}

	if oldestToken != "" {
		delete(tm.tokens, oldestToken)
	}
}

// SecureCompare безопасно сравнивает строки
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
