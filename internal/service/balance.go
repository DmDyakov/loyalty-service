package service

import (
	"context"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"loyalty-service/pkg/luhn"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/balance_repository.go -package=mocks . BalanceRepository
type BalanceRepository interface {
	GetAccrualSumByUser(ctx context.Context, userID int) (*decimal.Decimal, error)
	GetWithdrawnSumByUser(ctx context.Context, userID int) (*decimal.Decimal, error)
	SaveWithdrawal(ctx context.Context, userID int, orderNumber string, sum decimal.Decimal) error
	FindWithdrawalsByUser(ctx context.Context, userID int, limit int, offset int) ([]model.Withdrawal, error)
}

type BalanceService struct {
	repo   BalanceRepository
	cfg    *config.Config
	logger *zap.Logger
}

// NewBalanceService создает новый экземпляр сервиса баланса.
func NewBalanceService(repo BalanceRepository, cfg *config.Config, logger *zap.Logger) *BalanceService {
	return &BalanceService{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}
}

// GetUserBalance возвращает текущий баланс пользователя (начисления минус списания).
func (s *BalanceService) GetUserBalance(ctx context.Context, userID int) (*model.Balance, error) {
	accrualSum, err := s.repo.GetAccrualSumByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	withdrawnSum, err := s.repo.GetWithdrawnSumByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentBalance := decimal.Zero

	if accrualSum != nil {
		currentBalance = currentBalance.Add(*accrualSum)
	}

	if withdrawnSum != nil {
		currentBalance = currentBalance.Sub(*withdrawnSum)
	}

	return &model.Balance{
		Current:   &currentBalance,
		Withdrawn: withdrawnSum,
	}, nil
}

// Withdraw выполняет списание баллов лояльности в счёт оплаты заказа.
func (s *BalanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum decimal.Decimal) error {
	ok := luhn.Valid(orderNumber)
	if !ok {
		return errs.ErrInvalidOrderNumber
	}

	return s.repo.SaveWithdrawal(ctx, userID, orderNumber, sum)
}

// GetUserWithdrawals возвращает историю списаний пользователя с пагинацией.
func (s *BalanceService) GetUserWithdrawals(ctx context.Context, userID int, limit int, offset int) ([]model.Withdrawal, error) {

	return s.repo.FindWithdrawalsByUser(ctx, userID, limit, offset)
}
