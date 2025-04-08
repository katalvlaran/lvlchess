package repositories

import (
	"context"
	"fmt"

	"telega_chess/internal/db/models"
	"telega_chess/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// UsersRepository — набор методов для работы с таблицей users
type UsersRepository struct {
	pool *pgxpool.Pool
}

// NewUsersRepository — конструктор
func NewUsersRepository(pool *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{pool: pool}
}

// CreateOrUpdateUser — аналогичен старому db.CreateOrUpdateUser(...)
func (repo *UsersRepository) CreateOrUpdateUser(ctx context.Context, u *models.User) error {
	// валидация
	if err := u.Validate(); err != nil {
		return err
	}
	// подготовленный запрос
	sql := `
INSERT INTO users (id, user_name, first_name, chat_id, rating, wins, total_games)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
   SET user_name    = EXCLUDED.user_name,
       first_name   = EXCLUDED.first_name,
       chat_id      = EXCLUDED.chat_id,
       rating       = EXCLUDED.rating,
       wins         = EXCLUDED.wins,
       total_games  = EXCLUDED.total_games
    `
	_, err := repo.pool.Exec(ctx, sql,
		u.ID, u.Username, u.FirstName, u.ChatID, u.Rating, u.Wins, u.TotalGames)
	if err != nil {
		utils.Logger.Error("CreateOrUpdateUser error: "+err.Error(), zap.Error(err))
		return fmt.Errorf("CreateOrUpdateUser: %v", err)
	}
	return nil
}

// GetUserByID — аналог db.GetUserByID(id)
func (repo *UsersRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	sql := `SELECT id, user_name, first_name, chat_id, rating, wins, total_games FROM users WHERE id = $1;`
	row := repo.pool.QueryRow(ctx, sql, id)

	var u models.User
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
		return nil, fmt.Errorf("GetUserByID: %v", err)
	}
	return &u, nil
}
