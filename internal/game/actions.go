package game

import (
	"fmt"
	"time"
)

// ResignGame позволяет игроку сдаться
func (gh *GameHandler) ResignGame(roomID string, playerID int64) error {
	room, err := gh.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	if playerID != room.WhitePlayer && playerID != room.BlackPlayer {
		return fmt.Errorf("player not in this game")
	}

	event := &GameEvent{
		Type:      EventResign,
		Timestamp: time.Now(),
		PlayerID:  playerID,
	}
	gh.eventHandler.HandleEvent(event)

	room.State.IsFinished = true
	return nil
}

// OfferDraw предлагает ничью
func (gh *GameHandler) OfferDraw(roomID string, playerID int64) error {
	room, err := gh.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	if playerID != room.WhitePlayer && playerID != room.BlackPlayer {
		return fmt.Errorf("player not in this game")
	}

	// Здесь можно добавить логику для обработки предложения ничьей
	// Например, сохранить предложение и ждат�� ответа от противника

	return nil
}

// AcceptDraw принимает предложение ничьей
func (gh *GameHandler) AcceptDraw(roomID string, playerID int64) error {
	room, err := gh.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	event := &GameEvent{
		Type:      EventDraw,
		Timestamp: time.Now(),
		PlayerID:  playerID,
	}
	gh.eventHandler.HandleEvent(event)

	room.State.IsFinished = true
	return nil
}
