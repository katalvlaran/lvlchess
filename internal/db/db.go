package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// InitDB инициализирует пул соединений и сохраняет в глобальную переменную Pool.
func InitDB() {
	dsn := "postgres://katalvlaran:kj916t4rf@localhost:5432/telega_chess"
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Ошибка pgxpool.ParseConfig: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), config.ConnString())
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Проверим соединение
	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}

	log.Println("Успешное подключение к PostgreSQL")
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
		rating    INT DEFAULT 1000,
		wins      INT DEFAULT 0,
		total_games INT DEFAULT 0
	);
	`
	_, err := Pool.Exec(context.Background(), schemaUsers)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы users: %v", err)
	}

	schemaRooms := `
	CREATE TABLE IF NOT EXISTS rooms (
		room_id    VARCHAR(36) PRIMARY KEY,
		player1_id BIGINT NOT NULL,
		player2_id BIGINT,
		status     VARCHAR(20) NOT NULL DEFAULT('waiting'), -- waiting/playing/finished
		board_state TEXT,
		white_id   BIGINT,
		black_id   BIGINT,
		chat_id    BIGINT, -- для группового чата
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
	    CONSTRAINT fk_p1 FOREIGN KEY(player1_id) REFERENCES users(id),
	    CONSTRAINT fk_p2 FOREIGN KEY(player2_id) REFERENCES users(id)
	);
	`
	_, err = Pool.Exec(context.Background(), schemaRooms)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы rooms: %v", err)
	}
}
