package repository

import (
	"go.uber.org/zap"
)

type UserRepo struct {
	db     DB
	logger *zap.Logger
}

func NewUserRepo(db DB, l *zap.Logger) (*UserRepo, error) {
	return &UserRepo{
		db:     db,
		logger: l,
	}, nil
}
