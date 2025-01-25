package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"telega_chess/internal/db/models"
	"telega_chess/internal/utils"
)

const ErrUniqueViolation = "unique_violation"

// RoomsRepository — набор методов для работы с таблицей rooms
type RoomsRepository struct {
	pool *pgxpool.Pool
}

func NewRoomsRepository(pool *pgxpool.Pool) *RoomsRepository {
	return &RoomsRepository{pool: pool}
}

// CreateRoom — аналог старого db.CreateRoom(...)
func (r *RoomsRepository) CreateRoom(ctx context.Context, room *models.Room) error {
	if err := room.Validate(); err != nil {
		return fmt.Errorf("room.Validate: %w", err)
	}
	sql := `
INSERT INTO rooms (room_id, room_title, player1_id, status, board_state, is_white_turn, created_at, updated_at)
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
		// Здесь отлавливаем unique_violation
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				// Это ошибка unique_violation
				// нужно вернуть специальную ошибку, по которой
				// наверху поймём, что пара (player1, player2) уже существует.
				return errors.New(ErrUniqueViolation)
				// (Либо сделать кастомную ошибку e.g. errors.New("UniquePlayersPair"))
			}
		}
		utils.Logger.Error("INSERT:"+err.Error(), zap.Error(err))

		return fmt.Errorf("CreateRoom: %v", err)
	}
	return nil
}

// GetRoomByID — берёт одну комнату по room_id
func (r *RoomsRepository) GetRoomByID(ctx context.Context, roomID string) (*models.Room, error) {
	sql := `
SELECT 
  room_id, room_title, player1_id, player2_id, status,
  board_state, is_white_turn, white_id, black_id, chat_id,
  created_at, updated_at
FROM rooms
WHERE room_id = $1
`
	row := r.pool.QueryRow(ctx, sql, roomID)
	var rm models.Room
	//var p2id *int64
	//var chatID *int64
	//var whiteID *int64
	//var blackID *int64

	err := row.Scan(
		&rm.RoomID,
		&rm.RoomTitle,
		&rm.Player1ID,
		//&p2id,
		&rm.Player2ID,
		&rm.Status,
		&rm.BoardState,
		&rm.IsWhiteTurn,
		&rm.WhiteID, //&whiteID,
		&rm.BlackID, //&blackID,
		&rm.ChatID,  //&chatID,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByID: %v", err)
	}
	//rm.Player2ID = p2id
	//rm.WhiteID = whiteID
	//rm.BlackID = blackID
	//rm.ChatID = chatID
	return &rm, nil
}

// GetRoomByChatID — берёт одну комнату по chat_id
func (r *RoomsRepository) GetRoomByChatID(ctx context.Context, chatID int64) (*models.Room, error) {
	sql := `
SELECT 
  room_id, room_title, player1_id, player2_id, status,
  board_state, is_white_turn, white_id, black_id, chat_id,
  created_at, updated_at
FROM rooms
WHERE chat_id = $1
`
	row := r.pool.QueryRow(ctx, sql, chatID)
	var rm models.Room
	//var p2id *int64
	//var chatID *int64
	//var whiteID *int64
	//var blackID *int64

	err := row.Scan(
		&rm.RoomID,
		&rm.RoomTitle,
		&rm.Player1ID,
		//&p2id,
		&rm.Player2ID,
		&rm.Status,
		&rm.BoardState,
		&rm.IsWhiteTurn,
		&rm.WhiteID, //&whiteID,
		&rm.BlackID, //&blackID,
		&rm.ChatID,  //&chatID,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByID: %v", err)
	}
	//rm.Player2ID = p2id
	//rm.WhiteID = whiteID
	//rm.BlackID = blackID
	//rm.ChatID = chatID
	return &rm, nil
}

// GetRoomByPlayerIDs — берёт одну комнату по Player1ID и Player2ID
func (r *RoomsRepository) GetRoomByPlayerIDs(ctx context.Context, p1ID, p2ID int64) (*models.Room, error) {
	sql := `
SELECT 
    room_id, room_title, player1_id, player2_id, status,
	board_state, is_white_turn, white_id, black_id,chat_id 
FROM rooms 
WHERE status in('waiting', 'playing') 
  and,
`
	if p2ID != 0 {
		sql += fmt.Sprintf(
			` (player1_id = %d and player2_id = %d) OR (player1_id = %d and player2_id = %d);`,
			p1ID, p2ID, p2ID, p1ID)
	} else {
		sql += fmt.Sprintf(` (player1_id = %d and player2_id IS NULL) OR (player1_id IS NULL and player2_id = %d);`,
			p1ID, p2ID)
	}
	row := r.pool.QueryRow(ctx, sql, p1ID, p2ID)
	var rm models.Room
	//var p2id *int64
	//var chatID *int64
	//var whiteID *int64
	//var blackID *int64

	err := row.Scan(
		&rm.RoomID,
		&rm.RoomTitle,
		&rm.Player1ID,
		//&p2id,
		&rm.Player2ID,
		&rm.Status,
		&rm.BoardState,
		&rm.IsWhiteTurn,
		&rm.WhiteID, //&whiteID,
		&rm.BlackID, //&blackID,
		&rm.ChatID,  //&chatID,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRoomByID: %v", err)
	}
	//rm.Player2ID = p2id
	//rm.WhiteID = whiteID
	//rm.BlackID = blackID
	//rm.ChatID = chatID
	return &rm, nil
}

// UpdateRoom
func (r *RoomsRepository) UpdateRoom(ctx context.Context, room *models.Room) error {
	if err := room.Validate(); err != nil {
		return fmt.Errorf("UpdateRoom Validate: %w", err)
	}
	sql := `
UPDATE rooms
SET room_title = $1,
    player2_id = $2,
    status     = $3,
    board_state = $4,
    is_white_turn = $5,
    white_id   = $6,
    black_id   = $7,
    chat_id    = $8,
    updated_at = NOW()
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

// Пример получения списка "playing" / "waiting"
func (r *RoomsRepository) GetPlayingRoomsForUser(ctx context.Context, userID int64) ([]models.Room, error) {
	sqlQ := `
SELECT room_id, room_title, board_state, is_white_turn, updated_at
FROM rooms
WHERE status in ('waiting','playing')
AND (player1_id=$1 OR player2_id=$1)
ORDER BY updated_at DESC
`
	rows, err := r.pool.Query(ctx, sqlQ, userID)
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
		// Player1ID, Player2ID, etc. не извлекаем (или сделаем JOIN)
		result = append(result, rm)
	}
	return result, nil
}
