package game

import (
	"time"

	"github.com/notnil/chess"
)

// GameState представляет текущее состояние игры
type GameState struct {
	Game       *chess.Game
	WhiteTime  time.Duration
	BlackTime  time.Duration
	TimeLimit  time.Duration
	LastMove   time.Time
	IsFinished bool
}

// Room представляет игровую комнату
type Room struct {
	ID          string
	WhitePlayer int64 // Telegram User ID
	BlackPlayer int64 // Telegram User ID
	State       *GameState
	CreatedAt   time.Time
}

// Move представляет ход в игре
type Move struct {
	From      string    // Например, "e2"
	To        string    // Например, "e4"
	Promotion string    // Фигура для превращения пешки
	Timestamp time.Time // Время хода
}
