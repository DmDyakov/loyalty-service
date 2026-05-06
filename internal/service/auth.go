package service

import (
	"context"
	"loyalty-service/internal/config"
	"loyalty-service/internal/model"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/user_repository.go -package=mocks . UserRepository
type UserRepository interface {
	SaveUser(ctx context.Context, login string, passwordHash string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int) (*model.User, error)
}

type AuthService struct {
	userRepository UserRepository
	cfg            *config.Config
	logger         *zap.Logger
}

func NewAuthService(repo UserRepository, cfg *config.Config, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepository: repo,
		cfg:            cfg,
		logger:         logger,
	}
}
