package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/katalvlaran/telega-shess/internal/utils"
	"github.com/notnil/chess"
)

// RoomManager управляет игровыми комнатами
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
	log   *utils.Logger
}

// NewRoomManager создает новый менеджер комнат
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
		log:   utils.Logger(),
	}
}

// CreateRoom создает новую игровую комнату
func (rm *RoomManager) CreateRoom(creatorID int64, timeLimit time.Duration) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Генерация уникального ID комнаты
	roomID := fmt.Sprintf("room_%d", time.Now().UnixNano())

	// Создание новой игры
	game := chess.NewGame()

	room := &Room{
		ID:          roomID,
		WhitePlayer: creatorID, // Создатель всегда играет белыми
		CreatedAt:   time.Now(),
		State: &GameState{
			Game:      game,
			TimeLimit: timeLimit,
			LastMove:  time.Now(),
		},
	}

	rm.rooms[roomID] = room
	rm.log.WithField("room_id", roomID).Info("New room created")

	return room, nil
}

// GetRoom возвращает комнату по ID
func (rm *RoomManager) GetRoom(roomID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room %s not found", roomID)
	}
	return room, nil
}

// DeleteRoom удаляет комнату
func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[roomID]; exists {
		delete(rm.rooms, roomID)
		rm.log.WithField("room_id", roomID).Info("Room deleted")
	}
}
