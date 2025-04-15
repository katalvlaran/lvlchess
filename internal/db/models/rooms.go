package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/notnil/chess"
)

const (
	RoomStatusWaiting  = "waiting"  // A room is waiting for a second player
	RoomStatusPlaying  = "playing"  // A room has two players; game is ongoing
	RoomStatusFinished = "finished" // A room has ended (checkmate or draw)
)

// Room represents a single chess "room" or match session between players.
type Room struct {
	RoomID      string    `json:"room_id"`       // Unique identifier (UUID)
	RoomTitle   string    `json:"room_title"`    // Title/nickname of the room
	Player1ID   int64     `json:"player_1"`      // Telegram user ID of the first player
	Player2ID   *int64    `json:"player_2"`      // Telegram user ID of the second player, nil if not joined
	Status      string    `json:"status"`        // One of RoomStatusWaiting|RoomStatusPlaying|RoomStatusFinished
	BoardState  string    `json:"board_state"`   // FEN string representing current board position
	IsWhiteTurn bool      `json:"is_white_turn"` // Whose turn it is; 'true' means White's turn
	WhiteID     *int64    `json:"white_id"`      // Which player is assigned the White pieces
	BlackID     *int64    `json:"black_id"`      // Which player is assigned the Black pieces
	ChatID      *int64    `json:"chat_id"`       // Group chat ID if this room is associated with a Telegram group
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks basic constraints, e.g., non-empty RoomID, valid status, etc.
func (u *Room) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.RoomID, validation.Required),
		validation.Field(&u.Player1ID, validation.Required),
		validation.Field(&u.Player1ID, validation.NilOrNotEmpty),
		validation.Field(&u.Status, validation.Required,
			validation.In(RoomStatusWaiting, RoomStatusPlaying, RoomStatusFinished)),
		validation.Field(&u.BoardState, validation.Required),
	)
}

// PrepareNewRoom is a helper that builds a new Room object.
// By default, the BoardState is the standard chess initial position (via chess.NewGame().FEN()).
func PrepareNewRoom(p1ID int64, title string) *Room {
	return &Room{
		RoomID:      uuid.NewString(),
		RoomTitle:   title,
		Player1ID:   p1ID,
		Status:      RoomStatusWaiting,
		BoardState:  chess.NewGame().FEN(),
		IsWhiteTurn: true, // Typically starts with White
	}
}
