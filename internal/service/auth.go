package service

import (
	"context"
	"loyalty-service/internal/model"
)

//go:generate mockgen -destination=mocks/user_repository.go -package=mocks . UserRepository
type UserRepository interface {
	SaveUser(ctx context.Context, login string, passwordHash string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int) (*model.User, error)
}
