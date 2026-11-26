package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type UserIdentifier string
type AuthSecretKey string

const (
	UserIDKey UserIdentifier = "userId"
)

type Config struct {
	Address              string        `env:"RUN_ADDRESS"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            AuthSecretKey `env:"SECRET_KEY"`
	Database             PostgresConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		Address: "0.0.0.0:8080",
		Database: PostgresConfig{
			DSN:               "",
			ConnectionTimeout: ConnectionTimeout,
		},
		AccrualSystemAddress: "",
	}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "application address")
	flag.StringVar(&cfg.Database.DSN, "d", cfg.Database.DSN, "database DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	flag.Parse()

	return cfg, nil
}
