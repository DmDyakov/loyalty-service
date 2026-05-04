package repository

import (
	"go.uber.org/zap"
)

type OrderRepo struct {
	db     DB
	logger *zap.Logger
}

func NewOrderRepo(db DB, l *zap.Logger) (*OrderRepo, error) {
	return &OrderRepo{
		db:     db,
		logger: l,
	}, nil
}
