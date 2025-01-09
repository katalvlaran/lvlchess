package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/joho/godotenv"
)

var Cfg Config

type Config struct {
	OwnerID  int64  `env:"OWNER_ID"`
	BotToken string `env:"BOT_TOKEN"`
	PgUser   string `env:"PG_USER" envDefault:"root"`
	PgPass   string `env:"PG_PASS" envDefault:""`
	PgHost   string `env:"PG_HOST" envDefault:"localhost"`
	PgPort   string `env:"PG_PORT" envDefault:":5432"`
	PgDbName string `env:"PG_DB_NAME" envDefault:"telgaram_chess"`
	//SomeEnv string `env:"SOME_ENV" envDefault:"default value"``
}

func (c *Config) Validate() error {
	return validation.ValidateStruct(c,
		//validation.Field(&c.OwnerID, validation.Required),
		validation.Field(&c.BotToken, validation.Required),
		validation.Field(&c.PgUser, validation.Required),
		validation.Field(&c.PgPass, validation.Required),
		validation.Field(&c.PgHost, validation.Required),
		validation.Field(&c.PgPort, validation.Required),
		validation.Field(&c.PgDbName, validation.Required),
	)
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
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
