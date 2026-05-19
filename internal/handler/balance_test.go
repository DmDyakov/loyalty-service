package handler

import (
	"encoding/json"
	"errors"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/handler/mocks"
	"loyalty-service/internal/model"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestHandler_GetUserBalance(t *testing.T) {
	const (
		method       = "GET"
		endpoint     = "/api/user/balance"
		testJWTToken = "test-jwt-token"
		userID       = 1
	)

	current := decimal.NewFromFloat(500.5)
	withdrawn := decimal.NewFromFloat(42.0)

	t.Run("success", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			GetUserBalance(gomock.Any(), userID).
			Return(&model.Balance{
				Current:   &current,
				Withdrawn: &withdrawn,
			}, nil)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusOK, w.Code)

		var resp BalanceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, 500.5, resp.Current)
		assert.Equal(t, 42.0, resp.Withdrawn)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), gomock.Any()).
			Return(0, errs.ErrTokenInvalid)

		w := doTestRequestWithAuth(t, h, method, endpoint, "invalid-token", "", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			GetUserBalance(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_Withdraw(t *testing.T) {
	const (
		method       = "POST"
		endpoint     = "/api/user/balance/withdraw"
		testJWTToken = "test-jwt-token"
		userID       = 1
		contentType  = "application/json"
	)

	t.Run("success", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			Withdraw(gomock.Any(), userID, "2377225624", decimal.NewFromInt(751)).
			Return(nil)

		body := `{"order":"2377225624","sum":751}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("wrong content-type", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		body := `{"order":"2377225624","sum":751}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, "text/plain")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Content-Type must be application/json\n", w.Body.String())
	})

	t.Run("unauthorized", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), gomock.Any()).
			Return(0, errs.ErrTokenInvalid)

		body := `{"order":"2377225624","sum":751}`
		w := doTestRequestWithAuth(t, h, method, endpoint, "invalid-token", body, contentType)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		body := `invalid json`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid order number", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		body := `{"order":"abc","sum":751}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid sum - zero", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		body := `{"order":"2377225624","sum":0}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "sum must be greater than zero\n", w.Body.String())
	})

	t.Run("invalid sum - negative", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		body := `{"order":"2377225624","sum":-100}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			Withdraw(gomock.Any(), userID, "2377225624", decimal.NewFromInt(1000)).
			Return(errs.ErrInsufficientFunds)

		body := `{"order":"2377225624","sum":1000}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusPaymentRequired, w.Code)
	})

	t.Run("invalid order number from service", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			Withdraw(gomock.Any(), userID, "2377225624", decimal.NewFromInt(100)).
			Return(errs.ErrInvalidOrderNumber)

		body := `{"order":"2377225624","sum":100}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, errs.StatusUnprocessable, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			Withdraw(gomock.Any(), userID, "2377225624", decimal.NewFromInt(100)).
			Return(errors.New("db error"))

		body := `{"order":"2377225624","sum":100}`
		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, body, contentType)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetUserWithdrawals(t *testing.T) {
	const (
		method       = "GET"
		endpoint     = "/api/user/withdrawals?limit=10&offset=0"
		testJWTToken = "test-jwt-token"
		userID       = 1
	)

	testWithdrawals := []model.Withdrawal{
		{
			OrderNumber: "2377225624",
			Sum:         decimal.NewFromInt(500),
			ProcessedAt: time.Date(2020, 12, 9, 16, 9, 57, 0, time.FixedZone("", 3*60*60)),
		},
	}

	t.Run("success", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			GetUserWithdrawals(gomock.Any(), userID, 10, 0).
			Return(testWithdrawals, nil)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "2377225624")
		assert.Contains(t, w.Body.String(), "500")
	})

	t.Run("no withdrawals returns no content", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			GetUserWithdrawals(gomock.Any(), userID, 10, 0).
			Return([]model.Withdrawal{}, nil)

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		h, _, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), gomock.Any()).
			Return(0, errs.ErrTokenInvalid)

		w := doTestRequestWithAuth(t, h, method, endpoint, "invalid-token", "", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		h, mockBalance, mockAuth := setupTestBalanceHandler(t)

		mockAuth.EXPECT().
			ValidateToken(gomock.Any(), testJWTToken).
			Return(userID, nil)

		mockBalance.EXPECT().
			GetUserWithdrawals(gomock.Any(), userID, 10, 0).
			Return(nil, errors.New("db error"))

		w := doTestRequestWithAuth(t, h, method, endpoint, testJWTToken, "", "")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func setupTestBalanceHandler(t *testing.T) (*Handler, *mocks.MockBalanceService, *mocks.MockAuthService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockBalanceService := mocks.NewMockBalanceService(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockOrdersService := mocks.NewMockOrdersService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{
		RequestTimeout: 10 * time.Second,
		MaxResults:     100,
	}

	h := NewHandler(
		mockAuthService,
		mockOrdersService,
		mockBalanceService,
		cfg,
		logger,
	)
	return h, mockBalanceService, mockAuthService
}
