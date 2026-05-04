package service

import (
	"context"
	"loyalty-service/internal/model"

	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination=mocks/balance_repo.go -package=mocks . BalanceRepo
type BalanceRepo interface {
	SaveWithdrawal(ctx context.Context, orderNumber string, userID int, amount decimal.Decimal) (*model.Withdrawal, error)
	FindWithdrawalsByUser(ctx context.Context, userID int) ([]model.Withdrawal, error)
	GetBalanceByUser(ctx context.Context, userID int) (decimal.Decimal, error)
}
