package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress     string `env:"RUN_ADDRESS"`
	DatabaseDSN    string `env:"DATABASE_URI"`
	AccrualBaseURL string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AppEnv         string `env:"APP_ENV"`
}

const (
	defaultRunAddress     = "localhost:8080"
	defaultDatabaseDSN    = ""
	defaultAccrualBaseURL = ""
	defaultAppEnv         = "dev"
)

func New(flags []string) (*Config, error) {
	cfg := &Config{
		RunAddress:     defaultRunAddress,
		DatabaseDSN:    defaultDatabaseDSN,
		AccrualBaseURL: defaultAccrualBaseURL,
		AppEnv:         defaultAppEnv,
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

	switch cfg.AppEnv {
	case "prod", "dev":
	default:
		return fmt.Errorf("invalid APP_ENV: %s (expected prod or dev)", cfg.AppEnv)
	}
	return nil
}
