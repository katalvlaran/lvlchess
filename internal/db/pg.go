package db

import (
	"context"
	"fmt"

	"telega_chess/config"
	"telega_chess/internal/db/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"telega_chess/internal/utils"
)

var Pool *pgxpool.Pool

var (
	usersRepo              *repositories.UsersRepository
	roomsRepo              *repositories.RoomsRepository
	tournamentsRepo        *repositories.TournamentRepository
	tournamentSettingsRepo *repositories.TournamentSettingsRepository
)

// InitDB инициализирует пул соединений и сохраняет в глобальную переменную Pool.
func InitDB() {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s%s/%s",
		config.Cfg.PgUser,
		config.Cfg.PgPass,
		config.Cfg.PgHost,
		config.Cfg.PgPort,
		config.Cfg.PgDbName)
	dbCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		utils.Logger.Error("😖 Ошибка pgxpool.ParseConfig: 💀"+err.Error(), zap.Error(err))
	}

	pool, err := pgxpool.New(context.Background(), dbCfg.ConnString())
	if err != nil {
		utils.Logger.Error("😖 Ошибка подключения к БД: 💀"+err.Error(), zap.Error(err))
	}

	// Проверим соединение
	err = pool.Ping(context.Background())
	if err != nil {
		utils.Logger.Error("😖 БД недоступна 🤷"+err.Error(), zap.Error(err))
	}

	utils.Logger.Info("🦾 Успешное подключение к PostgreSQL 🗄")
	Pool = pool

	// Создаём экземпляры репозиториев
	usersRepo = repositories.NewUsersRepository(Pool)
	roomsRepo = repositories.NewRoomsRepository(Pool)
	tournamentsRepo = repositories.NewTournamentRepository(Pool)
	tournamentSettingsRepo = repositories.NewTournamentSettingsRepository(Pool)

	// Выполним миграцию (упрощённый вариант):
	initSchema()
}

// геттеры
func GetUsersRepo() *repositories.UsersRepository {
	return usersRepo
}

func GetRoomsRepo() *repositories.RoomsRepository {
	return roomsRepo
}

func GetTournamentsRepo() *repositories.TournamentRepository {
	return tournamentsRepo
}

func GetTournamentSettingsRepo() *repositories.TournamentSettingsRepository {
	return tournamentSettingsRepo
}

type Kline struct {
	OpenTime  int64   `json:"openTime"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	CloseTime int64   `json:"closeTime"`
}

// initSchema - создаём таблицы, если не созданы
func initSchema() {
	schemaUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT UNIQUE,
		OpenTime BIGINT,
		Open FLOAT,
		High FLOAT,
	    Low FLOAT,
		Close FLOAT,
		Volume FLOAT,
		CloseTime INT
	);`
	_, err := Pool.Exec(context.Background(), schemaUsers)
	if err != nil {
		utils.Logger.Error("😖 Ошибка создания таблицы users: 💀"+err.Error(), zap.Error(err))
	}

	schemaRooms := `
	CREATE TABLE IF NOT EXISTS rooms (
		 room_id       VARCHAR(36) PRIMARY KEY,
		 room_title	  TEXT,
		 player1_id    BIGINT NOT NULL,
		 player2_id    BIGINT,
		 status        VARCHAR(20) NOT NULL DEFAULT('waiting'), -- waiting/playing/finished
		 board_state   TEXT,
		 is_white_turn BOOLEAN NOT NULL DEFAULT true,
		 white_id      BIGINT,
		 black_id      BIGINT NULL,
		 chat_id       BIGINT, -- для группового чата
		 created_at    TIMESTAMP DEFAULT NOW(),
		 updated_at    TIMESTAMP DEFAULT NOW(),
		 CONSTRAINT fk_p1 FOREIGN KEY(player1_id) REFERENCES users(id),
		 CONSTRAINT fk_p2 FOREIGN KEY(player2_id) REFERENCES users(id),
		 CONSTRAINT players_pair UNIQUE (player1_id, player2_id)
	 );

	ALTER TABLE users ADD CONSTRAINT fk_curr_room FOREIGN KEY(current_room) REFERENCES rooms(room_id);`
	_, err = Pool.Exec(context.Background(), schemaRooms)
	if err != nil {
		utils.Logger.Error("😖 Ошибка создания таблицы rooms: 💀"+err.Error(), zap.Error(err))
	}

	schemaTournaments := `
	CREATE TABLE IF NOT EXISTS tournaments (
	  id          VARCHAR(36) PRIMARY KEY,
	  title       VARCHAR(255),
	  prise       TEXT,
	  players     JSONB,   -- массив ID-шников пользователей: [123, 456, ...]
	  status      INT DEFAULT 0,       -- 0=planned,1=active,2=finished
	  start_at    TIMESTAMP DEFAULT NOW(),
	  created_at  TIMESTAMP DEFAULT NOW(),
	  updated_at  TIMESTAMP DEFAULT NOW()
	);

	`
	_, err = Pool.Exec(context.Background(), schemaTournaments)
	if err != nil {
		utils.Logger.Error("😖 Ошибка создания таблицы tournaments: 💀"+err.Error(), zap.Error(err))
	}

	schemaTournamentSettings := `
	CREATE TABLE IF NOT EXISTS tournament_settings (
	  t_id   VARCHAR(36) NOT NULL,
	  r_id   VARCHAR(36) NOT NULL,
	  rank   INT DEFAULT 0,
	  status INT DEFAULT 0,
	  CONSTRAINT fk_tournament FOREIGN KEY (t_id) REFERENCES tournaments (id),
	  CONSTRAINT fk_room      FOREIGN KEY (r_id) REFERENCES rooms (room_id)
	);

	`
	_, err = Pool.Exec(context.Background(), schemaTournamentSettings)
	if err != nil {
		utils.Logger.Error("😖 Ошибка создания таблицы tournament_settings: 💀"+err.Error(), zap.Error(err))
	}
}
