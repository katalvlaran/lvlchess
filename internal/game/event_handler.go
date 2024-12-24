package game

import (
	"sync"
	"time"

	"github.com/katalvlaran/telega-shess/internal/utils"
)

// EventBuffer буферизует события для пакетной обработки
type EventBuffer struct {
	events []GameEvent
	mu     sync.Mutex
}

// NewEventBuffer создает новый буфер событий
func NewEventBuffer() *EventBuffer {
	buffer := &EventBuffer{}
	go buffer.processEvents()
	return buffer
}

// AddEvent добавляет событие в буфер
func (eb *EventBuffer) AddEvent(event GameEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.events = append(eb.events, event)
}

// processEvents обрабатывает события пакетами
func (eb *EventBuffer) processEvents() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		eb.mu.Lock()
		if len(eb.events) == 0 {
			eb.mu.Unlock()
			continue
		}

		events := make([]GameEvent, len(eb.events))
		copy(events, eb.events)
		eb.events = eb.events[:0]
		eb.mu.Unlock()

		// Обрабатываем события пакетом
		for _, event := range events {
			switch event.Type {
			case EventMove:
				handleMoveEvent(event)
			case EventGameEnd:
				handleGameEndEvent(event)
			case EventDraw:
				handleDrawEvent(event)
			}
		}
	}
}

// Обновляем GameHandler для использования буфера событий
func (gh *GameHandler) HandleEvent(event GameEvent) {
	gh.eventBuffer.AddEvent(event)
}

// Вспомогательные функции для обработки событий
func handleMoveEvent(event GameEvent) {
	// Обработка хода
	utils.Logger().WithField("event", event).Info("Processing move event")
}

func handleGameEndEvent(event GameEvent) {
	// Обработка окончания игры
	utils.Logger().WithField("event", event).Info("Processing game end event")
}

func handleDrawEvent(event GameEvent) {
	// Обработка ничьей
	utils.Logger().WithField("event", event).Info("Processing draw event")
}
