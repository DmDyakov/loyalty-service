package service

import (
	"context"
	"loyalty-service/internal/config"
	"loyalty-service/internal/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/order_repository.go -package=mocks . OrderRepository
type OrdersRepository interface {
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

type OrdersService struct {
	orderRepository OrdersRepository
	cfg             *config.Config
	logger          *zap.Logger
}

func NewOrdersService(repo OrdersRepository, cfg *config.Config, logger *zap.Logger) *OrdersService {
	return &OrdersService{
		orderRepository: repo,
		cfg:             cfg,
		logger:          logger,
	}
}

func (s *OrdersService) AddOrder(ctx context.Context, userID int, orderNumber string) error {
	panic("implement me")
}

func (s *OrdersService) GetUserOrders(ctx context.Context, userID int) ([]model.Order, error) {
	panic("implement me")
}
