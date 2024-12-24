package game

import (
	"fmt"
	"time"

	"github.com/katalvlaran/telega-shess/internal/utils"
	"github.com/notnil/chess"
)

// GameHandler обрабатывает игровую логику
type GameHandler struct {
	roomManager  *RoomManager
	validator    *RuleValidator
	renderer     *EnhancedBoardRenderer
	eventHandler *EventHandler
	log          *utils.Logger
}

// NewGameHandler создает новый обработчик игры
func NewGameHandler() *GameHandler {
	renderer := NewEnhancedBoardRenderer(DefaultTheme())
	return &GameHandler{
		roomManager:  NewRoomManager(),
		validator:    NewRuleValidator(),
		renderer:     renderer,
		eventHandler: NewEventHandler(renderer),
		log:          utils.Logger(),
	}
}

// MakeMove выполняет ход в игре
func (gh *GameHandler) MakeMove(roomID string, playerID int64, move Move) error {
	room, err := gh.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	// Проверка очереди хода
	if err := gh.validateTurn(room, playerID); err != nil {
		return err
	}

	// Создание хода в формате chess библиотеки
	algebraicMove := chess.UCINotation{}.Decode(room.State.Game.Position(), fmt.Sprintf("%s%s", move.From, move.To))
	if algebraicMove == nil {
		return fmt.Errorf("invalid move format")
	}

	// Проверка правил
	if err := gh.validator.ValidateMove(room.State.Game, algebraicMove); err != nil {
		return err
	}

	// Проверка валидности хода
	if err := room.State.Game.Move(algebraicMove); err != nil {
		return fmt.Errorf("invalid move: %v", err)
	}

	// Обновление времени
	now := time.Now()
	if room.State.Game.Position().Turn() == chess.White {
		room.State.BlackTime -= now.Sub(room.State.LastMove)
	} else {
		room.State.WhiteTime -= now.Sub(room.State.LastMove)
	}
	room.State.LastMove = now

	// Проверка взятия
	if pos := room.State.Game.Position(); pos.Board().Piece(algebraicMove.S2()) != chess.NoPiece {
		event := gh.generateEvent(room, algebraicMove, EventCapture)
		gh.eventHandler.HandleEvent(event)
	}

	// Проверка рокировки
	if isCastling(algebraicMove) {
		event := gh.generateEvent(room, algebraicMove, EventCastling)
		gh.eventHandler.HandleEvent(event)
	}

	// Проверка взятия на проходе
	if isEnPassant(pos, algebraicMove) {
		event := gh.generateEvent(room, algebraicMove, EventEnPassant)
		gh.eventHandler.HandleEvent(event)
	}

	// Проверяем результат хода
	newPos := room.State.Game.Position()
	if newPos.InCheck() {
		event := gh.generateEvent(room, algebraicMove, EventCheck)
		gh.eventHandler.HandleEvent(event)
	}

	if room.State.Game.Outcome() == chess.Checkmate {
		event := gh.generateEvent(room, algebraicMove, EventCheckmate)
		gh.eventHandler.HandleEvent(event)
		room.State.IsFinished = true
	}

	gh.log.WithFields(map[string]interface{}{
		"room_id":   roomID,
		"player_id": playerID,
		"move":      fmt.Sprintf("%s%s", move.From, move.To),
	}).Info("Move made")

	return nil
}

// validateTurn проверяет, чей сейчас ход
func (gh *GameHandler) validateTurn(room *Room, playerID int64) error {
	isWhiteTurn := room.State.Game.Position().Turn() == chess.White
	if (isWhiteTurn && playerID != room.WhitePlayer) || (!isWhiteTurn && playerID != room.BlackPlayer) {
		return fmt.Errorf("not your turn")
	}
	return nil
}

// GetRenderedBoard возвращает отрендеренную доску
func (gh *GameHandler) GetRenderedBoard(roomID string) (string, error) {
	room, err := gh.roomManager.GetRoom(roomID)
	if err != nil {
		return "", err
	}

	return gh.renderer.RenderEnhancedBoard(room.State.Game), nil
}

// generateEvent генерирует игровое событие
func (gh *GameHandler) generateEvent(room *Room, move *chess.Move, eventType EventType) *GameEvent {
	event := &GameEvent{
		Type:      eventType,
		Move:      move,
		Position:  room.State.Game.Position(),
		Timestamp: time.Now(),
		PlayerID:  gh.getCurrentPlayer(room),
		Extra:     make(map[string]interface{}),
	}

	// Добавляем дополнительную информацию в зависимости от типа события
	switch eventType {
	case EventPromotion:
		event.Extra["promoted_to"] = move.Promotion.String()
	case EventCastling:
		if move.S2() == chess.G1 || move.S2() == chess.G8 {
			event.Extra["side"] = "короткая"
		} else {
			event.Extra["side"] = "длинная"
		}
	}

	return event
}

// getCurrentPlayer возвращает ID текущего игрока
func (gh *GameHandler) getCurrentPlayer(room *Room) int64 {
	if room.State.Game.Position().Turn() == chess.White {
		return room.WhitePlayer
	}
	return room.BlackPlayer
}
