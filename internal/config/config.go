// Package config управляет конфигурацией приложения через флаги и переменные окружения.
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
	RunAddress      string        `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseDSN     string        `env:"DATABASE_URI" envDefault:""`
	AccrualBaseURL  string        `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:""`
	AppEnv          string        `env:"APP_ENV" envDefault:"dev"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"10s"`
	JWTSecret       string        `env:"JWT_SECRET" envDefault:"yu2ikdafk3sdfh52skdfh"`
	JWTExpiry       time.Duration `env:"JWT_EXPIRY" envDefault:"1h"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
	MaxResults      int           `env:"MAX_RESULTS" envDefault:"100"`
	PollingInterval time.Duration `env:"POLLING_INTERVAL" envDefault:"10s"`
	RequestInterval time.Duration `env:"REQUEST_INTERVAL" envDefault:"200ms"`
}

func New(flags []string) (*Config, error) {
	cfg := &Config{}

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
	if cfg.RequestTimeout <= 0 {
		return errors.New("request timeout can be > 0")
	}
	if cfg.JWTSecret == "" {
		return errors.New("JWT secret can not be empty")
	}

	if cfg.JWTExpiry <= 0 {
		return errors.New("JWT expiry can be > 0")
	}

	if cfg.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be > 0")
	}

	if cfg.MaxResults <= 0 {
		return errors.New("max results must be > 0")
	}

	if cfg.PollingInterval <= 0 {
		return errors.New("polling interval must be > 0")
	}

	if cfg.RequestInterval <= 0 {
		return errors.New("request interval must be > 0")
	}

	switch cfg.AppEnv {
	case "prod":
		if len(cfg.JWTSecret) < 32 {
			return errors.New("JWT secret must be at least 32 characters in production")
		}
	case "dev":
	default:
		return fmt.Errorf("invalid APP_ENV: %s (expected prod or dev)", cfg.AppEnv)
	}

	return nil
}
