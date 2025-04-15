package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"lvlchess/internal/db/models"
	"lvlchess/internal/utils"
)

// ErrUniqueViolation is a sentinel error message typically triggered
// by a Postgres UNIQUE constraint (e.g., "players_pair" on (player1_id, player2_id)).
const ErrUniqueViolation = "unique_violation"

/*
RoomsRepository provides CRUD-like operations for the "rooms" table.
It manages the creation of new chess rooms, retrieving and updating them.
*/
type RoomsRepository struct {
	pool *pgxpool.Pool
}

// NewRoomsRepository is a constructor function for RoomsRepository.
func NewRoomsRepository(pool *pgxpool.Pool) *RoomsRepository {
	return &RoomsRepository{pool: pool}
}

/*
CreateRoom inserts a new record into the "rooms" table. If the DB constraint
violates the unique pairing of Player1 + Player2, it returns an error
indicating ErrUniqueViolation (like 23505 for "unique_violation").
*/
func (r *RoomsRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	// Validate the model before inserting
	if err := room.Validate(); err != nil {
		return fmt.Errorf("room.Validate: %w", err)
	}

	sql := `
INSERT INTO rooms (
  room_id,
  room_title,
  player1_id,
  status,
  board_state,
  is_white_turn,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
`
	_, err := r.pool.Exec(ctx, sql,
		room.RoomID,
		room.RoomTitle,
		room.Player1ID,
		room.Status,
		room.BoardState,
		room.IsWhiteTurn,
	)
	if err != nil {
		// If the DB error is a unique violation on the constraint
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				// We transform that into our internal error constant
				return errors.New(ErrUniqueViolation)
			}
		}
		utils.Logger.Error("INSERT: "+err.Error(), zap.Error(err))
		return fmt.Errorf("CreateRoom: %v", err)
	}
	return nil
}

/*
GetRoomByID fetches a single room by its room_id (UUID).
Returns the matched record or an error if not found.
*/
func (r *RoomsRepository) GetRoomByID(ctx context.Context, roomID string) (*models.Room, error) {
	sql := `
SELECT 
  room_id,
  room_title,
  player1_id,
  player2_id,
  status,
  board_state,
  is_white_turn,
  white_id,
  black_id,
  chat_id,
  created_at,
  updated_at
FROM rooms
WHERE room_id = $1
`
	row := r.pool.QueryRow(ctx, sql, roomID)

	var rm models.Room
	err := row.Scan(
		&rm.RoomID,
		&rm.RoomTitle,
		&rm.Player1ID,
		&rm.Player2ID,
		&rm.Status,
		&rm.BoardState,
		&rm.IsWhiteTurn,
		&rm.WhiteID,
		&rm.BlackID,
		&rm.ChatID,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByID: %v", err)
	}
	return &rm, nil
}

/*
GetRoomByChatID allows looking up a room that is tied to a particular Telegram group chat.
If no such record is found, an error is returned.
*/
func (r *RoomsRepository) GetRoomByChatID(ctx context.Context, chatID int64) (*models.Room, error) {
	sql := `
SELECT 
  room_id,
  room_title,
  player1_id,
  player2_id,
  status,
  board_state,
  is_white_turn,
  white_id,
  black_id,
  chat_id,
  created_at,
  updated_at
FROM rooms
WHERE chat_id = $1
`
	row := r.pool.QueryRow(ctx, sql, chatID)

	var rm models.Room
	err := row.Scan(
		&rm.RoomID,
		&rm.RoomTitle,
		&rm.Player1ID,
		&rm.Player2ID,
		&rm.Status,
		&rm.BoardState,
		&rm.IsWhiteTurn,
		&rm.WhiteID,
		&rm.BlackID,
		&rm.ChatID,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByChatID: %v", err)
	}
	return &rm, nil
}

/*
GetRoomByPlayerIDs tries to find a room where the given players
are either in a "waiting" or "playing" status. For example,
it checks if the pair (p1ID, p2ID) or (p2ID, p1ID) matches an existing record.
*/
func (r *RoomsRepository) GetRoomByPlayerIDs(ctx context.Context, p1ID, p2ID int64) (*models.Room, error) {
	sql := `
SELECT 
    room_id,
    room_title,
    player1_id,
    player2_id,
    status,
    board_state,
    is_white_turn,
    white_id,
    black_id,
    chat_id,
    created_at,
    updated_at
FROM rooms
WHERE status IN('waiting','playing')
  AND 
`
	if p2ID != 0 {
		// If we have a second player's ID
		// We want something like: (player1_id = p1 AND player2_id = p2) OR (player1_id = p2 AND player2_id = p1)
		sql += fmt.Sprintf(
			`( (player1_id = %d AND player2_id = %d) OR (player1_id = %d AND player2_id = %d) )`,
			p1ID, p2ID, p2ID, p1ID)
	} else {
		// In case p2ID = 0 or something else
		sql += fmt.Sprintf(
			`( (player1_id = %d AND player2_id IS NULL) OR (player1_id IS NULL AND player2_id = %d) )`,
			p1ID, p2ID)
	}
	sql += ";"

	row := r.pool.QueryRow(ctx, sql, p1ID, p2ID)

	var rm models.Room
	err := row.Scan(
		&rm.RoomID,
		&rm.RoomTitle,
		&rm.Player1ID,
		&rm.Player2ID,
		&rm.Status,
		&rm.BoardState,
		&rm.IsWhiteTurn,
		&rm.WhiteID,
		&rm.BlackID,
		&rm.ChatID,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByPlayerIDs: %v", err)
	}
	return &rm, nil
}

/*
UpdateRoom modifies the existing record in "rooms", changing
fields like Title, second player, status, board_state, etc.
*/
func (r *RoomsRepository) UpdateRoom(ctx context.Context, room *models.Room) error {
	if err := room.Validate(); err != nil {
		return fmt.Errorf("UpdateRoom Validate: %w", err)
	}
	sql := `
UPDATE rooms
SET
    room_title     = $1,
    player2_id     = $2,
    status         = $3,
    board_state    = $4,
    is_white_turn  = $5,
    white_id       = $6,
    black_id       = $7,
    chat_id        = $8,
    updated_at     = NOW()
WHERE room_id = $9
`
	_, err := r.pool.Exec(ctx, sql,
		room.RoomTitle,
		room.Player2ID,
		room.Status,
		room.BoardState,
		room.IsWhiteTurn,
		room.WhiteID,
		room.BlackID,
		room.ChatID,
		room.RoomID,
	)
	if err != nil {
		return fmt.Errorf("UpdateRoom exec: %v", err)
	}
	return nil
}

/*
GetPlayingRoomsForUser attempts to return all "active" or "waiting" rooms
for the specified user. It filters by (player1_id=$1 OR player2_id=$1).
*/
func (r *RoomsRepository) GetPlayingRoomsForUser(ctx context.Context, userID int64) ([]models.Room, error) {
	sql := `
SELECT
  room_id,
  room_title,
  board_state,
  is_white_turn,
  updated_at
FROM rooms
WHERE status IN ('waiting','playing')
  AND (player1_id = $1 OR player2_id = $1)
ORDER BY updated_at DESC
`
	rows, err := r.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("GetPlayingRoomsForUser: %v", err)
	}
	defer rows.Close()

	var result []models.Room
	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.RoomID,
			&rm.RoomTitle,
			&rm.BoardState,
			&rm.IsWhiteTurn,
			&rm.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		// Note: we are not scanning p1, p2, WhiteID, etc. here
		result = append(result, rm)
	}
	return result, nil
}
