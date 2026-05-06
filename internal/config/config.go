package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress     string        `env:"RUN_ADDRESS"`
	DatabaseDSN    string        `env:"DATABASE_URI"`
	AccrualBaseURL string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AppEnv         string        `env:"APP_ENV"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT"`
}

const (
	defaultRunAddress     = "localhost:8080"
	defaultDatabaseDSN    = ""
	defaultAccrualBaseURL = ""
	defaultAppEnv         = "dev"
	defaultRequestTimeout = 10 * time.Second
)

func New(flags []string) (*Config, error) {
	cfg := &Config{
		RunAddress:     defaultRunAddress,
		DatabaseDSN:    defaultDatabaseDSN,
		AccrualBaseURL: defaultAccrualBaseURL,
		AppEnv:         defaultAppEnv,
		RequestTimeout: defaultRequestTimeout,
	}

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Warning: error loading .env file: %v", err)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("gofermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "address and port to run server")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database DSN")
	fs.StringVar(&cfg.AccrualBaseURL, "r", cfg.AccrualBaseURL, "Accrual system Base URL")
	fs.StringVar(&cfg.AppEnv, "e", cfg.AppEnv, "application environment (prod, dev)")
	fs.DurationVar(&cfg.RequestTimeout, "t", cfg.RequestTimeout, "request timeout")

	if err := fs.Parse(flags); err != nil {
		return nil, err
	}

	if err := cfg.validateConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) validateConfig() error {
	if cfg.RunAddress == "" {
		return errors.New("run Address can not be empty")
	}
	if cfg.DatabaseDSN == "" {
		return errors.New("database DSN can not be empty")
	}
	if cfg.AccrualBaseURL == "" {
		return errors.New("accrual Base URL can not be empty")
	}
	if cfg.RequestTimeout <= 0 {
		return errors.New("request timeout can be > 0")
	}

	switch cfg.AppEnv {
	case "prod", "dev":
	default:
		return fmt.Errorf("invalid APP_ENV: %s (expected prod or dev)", cfg.AppEnv)
	}
	return nil
}
