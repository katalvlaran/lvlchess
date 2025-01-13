package db

import (
	"context"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/notnil/chess"
	"go.uber.org/zap"

	"telega_chess/internal/utils"
)

const (
	RoomStatusWaiting  = "waiting"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"
)

// Room модель для таблицы rooms
type Room struct {
	RoomID      string
	Player1     *User     `json:"player_1"`
	Player2     *User     `json:"player_2"` // null, если второй игрок не присоединился
	Status      string    `json:"status"`   // waiting/playing/finished
	BoardState  string    `json:"board_state"`
	IsWhiteTurn bool      `json:"is_white_turn"`
	WhiteID     *int64    `json:"white_id"`
	BlackID     *int64    `json:"black_id"`
	ChatID      *int64    `json:"chat_id"` // null, если комнату-группу ещё не создали
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *Room) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.RoomID, validation.Required),
		validation.Field(&u.Player1, validation.Required),
		validation.Field(&u.Player2, validation.NilOrNotEmpty),
		validation.Field(&u.Status, validation.Required, validation.In(RoomStatusWaiting, RoomStatusPlaying, RoomStatusFinished)),
		validation.Field(&u.BoardState, validation.Required),
		//validation.Field(&u.IsWhiteTurn, validation.NilOrNotEmpty),
		//validation.Field(&u.WhiteID, validation.Required),
		//validation.Field(&u.BlackID, validation.Required),
		//validation.Field(&u.ChatID, validation.Required),
	)
}
func CreateRoom(player1ID int64) (*Room, error) {
	p1, err := GetUserByID(player1ID)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("GetUserByID%d:%s", player1ID, err.Error()), zap.Error(err))
		return nil, fmt.Errorf("CreateRoom: %v", err)
	}

	r := Room{
		RoomID:      uuid.NewString(),
		Player1:     p1,
		Status:      RoomStatusWaiting,
		BoardState:  chess.NewGame().FEN(),
		IsWhiteTurn: true,
	}
	utils.Logger.Info("CreateRoom:", zap.Any("r", r))

	if err = r.Validate(); err != nil {
		return nil, err
	}
	sql := `INSERT INTO rooms (room_id, player1_id, status, board_state, is_white_turn)
			VALUES ($1, $2, $3, $4, $5)`
	_, err = Pool.Exec(context.Background(), sql, r.RoomID, r.Player1.ID, r.Status, r.BoardState, r.IsWhiteTurn)
	if err != nil {
		utils.Logger.Error("INSERT:"+err.Error(), zap.Error(err))
		return nil, fmt.Errorf("CreateRoom: %v", err)
	}
	return &r, nil
}

func GetRoomByID(roomID string) (*Room, error) {
	var r Room
	var p1, p2 *User
	var p1ID, p2ID *int64
	sql := `SELECT room_id, player1_id, player2_id, status, board_state, is_white_turn, white_id, black_id,chat_id FROM rooms WHERE room_id = $1;`
	row := Pool.QueryRow(context.Background(), sql, roomID)
	utils.Logger.Info("GetRomByID:", zap.Any("roomID", roomID))
	utils.Logger.Info("GetRomByID:", zap.Any("row", row))
	err := row.Scan(
		&r.RoomID,
		&p1ID,
		&p2ID,
		&r.Status,
		&r.BoardState,
		&r.IsWhiteTurn,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
	)
	if err != nil {
		utils.Logger.Error("GetRoomByID:"+err.Error(), zap.Error(err))
		return nil, err
	}

	p1, err = GetUserByID(*p1ID)
	if err != nil {
		return nil, fmt.Errorf("CreateRoom: %v", err)
	}

	if p2ID != nil {
		p2, err = GetUserByID(*p2ID)
		if err != nil {
			return nil, fmt.Errorf("CreateRoom: %v", err)
		}
	}

	r.Player1, r.Player2 = p1, p2

	return &r, err
}

func GetRoomByChatID(chatID int64) (*Room, error) {
	var r Room
	var p1, p2 *User
	var p1ID, p2ID *int64
	sql := `SELECT room_id, player1_id, player2_id, status, board_state, is_white_turn, white_id, black_id,chat_id FROM rooms WHERE chat_id = $1;`
	row := Pool.QueryRow(context.Background(), sql, chatID)
	err := row.Scan(
		&r.RoomID,
		&p1ID,
		&p2ID,
		&r.Status,
		&r.BoardState,
		&r.IsWhiteTurn,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
	)
	if err != nil {
		return nil, err
	}

	p1, err = GetUserByID(*p1ID)
	if err != nil {
		return nil, fmt.Errorf("CreateRoom: %v", err)
	}

	if p2ID != nil {
		p2, err = GetUserByID(*p2ID)
		if err != nil {
			return nil, fmt.Errorf("CreateRoom: %v", err)
		}
	}

	r.Player1, r.Player2 = p1, p2

	return &r, err
}

func UpdateRoom(r *Room) error {
	var p2ID *int64
	if r.Player2 != nil {
		p2ID = &r.Player2.ID
	}

	if err := r.Validate(); err != nil {
		utils.Logger.Info("UpdateRoom", zap.Any("room:", r))
		utils.Logger.Error(err.Error(), zap.Error(err))
		return err
	}
	utils.Logger.Debug("UpdateRoom", zap.Any("room:", &r), zap.Any("p2ID:", p2ID))
	sql := `UPDATE rooms
			SET player2_id = $1,
			    status = $2,
			    board_state = $3,
			    is_white_turn = $4,
			    white_id = $5,
			    black_id = $6,
			    chat_id = $7,
			    updated_at = NOW()
			WHERE room_id = $8;`
	_, err := Pool.Exec(context.Background(), sql,
		p2ID,
		r.Status,
		r.BoardState,
		r.IsWhiteTurn,
		r.WhiteID,
		r.BlackID,
		r.ChatID,
		r.RoomID,
	)

	if err != nil {
		utils.Logger.Error(err.Error(), zap.Error(err))
		return fmt.Errorf("UpdateRoom: %v", err)
	}
	return nil
}
