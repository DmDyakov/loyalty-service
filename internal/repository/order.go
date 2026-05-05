package repository

import (
	"go.uber.org/zap"
)

type OrderRepository struct {
	db     DB
	logger *zap.Logger
}

func NewOrderRepository(db DB, l *zap.Logger) (*OrderRepository, error) {
	return &OrderRepository{
		db:     db,
		logger: l,
	}, nil
}
