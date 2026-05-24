package accrual

import (
	"context"
	"errors"
	"testing"
	"time"

	accrualclient "loyalty-service/internal/client/accrual"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"loyalty-service/internal/worker/accrual/mocks"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestPoller_Start(t *testing.T) {
	t.Run("stops on context cancel", func(t *testing.T) {
		poller, repo := setupTestPoller(t, 10*time.Millisecond, 10*time.Millisecond)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{}, nil).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			poller.Start(ctx)
			close(done)
		}()

		time.Sleep(50 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("poller did not stop")
		}
	})
}

func TestPoller_ProcessBatch(t *testing.T) {
	t.Run("processes orders successfully", func(t *testing.T) {
		poller, repo, client := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(),
				[]string{string(model.OrderNew), string(model.OrderProcessing)},
				defaultBatchSize, 0).
			Return([]string{"12345678903", "49927398716"}, nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "12345678903").
			Return(&accrualclient.OrderInfo{
				Number:  "12345678903",
				Status:  "PROCESSED",
				Accrual: decimal.NewFromInt(500),
			}, nil)

		repo.EXPECT().
			UpdateOrderInfo(gomock.Any(), "12345678903", model.OrderProcessed, decimal.NewFromInt(500)).
			Return(nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "49927398716").
			Return(&accrualclient.OrderInfo{
				Number: "49927398716",
				Status: "INVALID",
			}, nil)

		repo.EXPECT().
			UpdateOrderInfo(gomock.Any(), "49927398716", model.OrderInvalid, decimal.Decimal{}).
			Return(nil)

		poller.processBatch(context.Background())
	})

	t.Run("no orders to process", func(t *testing.T) {
		poller, repo, _ := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{}, nil)

		poller.processBatch(context.Background())
	})

	t.Run("find orders error", func(t *testing.T) {
		poller, repo, _ := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		poller.processBatch(context.Background())
	})

	t.Run("client returns nil order info", func(t *testing.T) {
		poller, repo, client := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{"12345678903"}, nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "12345678903").
			Return(nil, nil)

		poller.processBatch(context.Background())
	})

	t.Run("processing status does not update", func(t *testing.T) {
		poller, repo, client := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{"12345678903"}, nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "12345678903").
			Return(&accrualclient.OrderInfo{
				Number: "12345678903",
				Status: "PROCESSING",
			}, nil)

		poller.processBatch(context.Background())
	})

	t.Run("client error", func(t *testing.T) {
		poller, repo, client := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{"12345678903"}, nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "12345678903").
			Return(nil, errors.New("network error"))

		poller.processBatch(context.Background())
	})

	t.Run("rate limited", func(t *testing.T) {
		poller, repo, client := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{"12345678903"}, nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "12345678903").
			Return(nil, &errs.ErrRateLimited{RetryAfter: 10 * time.Millisecond})

		poller.processBatch(context.Background())
	})

	t.Run("update order error", func(t *testing.T) {
		poller, repo, client := setupTestPollerWithClient(t, 1*time.Hour, 0)

		repo.EXPECT().
			FindOrdersByStatuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]string{"12345678903"}, nil)

		client.EXPECT().
			GetOrderInfo(gomock.Any(), "12345678903").
			Return(&accrualclient.OrderInfo{
				Number:  "12345678903",
				Status:  "PROCESSED",
				Accrual: decimal.NewFromInt(500),
			}, nil)

		repo.EXPECT().
			UpdateOrderInfo(gomock.Any(), "12345678903", model.OrderProcessed, gomock.Any()).
			Return(errors.New("update error"))

		poller.processBatch(context.Background())
	})
}

func setupTestPoller(t *testing.T, pollingInterval, requestInterval time.Duration) (*Poller, *mocks.MockOrdersRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := mocks.NewMockAccrualClient(ctrl)
	mockRepo := mocks.NewMockOrdersRepository(ctrl)

	poller := NewPoller(mockClient, mockRepo, pollingInterval, requestInterval, zap.NewNop())
	return poller, mockRepo
}

func setupTestPollerWithClient(t *testing.T, pollingInterval, requestInterval time.Duration) (*Poller, *mocks.MockOrdersRepository, *mocks.MockAccrualClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockClient := mocks.NewMockAccrualClient(ctrl)
	mockRepo := mocks.NewMockOrdersRepository(ctrl)

	poller := NewPoller(mockClient, mockRepo, pollingInterval, requestInterval, zap.NewNop())
	return poller, mockRepo, mockClient
}
