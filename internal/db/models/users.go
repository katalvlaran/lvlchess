package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

const UnregisteredPrivateChat = 0

type User struct {
	ID          int64  `json:"id"` // Telegram user ID
	Username    string `json:"username"`
	FirstName   string `json:"firstName"`
	ChatID      int64  `json:"chatID"`
	CurrentRoom *Room  `json:"current_room"`
	Rating      int    `json:"rating"`
	Wins        int    `json:"wins"`
	TotalGames  int    `json:"totalGames"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.ID, validation.Required),
		validation.Field(&u.Username, validation.Required, validation.Length(0, 255)),
		validation.Field(&u.FirstName, validation.Required, validation.Length(0, 255)),
		validation.Field(&u.ChatID, validation.Required),
		//validation.Field(&u.CurrentRoom, validation.NilOrNotEmpty),
		//validation.Field(&u.Rating, validation.Required),
		//validation.Field(&u.Wins, validation.Required),
		//validation.Field(&u.TotalGames, validation.Required),
	)
}
