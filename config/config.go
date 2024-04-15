package config

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Database   string   `env:"DATABASE" envDefault:"broker.db"`
	BrokerPort string   `env:"BROKER_PORT" envDefault:":8070"`
	NodePort   string   `env:"NODE_PORT" envDefault:":8071"`
	Nodes      []string `env:"NODES" envSeparator:"," envDefault:""`
	Username   string   `env:"USERNAME" envDefault:"admin"`
	Password   string   `env:"PASSWORD" envDefault:"password"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
