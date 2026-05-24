package service

import (
	"context"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestService_AddOrder(t *testing.T) {
	const (
		validUserID        = 1
		validOrderNumber   = "49927398716"
		invalidOrderNumber = "invalid_order_number"
	)

	t.Run("success", func(t *testing.T) {
		s, repo := setupTestOrdersService(t)
		repo.EXPECT().
			SaveOrder(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		err := s.AddOrder(context.Background(), validUserID, validOrderNumber)
		assert.NoError(t, err)
	})

	t.Run("invalid order number", func(t *testing.T) {
		s, repo := setupTestOrdersService(t)
		repo.EXPECT().
			SaveOrder(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		err := s.AddOrder(context.Background(), validUserID, invalidOrderNumber)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidOrderNumber)
	})

	t.Run("order has been uploaded by another user", func(t *testing.T) {
		s, repo := setupTestOrdersService(t)
		repo.EXPECT().
			SaveOrder(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errs.ErrOrderAlreadyExists)

		repo.EXPECT().
			FindUserIDByOrderNumber(gomock.Any(), gomock.Any()).
			Return(2, nil)
		err := s.AddOrder(context.Background(), 1, validOrderNumber)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrOrderUploadedByAnother)
	})

	t.Run("order has been uploaded by current user", func(t *testing.T) {
		s, repo := setupTestOrdersService(t)
		repo.EXPECT().
			SaveOrder(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errs.ErrOrderAlreadyExists)

		repo.EXPECT().
			FindUserIDByOrderNumber(gomock.Any(), gomock.Any()).
			Return(1, nil)
		err := s.AddOrder(context.Background(), 1, validOrderNumber)

		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrOrderAlreadyExists)
	})

}

func setupTestOrdersService(t *testing.T) (*OrdersService, *mocks.MockOrdersRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockOrdersRepository := mocks.NewMockOrdersRepository(ctrl)
	cfg := &config.Config{}
	logger := zap.NewNop()
	s := NewOrdersService(mockOrdersRepository, cfg, logger)
	return s, mockOrdersRepository
}
