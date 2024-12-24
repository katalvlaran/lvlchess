package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	Telegram struct {
		Token   string `mapstructure:"token"`
		Timeout int    `mapstructure:"timeout"`
		Debug   bool   `mapstructure:"debug"`
	} `mapstructure:"telegram"`

	Game struct {
		DefaultTimeLimit int   `mapstructure:"default_time_limit"`
		TimeControls     []int `mapstructure:"time_controls"`
	} `mapstructure:"game"`

	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
		File   string `mapstructure:"file"`
	} `mapstructure:"logging"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
