package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"telega_chess/internal/db/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TournamentRepository struct {
	pool *pgxpool.Pool
}

func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{pool: pool}
}

// CreateTournament - вставляет новую запись в table 'tournament'
func (r *TournamentRepository) CreateTournament(ctx context.Context, t *models.Tournament) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	// сериализуем t.Players -> JSON
	playersJSON, err := json.Marshal(t.Players)
	if err != nil {
		return fmt.Errorf("CreateTournament: marshal players: %w", err)
	}

	sql := `
INSERT INTO tournament(id, title, prise, players, status, start_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
`
	_, err = r.pool.Exec(ctx, sql,
		t.ID,
		t.Title,
		t.Prise,
		playersJSON,
		t.Status,
		t.StartAt,
		t.CreatedAt,
		t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("CreateTournament: %w", err)
	}
	return nil
}

// GetTournamentByID - возвращает одну запись
func (r *TournamentRepository) GetTournamentByID(ctx context.Context, tid string) (*models.Tournament, error) {
	sql := `
SELECT id, title, prise, players, status, start_at, created_at, updated_at
FROM tournament
WHERE id = $1
`
	row := r.pool.QueryRow(ctx, sql, tid)
	var t models.Tournament
	var playersJSON []byte
	err := row.Scan(
		&t.ID,
		&t.Title,
		&t.Prise,
		&playersJSON,
		&t.Status,
		&t.StartAt,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTournamentByID: %w", err)
	}
	// парсим players
	if err := json.Unmarshal(playersJSON, &t.Players); err != nil {
		// можно просто warning вывести
		// t.Players = []int64{} // fallback
	}
	return &t, nil
}

// AddPlayer - добавляет игрока в массив players
func (r *TournamentRepository) AddPlayer(ctx context.Context, tid string, userID int64) error {
	// 1) Берём турнир
	t, err := r.GetTournamentByID(ctx, tid)
	if err != nil {
		return err
	}
	// 2) Проверяем, нет ли уже userID в t.Players
	for _, p := range t.Players {
		if p == userID {
			// уже есть
			return nil
		}
	}
	// 3) Добавляем
	t.Players = append(t.Players, userID)
	t.UpdatedAt = time.Now()

	playersJSON, err := json.Marshal(t.Players)
	if err != nil {
		return fmt.Errorf("AddPlayer: %w", err)
	}

	// 4) UPDATE
	sql := `
UPDATE tournament
SET players = $1,
    updated_at = NOW()
WHERE id = $2
`
	_, err = r.pool.Exec(ctx, sql, playersJSON, tid)
	if err != nil {
		return fmt.Errorf("AddPlayer: update players: %w", err)
	}

	return nil
}

// StartTournament - переводит турнир в status=1 (active), устанавливает start_at=NOW()
func (r *TournamentRepository) StartTournament(ctx context.Context, tid string) error {
	sql := `
UPDATE tournament
SET status=$1,
    start_at=NOW(),
    updated_at=NOW()
WHERE id=$2
`
	_, err := r.pool.Exec(ctx, sql, models.TournamentStatusActive, tid)
	if err != nil {
		return fmt.Errorf("StartTournament: %w", err)
	}
	return nil
}
