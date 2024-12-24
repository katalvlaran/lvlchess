package game

import (
	"time"

	"github.com/notnil/chess"
)

// GameState представляет состояние игры
type GameState struct {
	ID            string
	Game          *chess.Game
	WhitePlayerID int64
	BlackPlayerID int64
	Status        string
	LastActivity  time.Time
	DrawOffer     *DrawOffer
}

// DrawOffer представляет предложение ничьей
type DrawOffer struct {
	OfferedBy int64
	OfferedAt time.Time
}

// NewGameState создает новое состояние игры
func NewGameState(id string, whitePlayerID, blackPlayerID int64) *GameState {
	return &GameState{
		ID:            id,
		Game:          chess.NewGame(),
		WhitePlayerID: whitePlayerID,
		BlackPlayerID: blackPlayerID,
		Status:        GameStateNew,
		LastActivity:  time.Now(),
	}
}

// UpdateLastActivity обновляет время последней активности
func (gs *GameState) UpdateLastActivity() {
	gs.LastActivity = time.Now()
}

// IsPlayerTurn проверяет, ход ли игрока
func (gs *GameState) IsPlayerTurn(playerID int64) bool {
	if gs.Game.Position().Turn() == chess.White {
		return playerID == gs.WhitePlayerID
	}
	return playerID == gs.BlackPlayerID
}

// GetOpponentID возвращает ID оппонента
func (gs *GameState) GetOpponentID(playerID int64) int64 {
	if playerID == gs.WhitePlayerID {
		return gs.BlackPlayerID
	}
	return gs.WhitePlayerID
}
