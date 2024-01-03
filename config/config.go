package config

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	Database   string   `env:"DATABASE" envDefault:"broker.db"`
	BrokerPort string   `env:"BROKER_PORT" envDefault:":8070"`
	NodePort   string   `env:"NODE_PORT" envDefault:":8071"`
	Nodes      []string `env:"NODES" envSeparator:" " envDefault:"localhost:8081 localhost:8091"`
}

func NewConfig() (Config, error) {
	var cfg Config

	godotenv.Load()

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
