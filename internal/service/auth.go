package service

import (
	"context"
	"loyalty-service/internal/model"
)

//go:generate mockgen -destination=mocks/user_repo.go -package=mocks . UserRepo
type UserRepo interface {
	SaveUser(ctx context.Context, login string, passwordHash string) (*model.User, error)
	GetUserByID(ctx context.Context, userID int) (*model.User, error)
}
