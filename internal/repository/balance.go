package repository

import (
	"go.uber.org/zap"
)

type BalanceRepository struct {
	db     DB
	logger *zap.Logger
}

func NewBalanceRepository(db DB, l *zap.Logger) (*BalanceRepository, error) {
	return &BalanceRepository{
		db:     db,
		logger: l,
	}, nil
}
