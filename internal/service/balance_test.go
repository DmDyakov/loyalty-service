package service

import (
	"context"
	"errors"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"loyalty-service/internal/service/mocks"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestBalanceService_GetUserBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		accrualSum := decimal.NewFromFloat(500.5)
		withdrawnSum := decimal.NewFromFloat(42.0)
		expectedCurrent := decimal.NewFromFloat(458.5)

		repo.EXPECT().
			GetAccrualSumByUser(gomock.Any(), 1).
			Return(&accrualSum, nil)

		repo.EXPECT().
			GetWithdrawnSumByUser(gomock.Any(), 1).
			Return(&withdrawnSum, nil)

		balance, err := s.GetUserBalance(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, balance)
		assert.True(t, expectedCurrent.Equal(*balance.Current))
		assert.True(t, withdrawnSum.Equal(*balance.Withdrawn))
	})

	t.Run("accrual sum error", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			GetAccrualSumByUser(gomock.Any(), 1).
			Return(nil, errors.New("db error"))

		balance, err := s.GetUserBalance(context.Background(), 1)

		require.Error(t, err)
		assert.Nil(t, balance)
	})

	t.Run("withdrawn sum error", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		accrualSum := decimal.NewFromFloat(500.5)

		repo.EXPECT().
			GetAccrualSumByUser(gomock.Any(), 1).
			Return(&accrualSum, nil)

		repo.EXPECT().
			GetWithdrawnSumByUser(gomock.Any(), 1).
			Return(nil, errors.New("db error"))

		balance, err := s.GetUserBalance(context.Background(), 1)

		require.Error(t, err)
		assert.Nil(t, balance)
	})

	t.Run("zero balance", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		zero := decimal.Zero

		repo.EXPECT().
			GetAccrualSumByUser(gomock.Any(), 1).
			Return(&zero, nil)

		repo.EXPECT().
			GetWithdrawnSumByUser(gomock.Any(), 1).
			Return(&zero, nil)

		balance, err := s.GetUserBalance(context.Background(), 1)

		require.NoError(t, err)
		assert.NotNil(t, balance)
		assert.True(t, decimal.Zero.Equal(*balance.Current))
		assert.True(t, decimal.Zero.Equal(*balance.Withdrawn))
	})
}

func TestBalanceService_Withdraw(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			SaveWithdrawal(gomock.Any(), 1, "2377225624", decimal.NewFromInt(100)).
			Return(nil)

		err := s.Withdraw(context.Background(), 1, "2377225624", decimal.NewFromInt(100))
		assert.NoError(t, err)
	})

	t.Run("invalid order number", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			SaveWithdrawal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		err := s.Withdraw(context.Background(), 1, "invalid", decimal.NewFromInt(100))
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidOrderNumber)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			SaveWithdrawal(gomock.Any(), 1, "2377225624", decimal.NewFromInt(1000)).
			Return(errs.ErrInsufficientFunds)

		err := s.Withdraw(context.Background(), 1, "2377225624", decimal.NewFromInt(1000))
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInsufficientFunds)
	})

	t.Run("db error", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			SaveWithdrawal(gomock.Any(), 1, "2377225624", decimal.NewFromInt(100)).
			Return(errors.New("db error"))

		err := s.Withdraw(context.Background(), 1, "2377225624", decimal.NewFromInt(100))
		require.Error(t, err)
	})
}

func TestBalanceService_GetUserWithdrawals(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		expected := []model.Withdrawal{
			{
				OrderNumber: "2377225624",
				Sum:         decimal.NewFromInt(500),
			},
		}

		repo.EXPECT().
			FindWithdrawalsByUser(gomock.Any(), 1, 10, 0).
			Return(expected, nil)

		result, err := s.GetUserWithdrawals(context.Background(), 1, 10, 0)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "2377225624", result[0].OrderNumber)
	})

	t.Run("empty result", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			FindWithdrawalsByUser(gomock.Any(), 1, 10, 0).
			Return([]model.Withdrawal{}, nil)

		result, err := s.GetUserWithdrawals(context.Background(), 1, 10, 0)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("db error", func(t *testing.T) {
		s, repo := setupTestBalanceService(t)

		repo.EXPECT().
			FindWithdrawalsByUser(gomock.Any(), 1, 10, 0).
			Return(nil, errors.New("db error"))

		result, err := s.GetUserWithdrawals(context.Background(), 1, 10, 0)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func setupTestBalanceService(t *testing.T) (*BalanceService, *mocks.MockBalanceRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockBalanceRepository := mocks.NewMockBalanceRepository(ctrl)
	cfg := &config.Config{}
	logger := zap.NewNop()
	s := NewBalanceService(mockBalanceRepository, cfg, logger)
	return s, mockBalanceRepository
}
