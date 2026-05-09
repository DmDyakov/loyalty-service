package repository

import (
	"context"
	"loyalty-service/internal/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type OrdersRepository struct {
	db     *DB
	logger *zap.Logger
}

func NewOrdersRepository(db *DB, l *zap.Logger) *OrdersRepository {
	return &OrdersRepository{
		db:     db,
		logger: l,
	}
}

func (r *OrdersRepository) SaveOrder(ctx context.Context, orderNumber string, userID int) (*model.Order, error) {
	panic("not implemented")
}

func (r *OrdersRepository) FindOrdersByUser(ctx context.Context, userID int) ([]model.Order, error) {
	panic("not implemented")
}

func (r *OrdersRepository) FindOrdersByStatuses(ctx context.Context, statuses []model.OrderStatus) {
	panic("not implemented")
}

func (r *OrdersRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status model.OrderStatus, accrual decimal.Decimal) (*model.Order, error) {
	panic("not implemented")
}
