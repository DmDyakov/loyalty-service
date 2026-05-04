package app

import (
	"fmt"
	"loyalty-service/internal/config"
	"loyalty-service/internal/logger"

	"go.uber.org/zap"
)

func Run(args []string) error {
	cfg, err := config.New(args)
	if err != nil {
		return fmt.Errorf("failed to create app config: %s", err)
	}

	if cfg == nil {
		return fmt.Errorf("app config is nil")
	}

	logger, err := logger.NewZapLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Server started",
		zap.String("url", cfg.RunAddress),
	)

	return nil
}
