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
	OwnerID  int64  `env:"OWNER_ID"`
	BotToken string `env:"BOT_TOKEN"`
	PgUser   string `env:"PG_USER" envDefault:"root"`
	PgPass   string `env:"PG_PASS" envDefault:""`
	PgHost   string `env:"PG_HOST" envDefault:"localhost"`
	PgPort   string `env:"PG_PORT" envDefault:":5432"`
	PgDbName string `env:"PG_DB_NAME" envDefault:"telgaram_chess"`
	// If needed, add more env fields here...
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
LoadConfig tries to:
 1. Load .env file if present.
 2. Parse environment variables into Cfg.
 3. Validate presence of mandatory fields.

Returns a pointer if you need, or logs the error.
*/
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Not necessarily fatal if .env is absent; you might rely on real env variables
		return new(Config), fmt.Errorf("env, load: %w", err)
	}
	if err := env.Parse(&Cfg); err != nil {
		return new(Config), fmt.Errorf("env, parse: %w", err)
	}
	if err := Cfg.Validate(); err != nil {
		return new(Config), fmt.Errorf("config, validate: %w", err)
	}
	return &Cfg, nil
}
