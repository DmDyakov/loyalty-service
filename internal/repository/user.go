package repository

import (
	"go.uber.org/zap"
)

type UserRepository struct {
	db     DB
	logger *zap.Logger
}

func NewUserRepository(db DB, l *zap.Logger) (*UserRepository, error) {
	return &UserRepository{
		db:     db,
		logger: l,
	}, nil
}
