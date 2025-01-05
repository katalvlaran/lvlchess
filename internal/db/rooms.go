package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Room модель для таблицы rooms

type Room struct {
	RoomID          string
	Player1ID       int64
	Player2ID       *int64 // null, если второй игрок не присоединился
	Status          string
	BoardState      *string // null, если ещё не зафиксировали доску
	WhiteID         *int64
	BlackID         *int64
	ChatID          *int64  // null, если комнату-группу ещё не создали
	Player1Username *string // null, если не знаем username
	Player2Username *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CreateRoom — как и раньше, но без FEN/цветов, т.к. их присвоим, когда второй игрок зайдёт
func CreateRoom(player1ID int64) (Room, error) {
	r := Room{
		RoomID:     uuid.NewString(),
		Player1ID:  player1ID,
		Player2ID:  nil,
		Status:     "waiting",
		BoardState: nil,
		WhiteID:    nil,
		BlackID:    nil,
	}

	sql := `INSERT INTO rooms (room_id, player1_id, status)
			VALUES ($1, $2, $3)`
	_, err := Pool.Exec(context.Background(), sql,
		r.RoomID, r.Player1ID, r.Status)
	if err != nil {
		return Room{}, fmt.Errorf("ошибка INSERT rooms: %v", err)
	}
	return r, nil
}

// UpdateRoom обобщённо обновляет любую информацию
func UpdateRoom(r Room) error {
	sql := `UPDATE rooms
            SET player2_id = $1,
                status     = $2,
                board_state= $3,
                white_id   = $4,
                black_id   = $5,
                chat_id    = $6,
                player1_username = $7,
                player2_username = $8,
                updated_at = NOW()
            WHERE room_id  = $9;`
	_, err := Pool.Exec(context.Background(), sql,
		r.Player2ID,
		r.Status,
		r.BoardState,
		r.WhiteID,
		r.BlackID,
		r.ChatID,
		r.Player1Username,
		r.Player2Username,
		r.RoomID,
	)
	if err != nil {
		return fmt.Errorf("не удалось UPDATE rooms: %v", err)
	}
	return nil
}

// GetRoomByID достаём все поля, включая board_state, white_id, black_id
func GetRoomByID(roomID string) (Room, error) {
	var r Room
	sql := `SELECT room_id, player1_id, player2_id,
                   status, board_state, white_id, black_id,
                   chat_id, player1_username, player2_username,
                   created_at, updated_at
            FROM rooms
            WHERE room_id = $1;`
	row := Pool.QueryRow(context.Background(), sql, roomID)
	err := row.Scan(
		&r.RoomID,
		&r.Player1ID,
		&r.Player2ID,
		&r.Status,
		&r.BoardState,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
		&r.Player1Username,
		&r.Player2Username,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err != nil {
		return Room{}, fmt.Errorf("GetRoomByID: %v", err)
	}
	if r.RoomID == "" {
		return Room{}, errors.New("комната не найдена")
	}
	return r, nil
}

// GetRoomByID достаём все поля, включая board_state, white_id, black_id
func GetRoomByChatID(chatID string) (Room, error) {
	var r Room
	sql := `SELECT * FROM rooms WHERE chat_id = $1;`
	row := Pool.QueryRow(context.Background(), sql, chatID)
	err := row.Scan(
		&r.RoomID,
		&r.Player1ID,
		&r.Player2ID,
		&r.Status,
		&r.BoardState,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
		&r.Player1Username,
		&r.Player2Username,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err != nil {
		return Room{}, fmt.Errorf("GetRoomByChatID: %v", err)
	}
	if r.RoomID == "" {
		return Room{}, errors.New("комната не найдена")
	}
	return r, nil
}

// DeleteRoom удаляет запись из таблицы rooms
func DeleteRoom(roomID string) error {
	sql := `DELETE FROM rooms WHERE room_id = $1;`
	_, err := Pool.Exec(context.Background(), sql, roomID)
	if err != nil {
		return fmt.Errorf("ошибка удаления комнаты: %v", err)
	}
	return nil
}

// Пример - можно проверить статус комнаты
func IsRoomExist(roomID string) (bool, error) {
	r, err := GetRoomByID(roomID)
	if err != nil {
		return false, err
	}
	if r.RoomID == "" {
		return false, errors.New("комната не найдена")
	}
	return true, nil
}
