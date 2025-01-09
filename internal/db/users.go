package db

import (
	"context"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"go.uber.org/zap"

	"telega_chess/internal/utils"
)

const UnregisteredPrivateChat = 0

type User struct {
	ID         int64 // Telegram user ID
	Username   string
	FirstName  string
	ChatID     int64
	Rating     int
	Wins       int
	TotalGames int
}

func (u *User) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.ID, validation.Required),
		validation.Field(&u.Username, validation.Required),
		//validation.Field(&u.FirstName, validation.Required),
		validation.Field(&u.ChatID, validation.Required),
		//validation.Field(&u.Rating, validation.Required),
		//validation.Field(&u.Wins, validation.Required),
		//validation.Field(&u.TotalGames, validation.Required),
	)
}

// CreateOrUpdateUser - сохраняем/обновляем юзера
func CreateOrUpdateUser(u *User) error {
	utils.Logger.Info("CreateOrUpdateUser:", zap.Any("user", u))
	sql := `
	INSERT INTO users (id, user_name, first_name, chat_id, rating, wins, total_games)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (id) DO UPDATE
	    SET user_name = $2,
	        first_name = $3,
	        chat_id = $4,
	        rating = $5,
	        wins = $6,
	        total_games = $7
	`
	_, err := Pool.Exec(context.Background(), sql,
		u.ID, u.Username, u.FirstName, u.ChatID, u.Rating, u.Wins, u.TotalGames)
	utils.Logger.Error("INSERT:", zap.Error(err))
	if err != nil {
		return fmt.Errorf("CreateOrUpdateUser: %v", err)

	}
	return nil
}

func GetUserByID(id int64) (*User, error) {
	var u User
	sql := `SELECT * FROM users WHERE id = $1;`
	row := Pool.QueryRow(context.Background(), sql, id)
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.FirstName,
		&u.ChatID,
		&u.Rating,
		&u.Wins,
		&u.TotalGames,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Пример расширений: UpdateWins, UpdateRating и т.д.
