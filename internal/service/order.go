package service

import (
	"context"
	"loyalty-service/internal/model"

	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination=mocks/order_repo.go -package=mocks . OrderRepo
type OrderRepo interface {
	SaveOrder(ctx context.Context, orderNumber string, userID int) (*model.Order, error)
	FindOrdersByUser(ctx context.Context, userID int) ([]model.Order, error)
	FindOrdersByStatuses(ctx context.Context, statuses []model.OrderStatus)
	UpdateOrderStatus(
		ctx context.Context,
		orderNumber string,
		status model.OrderStatus,
		accrual decimal.Decimal,
	) (*model.Order, error)
}
