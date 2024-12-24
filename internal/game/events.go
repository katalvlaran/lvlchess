package game

import (
	"time"

	"github.com/notnil/chess"
)

// EventType определяет тип игрового события
type EventType int

const (
	EventMove EventType = iota
	EventCapture
	EventCheck
	EventCheckmate
	EventDraw
	EventPromotion
	EventCastling
	EventEnPassant
	EventTimeOut
	EventResign
)

// GameEvent представляет событие в игре
type GameEvent struct {
	Type      EventType
	Move      *chess.Move
	Position  *chess.Position
	Timestamp time.Time
	PlayerID  int64
	Extra     map[string]interface{} // Дополнительные данные события
}

// EventHandler обрабатывает игровые события
type EventHandler struct {
	renderer *EnhancedBoardRenderer
	log      *utils.Logger
}

// NewEventHandler создает новый обработчик событий
func NewEventHandler(renderer *EnhancedBoardRenderer) *EventHandler {
	return &EventHandler{
		renderer: renderer,
		log:      utils.Logger(),
	}
}

// HandleEvent обрабатывает игровое событие и возвращает сообщение для отображения
func (eh *EventHandler) HandleEvent(event *GameEvent) string {
	var message string

	switch event.Type {
	case EventCapture:
		message = eh.handleCapture(event)
	case EventCheck:
		message = eh.handleCheck(event)
	case EventCheckmate:
		message = eh.handleCheckmate(event)
	case EventPromotion:
		message = eh.handlePromotion(event)
	case EventCastling:
		message = eh.handleCastling(event)
	case EventEnPassant:
		message = eh.handleEnPassant(event)
	case EventTimeOut:
		message = eh.handleTimeout(event)
	case EventResign:
		message = eh.handleResign(event)
	}

	eh.log.WithFields(map[string]interface{}{
		"event_type": event.Type,
		"player_id":  event.PlayerID,
		"timestamp":  event.Timestamp,
	}).Info("Game event processed")

	return message
}

// Обработчики конкретных событий
func (eh *EventHandler) handleCapture(event *GameEvent) string {
	return utils.CaptureSymbol + " Взятие фигуры!"
}

func (eh *EventHandler) handleCheck(event *GameEvent) string {
	return "⚠️ Шах!"
}

func (eh *EventHandler) handleCheckmate(event *GameEvent) string {
	return "🏆 Мат! Игра окончена."
}

func (eh *EventHandler) handlePromotion(event *GameEvent) string {
	promotedTo := event.Extra["promoted_to"].(string)
	return utils.PromotionSymbol + " Превращение пешки в " + promotedTo + " " + utils.SparklesSymbol
}

func (eh *EventHandler) handleCastling(event *GameEvent) string {
	side := event.Extra["side"].(string)
	return "🏰 Рокировка " + side
}

func (eh *EventHandler) handleEnPassant(event *GameEvent) string {
	return "👻 Взятие на проходе!"
}

func (eh *EventHandler) handleTimeout(event *GameEvent) string {
	return "⏰ Время истекло!"
}

func (eh *EventHandler) handleResign(event *GameEvent) string {
	return "🏳️ Игрок сдался"
}
