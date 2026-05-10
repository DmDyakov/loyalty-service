package handler

import (
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/handler/mocks"
	"loyalty-service/internal/model"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestHandler_AddOrder(t *testing.T) {
	const (
		method           = "POST"
		endpoint         = "/api/user/orders"
		validOrderNumber = "12345678903"
		contentType      = "text/plain"
		testJWTToken     = "test-jwt-token"
		userID           = 1
	)

	t.Run("success", func(t *testing.T) {
		h, mockOrders, mockAuth := setupTestOrdersHandler(t)

		mockAuth.EXPECT().ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockOrders.EXPECT().AddOrder(gomock.Any(), userID, validOrderNumber).
			Return(nil)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, validOrderNumber, contentType)
		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h, mockOrders, mockAuth := setupTestOrdersHandler(t)

		mockAuth.EXPECT().ValidateToken(gomock.Any(), gomock.Any()).
			Return(0, errs.ErrTokenInvalid)

		mockOrders.EXPECT().AddOrder(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, validOrderNumber, contentType)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

}

func TestHandler_GetUserOrders(t *testing.T) {
	const (
		method       = "GET"
		endpoint     = "/api/user/orders?limit=10&offset=0"
		testJWTToken = "test-jwt-token"
		userID       = 1
	)

	accrual := decimal.NewFromFloat(500.5)
	testOrders := []model.Order{
		{
			Number:     "12345678903",
			Status:     "PROCESSED",
			Accrual:    &accrual,
			UploadedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			Number:     "12345678904",
			Status:     "NEW",
			Accrual:    nil,
			UploadedAt: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	t.Run("success with orders", func(t *testing.T) {
		h, mockOrders, mockAuth := setupTestOrdersHandler(t)

		mockAuth.EXPECT().ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockOrders.EXPECT().GetUserOrders(gomock.Any(), userID, 10, 0).
			Return(testOrders, nil)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "12345678903")
		assert.Contains(t, w.Body.String(), "PROCESSED")
	})

	t.Run("no orders returns no content", func(t *testing.T) {
		h, mockOrders, mockAuth := setupTestOrdersHandler(t)

		mockAuth.EXPECT().ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockOrders.EXPECT().GetUserOrders(gomock.Any(), userID, 10, 0).
			Return([]model.Order{}, nil)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func setupTestOrdersHandler(t *testing.T) (*Handler, *mocks.MockOrdersService, *mocks.MockAuthService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockOrdersService := mocks.NewMockOrdersService(ctrl)
	mocksAuthService := mocks.NewMockAuthService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{RequestTimeout: 10 * time.Second}

	h := NewHandler(mocksAuthService, mockOrdersService, cfg, logger)
	return h, mockOrdersService, mocksAuthService
}
