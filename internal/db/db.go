package db

import (
	"context"
	"fmt"

	"telega_chess/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"telega_chess/internal/utils"
)

var Pool *pgxpool.Pool

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

	// Выполним миграцию (упрощённый вариант):
	initSchema()
}

// initSchema - создаём таблицы, если не созданы
func initSchema() {
	schemaUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id        BIGINT UNIQUE,
		user_name  VARCHAR(255),
		first_name VARCHAR(255),
		chat_id   BIGINT DEFAULT(0),   -- 0 если ещё не знаем
	    current_room VARCHAR(36) NULL,
		rating    INT DEFAULT 1000,
		wins      INT DEFAULT 0,
		total_games INT DEFAULT 0
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
}
