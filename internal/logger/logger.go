package logger

import (
	"loyalty-service/internal/config"

	"go.uber.org/zap"
)

func NewZapLogger(cfg *config.Config) (*zap.Logger, error) {
	switch cfg.AppEnv {
	case "prod":
		return zap.NewProduction()
	default:
		return zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	}
}
