package service

import (
	"context"
	"errors"
	"fmt"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"loyalty-service/pkg/luhn"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/order_repository.go -package=mocks . OrderRepository
type OrdersRepository interface {
	SaveOrder(ctx context.Context, userID int, orderNumber string) error
	FindOrdersByUser(ctx context.Context, userID int, limit int, offset int) ([]model.Order, error)
	FindOrdersByStatuses(ctx context.Context, statuses []string, limit int, offset int) ([]string, error)
	FindUserIDByOrderNumber(ctx context.Context, orderNumber string) (int, error)
	UpdateOrderStatus(
		ctx context.Context,
		orderNumber string,
		status model.OrderStatus,
		accrual decimal.Decimal,
	) error
}

type OrdersService struct {
	repo   OrdersRepository
	cfg    *config.Config
	logger *zap.Logger
}

func NewOrdersService(repo OrdersRepository, cfg *config.Config, logger *zap.Logger) *OrdersService {
	return &OrdersService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *OrdersService) AddOrder(ctx context.Context, userID int, orderNumber string) error {
	ok := luhn.Valid(orderNumber)
	if !ok {
		return errs.ErrInvalidOrderNumber
	}

	if err := s.repo.SaveOrder(ctx, userID, orderNumber); err != nil {
		switch {
		case errors.Is(err, errs.ErrOrderAlreadyExists):
			id, err := s.repo.FindUserIDByOrderNumber(ctx, orderNumber)
			if err != nil {
				return fmt.Errorf("failed to find owner of order %s: %w", orderNumber, err)
			}
			if id != userID {
				return errs.ErrOrderUploadedByAnother
			}
			return errs.ErrOrderAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func (s *OrdersService) GetUserOrders(ctx context.Context, userID int, limit int, offset int) ([]model.Order, error) {
	return s.repo.FindOrdersByUser(ctx, userID, limit, offset)
}

func (s *OrdersService) UpdateOrderStatus(ctx context.Context, orderNumber string, status model.OrderStatus, accrual decimal.Decimal) error {
	return s.repo.UpdateOrderStatus(ctx, orderNumber, status, accrual)
}
