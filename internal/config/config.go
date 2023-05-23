package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	GRPCAddress string `env:"KEEPER_ADDRESS" envDefault:":3200"`
}

func GetConfig() (*Config, error) {

	cfg := &Config{}

	flag.StringVar(&cfg.GRPCAddress, "g", cfg.GRPCAddress, "KEEPER_ADDRESS")
	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
