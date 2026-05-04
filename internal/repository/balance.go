package repository

import (
	"go.uber.org/zap"
)

type BalanceRepo struct {
	db     DB
	logger *zap.Logger
}

func NewBalanceRepo(db DB, l *zap.Logger) (*BalanceRepo, error) {
	return &BalanceRepo{
		db:     db,
		logger: l,
	}, nil
}
