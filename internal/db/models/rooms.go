package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/notnil/chess"
)

const (
	RoomStatusWaiting  = "waiting"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"
)

// Room модель для таблицы rooms
type Room struct {
	RoomID    string `json:"room_id"`
	RoomTitle string `json:"room_title"`
	Player1ID int64  `json:"player_1"`
	Player2ID *int64 `json:"player_2"` // null, если второй игрок не присоединился
	//Player1     *User     `json:"player_1"`
	//Player2     *User     `json:"player_2"` // null, если второй игрок не присоединился
	Status      string    `json:"status"` // waiting/playing/finished
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
		validation.Field(&u.Player1ID, validation.Required),
		validation.Field(&u.Player1ID, validation.NilOrNotEmpty),
		validation.Field(&u.Status, validation.Required, validation.In(RoomStatusWaiting, RoomStatusPlaying, RoomStatusFinished)),
		validation.Field(&u.BoardState, validation.Required),
		//validation.Field(&u.IsWhiteTurn, validation.NilOrNotEmpty),
		//validation.Field(&u.WhiteID, validation.Required),
		//validation.Field(&u.BlackID, validation.Required),
		//validation.Field(&u.ChatID, validation.Required),
	)
}

func PrepareNewRoom(p1ID int64, title string) *Room {
	return &Room{
		RoomID:      uuid.NewString(),
		RoomTitle:   title,
		Player1ID:   p1ID,
		Status:      RoomStatusWaiting,
		BoardState:  chess.NewGame().FEN(),
		IsWhiteTurn: true,
	}
}
