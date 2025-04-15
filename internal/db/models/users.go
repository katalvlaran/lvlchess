package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

// UnregisteredPrivateChat = 0, used if a user doesn't have a personal chat ID assigned, or is unknown
const UnregisteredPrivateChat = 0

// User corresponds to the table "users" in the DB, storing basic info about each Telegram user.
type User struct {
	ID          int64  `json:"id"`           // Telegram user ID
	Username    string `json:"username"`     // e.g., @katalvlaran
	FirstName   string `json:"firstName"`    // If needed for display
	ChatID      int64  `json:"chatID"`       // A personal or private chat ID with the bot
	CurrentRoom *Room  `json:"current_room"` // Possibly unused. If needed, references the user's current room
	Rating      int    `json:"rating"`       // optional
	Wins        int    `json:"wins"`         // optional
	TotalGames  int    `json:"totalGames"`   // optional
}

// Validate ensures the user has an ID, username, etc.
func (u *User) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.ID, validation.Required),
		validation.Field(&u.Username, validation.Required, validation.Length(0, 255)),
		validation.Field(&u.FirstName, validation.Required, validation.Length(0, 255)),
		validation.Field(&u.ChatID, validation.Required),
	)
}
