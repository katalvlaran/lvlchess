package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/joho/godotenv"
)

/*
Cfg is a global variable that holds the parsed configuration.
We use `env` and `godotenv` to load from .env file and environment variables.

Typically:
- BOT_TOKEN: Telegram Bot Token
- OWNER_ID: Optional numeric ID of the superuser
- PG_*: PostgreSQL credentials
*/
var Cfg Config

// Config fields mapped to environment variables
type Config struct {
	PGConfig
	TelegramConfig
}

// PGConfig fields mapped to environment variables
type PGConfig struct {
	PgUser   string `env:"PG_USER" envDefault:"root"`
	PgPass   string `env:"PG_PASS" envDefault:""`
	PgHost   string `env:"PG_HOST" envDefault:"localhost"`
	PgPort   string `env:"PG_PORT" envDefault:":5432"`
	PgDbName string `env:"PG_DB_NAME" envDefault:"telgaram_chess"`
	// If needed, add more env fields here...
}

// TelegramConfig fields mapped to environment variables
type TelegramConfig struct {
	OwnerID       int64  `env:"OWNER_ID"`
	BotToken      string `env:"BOT_TOKEN"`
	GameShortName string `env:"GAME_SHORT_NAME"`
	GameURL       string `env:"GAME_URL"`
}

/*
Validate ensures that critical fields are set (e.g., BotToken).
By default, Go-ozzo-validation or caarlos0/env can parse them from environment.
*/
func (c *Config) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.BotToken, validation.Required),
		validation.Field(&c.PgUser, validation.Required),
		validation.Field(&c.PgPass, validation.Required),
		validation.Field(&c.PgHost, validation.Required),
		validation.Field(&c.PgPort, validation.Required),
		validation.Field(&c.PgDbName, validation.Required),
	)
}

/*
Validate ensures that critical fields are set (e.g., BotToken).
By default, Go-ozzo-validation or caarlos0/env can parse them from environment.
*/
func (c *PGConfig) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.PgUser, validation.Required),
		validation.Field(&c.PgPass, validation.Required),
		validation.Field(&c.PgHost, validation.Required),
		validation.Field(&c.PgPort, validation.Required),
		validation.Field(&c.PgDbName, validation.Required),
	)
}

/*
Validate ensures that critical fields are set (e.g., BotToken).
By default, Go-ozzo-validation or caarlos0/env can parse them from environment.
*/
func (c *TelegramConfig) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.BotToken, validation.Required),
		validation.Field(&c.GameShortName, validation.Required),
		validation.Field(&c.GameURL, validation.Required),
	)
}

/*
LoadConfig tries to:
 1. Load .env file if present.
 2. Parse environment variables into Cfg.
 3. Validate presence of mandatory fields.

Returns a pointer if you need, or logs the error.
*/

func LoadConfig() error {
	// Attempt to load from .env file if present. It's fine if it doesn't exist.
	if err := godotenv.Load(); err != nil {
		// Not necessarily fatal; we can log it. But let's just wrap it:
		fmt.Printf("Warning: .env load error: %v\n", err)
	}
	if err := env.Parse(&Cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := Cfg.PGConfig.Validate(); err != nil {
		return fmt.Errorf("PGConfig config invalid: %w", err)
	}
	if err := Cfg.TelegramConfig.Validate(); err != nil {
		return fmt.Errorf("TelegramConfig config invalid: %w", err)
	}
	if err := Cfg.Validate(); err != nil {
		return fmt.Errorf("top-level config invalid: %w", err)
	}

	return nil
}
