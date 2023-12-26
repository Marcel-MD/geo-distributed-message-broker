package config

import (
	"log/slog"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string `env:"PORT" envDefault:":8080"`
	Env  string `env:"ENV" envDefault:"dev"`
}

func NewConfig() (Config, error) {
	var cfg Config

	err := godotenv.Load()
	if err != nil {
		slog.Error(err.Error())
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
