package telegram

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/katalvlaran/telega-shess/internal/game"
	"github.com/katalvlaran/telega-shess/internal/utils"
)

// Bot представляет основную структуру бота
type Bot struct {
	api         *tgbotapi.BotAPI
	gameHandler *game.GameHandler
	log         *utils.Logger
	updates     tgbotapi.UpdatesChannel
	sessions    *SessionManager
}

// Session хранит информацию о текущей игровой сессии пользователя
type Session struct {
	UserID      int64
	CurrentRoom string
	State       SessionState
	LastCommand string
	WaitingFor  WaitingState
	LastMessage *tgbotapi.Message
	OpponentID  int64
	TimeControl int
}

// SessionState определяет состояние сессии пользователя
type SessionState int

const (
	StateIdle SessionState = iota
	StateCreatingRoom
	StateJoiningRoom
	StateInGame
	StateSelectingTimeControl
	StateWaitingForOpponent
)

// WaitingState определяет, чего ожидает бот от пользователя
type WaitingState int

const (
	WaitingNone WaitingState = iota
	WaitingForMove
	WaitingForPromotion
	WaitingForDrawResponse
	WaitingForTimeControl
)

// SessionManager управляет игровыми сессиями
type SessionManager struct {
	sessions map[int64]*Session
	mu       sync.RWMutex
}

// NewSessionManager создает новый менеджер сессий
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]*Session),
	}
}

// GetSession возвращает сессию пользователя
func (sm *SessionManager) GetSession(userID int64) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return nil
	}
	return session
}

// CreateSession создает новую сессию
func (sm *SessionManager) CreateSession(userID int64) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		UserID:     userID,
		State:      StateIdle,
		WaitingFor: WaitingNone,
	}
	sm.sessions[userID] = session
	return session
}

// UpdateSession обновляет состояние сессии
func (sm *SessionManager) UpdateSession(userID int64, update func(*Session)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		update(session)
	}
}
