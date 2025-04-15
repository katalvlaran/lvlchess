package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"lvlchess/internal/db/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
TournamentRepository manages the creation, retrieval, and updates
of the "tournaments" table, which holds data about multi-player tournaments.
*/
type TournamentRepository struct {
	pool *pgxpool.Pool
}

// NewTournamentRepository constructs a new TournamentRepository with the given pgxpool.
func NewTournamentRepository(pool *pgxpool.Pool) *TournamentRepository {
	return &TournamentRepository{pool: pool}
}

/*
CreateTournament inserts a new tournament record into the DB.
If the ID is empty, it generates a new UUID.
If CreatedAt is zero, it sets them to time.Now(), etc.
*/
func (r *TournamentRepository) CreateTournament(ctx context.Context, t *models.Tournament) error {
	// If ID is not provided, generate a fresh UUID
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}

	// Convert Players slice to JSON
	playersJSON, err := json.Marshal(t.Players)
	if err != nil {
		return fmt.Errorf("CreateTournament: marshal players: %w", err)
	}

	sql := `
INSERT INTO tournament (
  id,
  title,
  prise,
  players,
  status,
  start_at,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

/*
GetTournamentByID fetches a tournament row by its ID.
It also unmarshals the JSON players array into []int64.
*/
func (r *TournamentRepository) GetTournamentByID(ctx context.Context, tid string) (*models.Tournament, error) {
	sql := `
SELECT
  id,
  title,
  prise,
  players,
  status,
  start_at,
  created_at,
  updated_at
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

	// parse the array of players
	if unmarshalErr := json.Unmarshal(playersJSON, &t.Players); unmarshalErr != nil {
		// we can do a fallback or at least log it
		// t.Players = []int64{} // fallback if needed
	}
	return &t, nil
}

/*
AddPlayer modifies the 'players' array for a given tournament (tid)
by adding a userID if not already present. Then updates the record with the new JSON array.
*/
func (r *TournamentRepository) AddPlayer(ctx context.Context, tid string, userID int64) error {
	// 1) Retrieve the tournament
	t, err := r.GetTournamentByID(ctx, tid)
	if err != nil {
		return err
	}

	// 2) Check if user is already in t.Players
	for _, p := range t.Players {
		if p == userID {
			// Already in the tournament
			return nil
		}
	}

	// 3) Add new user ID to the players slice
	t.Players = append(t.Players, userID)
	t.UpdatedAt = time.Now()

	// 4) Marshal the new array
	playersJSON, err := json.Marshal(t.Players)
	if err != nil {
		return fmt.Errorf("AddPlayer: %w", err)
	}

	// 5) Perform an UPDATE to store the updated players array
	sql := `
UPDATE tournament
SET
  players    = $1,
  updated_at = NOW()
WHERE id = $2
`
	_, err = r.pool.Exec(ctx, sql, playersJSON, tid)
	if err != nil {
		return fmt.Errorf("AddPlayer: update players: %w", err)
	}
	return nil
}

/*
StartTournament sets the tournament status to active (1), and sets start_at=NOW().
*/
func (r *TournamentRepository) StartTournament(ctx context.Context, tid string) error {
	sql := `
UPDATE tournament
SET
  status   = $1,
  start_at = NOW(),
  updated_at = NOW()
WHERE id = $2
`
	_, err := r.pool.Exec(ctx, sql, models.TournamentStatusActive, tid)
	if err != nil {
		return fmt.Errorf("StartTournament: %w", err)
	}
	return nil
}
