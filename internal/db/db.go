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
		log.Fatalf("Ошибка парсинга DSN: %v", err)
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

	// Дополнительно: создать таблицу, если не создана
	createRoomsTable()
}

// createRoomsTable - пример (упрощённо, можно вынести в миграции)
func createRoomsTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS rooms (
		room_id    VARCHAR(36) PRIMARY KEY,
		player1_id BIGINT NOT NULL,
		player2_id BIGINT,
		status     VARCHAR(20) NOT NULL,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	`
	_, err := Pool.Exec(context.Background(), sql)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы rooms: %v", err)
	}
}
