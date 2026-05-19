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

//go:generate mockgen -destination=mocks/order_repository.go -package=mocks . OrdersRepository
type OrdersRepository interface {
	SaveOrder(ctx context.Context, userID int, orderNumber string) error
	FindOrdersByUser(ctx context.Context, userID int, limit int, offset int) ([]model.Order, error)
	FindOrdersByStatuses(ctx context.Context, statuses []string, limit int, offset int) ([]string, error)
	FindUserIDByOrderNumber(ctx context.Context, orderNumber string) (int, error)
	UpdateOrderInfo(
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

// NewOrdersService создает новый экземпляр сервиса заказов.
func NewOrdersService(repo OrdersRepository, cfg *config.Config, logger *zap.Logger) *OrdersService {
	return &OrdersService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

// AddOrder добавляет новый заказ пользователя с проверкой номера по алгоритму Луна.
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

// GetUserOrders возвращает список заказов пользователя с пагинацией.
func (s *OrdersService) GetUserOrders(ctx context.Context, userID int, limit int, offset int) ([]model.Order, error) {
	return s.repo.FindOrdersByUser(ctx, userID, limit, offset)
}

// UpdateOrderInfo обновляет статус и сумму начисления для заказа.
func (s *OrdersService) UpdateOrderInfo(ctx context.Context, orderNumber string, status model.OrderStatus, accrual decimal.Decimal) error {
	return s.repo.UpdateOrderInfo(ctx, orderNumber, status, accrual)
}
