package config

import (
	"flag"
	"fmt"

	env "github.com/caarlos0/env/v11"
)

const (
	DefaultRunAddress           = "localhost:8000"
	DefaultAccrualSystemAddress = "http://localhost:8080"
	DefaultPublicKeyPath        = "public.pem"
	DefaultPrivateKeyPath       = "private.pem"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	PublicKeyPath        string `env:"PUBLIC_KEY_PATH"`
	PrivateKeyPath       string `env:"PRIVATE_KEY_PATH"`
}

func Init() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", DefaultRunAddress, "run address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database uri")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", DefaultAccrualSystemAddress, "accrual system address")
	flag.StringVar(&cfg.PublicKeyPath, "p", DefaultPublicKeyPath, "public key path")
	flag.StringVar(&cfg.PrivateKeyPath, "s", DefaultPrivateKeyPath, "private key path")

	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("can not parse env: %w", err)
	}

	return &cfg, nil
}
