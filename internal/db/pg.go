package db

import (
	"context"
	"fmt"

	"lvlchess/config"
	"lvlchess/internal/db/repositories"
	"lvlchess/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Global variables for the DB connection pool and repository references
var (
	Pool                   *pgxpool.Pool
	usersRepo              *repositories.UsersRepository
	roomsRepo              *repositories.RoomsRepository
	tournamentsRepo        *repositories.TournamentRepository
	tournamentSettingsRepo *repositories.TournamentSettingsRepository
)

/*
InitDB handles:
 1. Reading config (PgUser, PgPass, etc.) from config.Cfg
 2. Constructing the DSN (postgres://...)
 3. Creating the pgxpool connection
 4. Attempting a Ping() to confirm connectivity
 5. Setting up global repository objects (e.g. usersRepo, roomsRepo)
 6. Optionally calling initSchema() to create DB tables if they don't exist
*/
func InitDB() {
	// Build DSN from environment
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s%s/%s",
		config.Cfg.PgUser,
		config.Cfg.PgPass,
		config.Cfg.PgHost,
		config.Cfg.PgPort,
		config.Cfg.PgDbName,
	)

	dbCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		utils.Logger.Error("Could not parse DSN for PostgreSQL", zap.Error(err))
	}

	// Create a pool
	pool, err := pgxpool.New(context.Background(), dbCfg.ConnString())
	if err != nil {
		utils.Logger.Error("Unable to connect to PostgreSQL", zap.Error(err))
	}

	// Test the connection quickly
	err = pool.Ping(context.Background())
	if err != nil {
		utils.Logger.Error("PostgreSQL unreachable", zap.Error(err))
	}

	utils.Logger.Info("Successfully connected to PostgreSQL")

	Pool = pool

	// Initialize repository instances
	usersRepo = repositories.NewUsersRepository(Pool)
	roomsRepo = repositories.NewRoomsRepository(Pool)
	tournamentsRepo = repositories.NewTournamentRepository(Pool)
	tournamentSettingsRepo = repositories.NewTournamentSettingsRepository(Pool)

	// Run a basic schema creation script
	initSchema()
}

// GetUsersRepo returns the global UsersRepository singleton
func GetUsersRepo() *repositories.UsersRepository {
	return usersRepo
}

// GetRoomsRepo returns the global RoomsRepository singleton
func GetRoomsRepo() *repositories.RoomsRepository {
	return roomsRepo
}

// GetTournamentsRepo returns the global TournamentRepository singleton
func GetTournamentsRepo() *repositories.TournamentRepository {
	return tournamentsRepo
}

// GetTournamentSettingsRepo returns the global TournamentSettingsRepository singleton
func GetTournamentSettingsRepo() *repositories.TournamentSettingsRepository {
	return tournamentSettingsRepo
}

/*
initSchema():

	Simple "migration" approach.
	It creates required tables if they do not exist.
	You could also integrate a more robust migration tool (like goose or migrate).
*/
func initSchema() {
	schemaUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id        BIGINT UNIQUE,
		user_name  VARCHAR(255),
		first_name VARCHAR(255),
		chat_id   BIGINT DEFAULT(0),
	    current_room VARCHAR(36) NULL,
		rating    INT DEFAULT 1000,
		wins      INT DEFAULT 0,
		total_games INT DEFAULT 0
	);`
	if _, err := Pool.Exec(context.Background(), schemaUsers); err != nil {
		utils.Logger.Error("Error creating users table", zap.Error(err))
	}

	schemaRooms := `
	CREATE TABLE IF NOT EXISTS rooms (
		 room_id       VARCHAR(36) PRIMARY KEY,
		 room_title    TEXT,
		 player1_id    BIGINT NOT NULL,
		 player2_id    BIGINT,
		 status        VARCHAR(20) NOT NULL DEFAULT('waiting'), -- waiting/playing/finished
		 board_state   TEXT,
		 is_white_turn BOOLEAN NOT NULL DEFAULT true,
		 white_id      BIGINT,
		 black_id      BIGINT NULL,
		 chat_id       BIGINT, -- if referencing a group
		 created_at    TIMESTAMP DEFAULT NOW(),
		 updated_at    TIMESTAMP DEFAULT NOW(),
		 CONSTRAINT fk_p1 FOREIGN KEY(player1_id) REFERENCES users(id),
		 CONSTRAINT fk_p2 FOREIGN KEY(player2_id) REFERENCES users(id),
		 CONSTRAINT players_pair UNIQUE (player1_id, player2_id)
	 );

	ALTER TABLE users ADD CONSTRAINT fk_curr_room 
	    FOREIGN KEY(current_room) REFERENCES rooms(room_id);
	`
	if _, err := Pool.Exec(context.Background(), schemaRooms); err != nil {
		utils.Logger.Error("Error creating rooms table", zap.Error(err))
	}

	schemaTournaments := `
	CREATE TABLE IF NOT EXISTS tournaments (
	  id          VARCHAR(36) PRIMARY KEY,
	  title       VARCHAR(255),
	  prise       TEXT,
	  players     JSONB,   -- array of user IDs in JSON
	  status      INT DEFAULT 0,       -- 0=planned,1=active,2=finished
	  start_at    TIMESTAMP DEFAULT NOW(),
	  created_at  TIMESTAMP DEFAULT NOW(),
	  updated_at  TIMESTAMP DEFAULT NOW()
	);
	`
	if _, err := Pool.Exec(context.Background(), schemaTournaments); err != nil {
		utils.Logger.Error("Error creating tournaments table", zap.Error(err))
	}

	schemaTournamentSettings := `
	CREATE TABLE IF NOT EXISTS tournament_settings (
	  t_id   VARCHAR(36) NOT NULL,
	  r_id   VARCHAR(36) NOT NULL,
	  rank   INT DEFAULT 0,
	  status INT DEFAULT 0,
	  CONSTRAINT fk_tournament FOREIGN KEY (t_id) REFERENCES tournaments(id),
	  CONSTRAINT fk_room       FOREIGN KEY (r_id) REFERENCES rooms(room_id)
	);
	`
	if _, err := Pool.Exec(context.Background(), schemaTournamentSettings); err != nil {
		utils.Logger.Error("Error creating tournament_settings table", zap.Error(err))
	}
}
