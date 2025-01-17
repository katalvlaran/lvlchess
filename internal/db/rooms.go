package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/notnil/chess"
	"go.uber.org/zap"

	"telega_chess/internal/utils"
)

const (
	RoomStatusWaiting  = "waiting"
	RoomStatusPlaying  = "playing"
	RoomStatusFinished = "finished"

	ErrUniqueViolation = "unique_violation"
)

// Room модель для таблицы rooms
type Room struct {
	RoomID      string    `json:"room_id"`
	RoomTitle   string    `json:"room_title"`
	Player1     *User     `json:"player_1"`
	Player2     *User     `json:"player_2"` // null, если второй игрок не присоединился
	Status      string    `json:"status"`   // waiting/playing/finished
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
		validation.Field(&u.Player1, validation.Required),
		validation.Field(&u.Player2, validation.NilOrNotEmpty),
		validation.Field(&u.Status, validation.Required, validation.In(RoomStatusWaiting, RoomStatusPlaying, RoomStatusFinished)),
		validation.Field(&u.BoardState, validation.Required),
		//validation.Field(&u.IsWhiteTurn, validation.NilOrNotEmpty),
		//validation.Field(&u.WhiteID, validation.Required),
		//validation.Field(&u.BlackID, validation.Required),
		//validation.Field(&u.ChatID, validation.Required),
	)
}
func CreateRoom(r *Room) error {
	utils.Logger.Info("CreateRoom:", zap.Any("r", r))
	if err := r.Validate(); err != nil {
		return fmt.Errorf("r.Validate(): %v", err)
	}
	sql := `INSERT INTO rooms (room_id, room_title, player1_id, status, board_state, is_white_turn)
			VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := Pool.Exec(context.Background(), sql, &r.RoomID, &r.RoomTitle, &r.Player1.ID, &r.Status, &r.BoardState, &r.IsWhiteTurn)
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

		return fmt.Errorf("INSERT: %v", err)
	}

	return nil
}

func PrepareNewRoom(p1 *User, title string) *Room {
	return &Room{
		RoomID:      uuid.NewString(),
		RoomTitle:   title,
		Player1:     p1,
		Status:      RoomStatusWaiting,
		BoardState:  chess.NewGame().FEN(),
		IsWhiteTurn: true,
	}
}

func GetRoomByID(roomID string) (*Room, error) {
	var r Room
	var p1ID, p2ID *int64
	sql := `SELECT room_id, room_title, player1_id, player2_id, status, board_state, is_white_turn, white_id, black_id,chat_id FROM rooms WHERE room_id = $1;`
	row := Pool.QueryRow(context.Background(), sql, roomID)
	utils.Logger.Info("GetRomByID:", zap.Any("roomID", roomID))
	err := row.Scan(
		&r.RoomID,
		&r.RoomTitle,
		&p1ID,
		&p2ID,
		&r.Status,
		&r.BoardState,
		&r.IsWhiteTurn,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
	)
	utils.Logger.Info("GetRomByID:", zap.Any("room", r))
	if err != nil {
		utils.Logger.Error("GetRoomByID:"+err.Error(), zap.Error(err))
		return nil, err
	}

	r.Player1, err = GetUserByID(*p1ID)
	if err != nil {
		return nil, fmt.Errorf("CreateRoom: %v", err)
	}

	if p2ID != nil {
		r.Player2, err = GetUserByID(*p2ID)
		if err != nil {
			return nil, fmt.Errorf("CreateRoom: %v", err)
		}
	}

	return &r, err
}

func GetRoomByChatID(chatID int64) (*Room, error) {
	var r Room
	var p1ID, p2ID *int64
	sql := `SELECT room_id, room_title, player1_id, player2_id, status, board_state, is_white_turn, white_id, black_id,chat_id FROM rooms WHERE chat_id = $1;`
	row := Pool.QueryRow(context.Background(), sql, chatID)
	err := row.Scan(
		&r.RoomID,
		&r.RoomTitle,
		&p1ID,
		&p2ID,
		&r.Status,
		&r.BoardState,
		&r.IsWhiteTurn,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
	)
	if err != nil {
		return nil, err
	}

	r.Player1, err = GetUserByID(*p1ID)
	if err != nil {
		return nil, fmt.Errorf("CreateRoom: %v", err)
	}

	if p2ID != nil {
		r.Player2, err = GetUserByID(*p2ID)
		if err != nil {
			return nil, fmt.Errorf("CreateRoom: %v", err)
		}
	}

	return &r, err
}

func GetRoomByPlayerIDs(p1ID, p2ID int64) (*Room, error) {
	var r Room
	sql := `SELECT room_id, room_title, player1_id, player2_id, status, board_state, is_white_turn, white_id, black_id,chat_id `
	sql += fmt.Sprintf(`FROM rooms WHERE status in('%s','%s') and`, RoomStatusWaiting, RoomStatusPlaying)
	if p2ID != 0 {
		sql += fmt.Sprintf(
			` (player1_id = %d and player2_id = %d) OR (player1_id = %d and player2_id = %d);`,
			p1ID, p2ID, p2ID, p1ID)
	} else {
		sql += fmt.Sprintf(` (player1_id = %d and player2_id IS NULL) OR (player1_id IS NULL and player2_id = %d);`,
			p1ID, p2ID)
	}
	row := Pool.QueryRow(context.Background(), sql)
	err := row.Scan(
		&r.RoomID,
		&r.RoomTitle,
		&p1ID,
		&p2ID,
		&r.Status,
		&r.BoardState,
		&r.IsWhiteTurn,
		&r.WhiteID,
		&r.BlackID,
		&r.ChatID,
	)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("sql:%s - %s", sql, err.Error()), zap.Error(err))
		return nil, fmt.Errorf("findRoomByPlayerIDs: %v", err)
	}

	r.Player1, err = GetUserByID(p1ID)
	if err != nil {
		utils.Logger.Error("GetUserByID:"+err.Error(), zap.Error(err))
		return nil, fmt.Errorf("findRoomByPlayerIDs: %v", err)
	}

	if p2ID != 0 {
		r.Player2, err = GetUserByID(p2ID)
		if err != nil {
			utils.Logger.Error("GetUserByID:"+err.Error(), zap.Error(err))
			return nil, fmt.Errorf("findRoomByPlayerIDs: %v", err)
		}
	}

	return &r, err
}

func UpdateRoom(r *Room) error {
	var p2ID *int64
	if r.Player2 != nil {
		p2ID = &r.Player2.ID
	}

	if err := r.Validate(); err != nil {
		utils.Logger.Info("UpdateRoom", zap.Any("room:", r))
		utils.Logger.Error(err.Error(), zap.Error(err))
		return err
	}
	utils.Logger.Debug("UpdateRoom", zap.Any("room:", &r), zap.Any("p2ID:", p2ID))
	sql := `UPDATE rooms
			SET player2_id = $1,
			    room_title = $2,
			    status = $3,
			    board_state = $4,
			    is_white_turn = $5,
			    white_id = $6,
			    black_id = $7,
			    chat_id = $8,
			    updated_at = NOW()
			WHERE room_id = $9;`
	_, err := Pool.Exec(context.Background(), sql,
		p2ID,
		r.RoomTitle,
		r.Status,
		r.BoardState,
		r.IsWhiteTurn,
		r.WhiteID,
		r.BlackID,
		r.ChatID,
		r.RoomID,
	)

	if err != nil {
		utils.Logger.Error(err.Error(), zap.Error(err))
		return fmt.Errorf("UpdateRoom: %v", err)
	}
	return nil
}

func GetPlayingRoomsForUser(userID int64) ([]Room, error) {
	sqlQuery := `
       SELECT room_id, room_title, board_state, is_white_turn, updated_at
--        SELECT room_id, room_title, player1_id, player2_id, status, board_state, is_white_turn, white_id, black_id, chat_id, created_at, updated_at
         FROM rooms
        WHERE status in ($1 ,$2)
          AND (player1_id = $3 OR player2_id = $3)
        ORDER BY updated_at DESC
    `
	rows, err := Pool.Query(context.Background(), sqlQuery, RoomStatusWaiting, RoomStatusPlaying, userID)
	if err != nil {
		return nil, fmt.Errorf("GetPlayingRoomsForUser: %v", err)
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var r Room
		err := rows.Scan(
			&r.RoomID,
			&r.RoomTitle,
			&r.BoardState,
			&r.IsWhiteTurn,
			&r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetPlayingRoomsForUser scan: %v", err)
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}
