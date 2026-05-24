package accrual

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClient_GetOrderInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, srv := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(OrderInfo{
				Number:  "12345678903",
				Status:  "PROCESSED",
				Accrual: decimal.NewFromFloat(500.5),
			})
		})
		defer srv.Close()

		info, err := client.GetOrderInfo(context.Background(), "12345678903")

		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "12345678903", info.Number)
		assert.Equal(t, "PROCESSED", info.Status)
	})

	t.Run("no content", func(t *testing.T) {
		client, srv := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		defer srv.Close()

		info, err := client.GetOrderInfo(context.Background(), "99999999999")

		require.NoError(t, err)
		assert.Nil(t, info)
	})

	t.Run("rate limited with retry-after", func(t *testing.T) {
		client, srv := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
		})
		defer srv.Close()

		info, err := client.GetOrderInfo(context.Background(), "12345678903")

		require.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "rate limited")
	})

	t.Run("rate limited without retry-after", func(t *testing.T) {
		client, srv := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		})
		defer srv.Close()

		info, err := client.GetOrderInfo(context.Background(), "12345678903")

		require.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "rate limited")
	})

	t.Run("unexpected status", func(t *testing.T) {
		client, srv := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer srv.Close()

		info, err := client.GetOrderInfo(context.Background(), "12345678903")

		require.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "unexpected status")
	})

	t.Run("invalid json", func(t *testing.T) {
		client, srv := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("invalid json"))
		})
		defer srv.Close()

		info, err := client.GetOrderInfo(context.Background(), "12345678903")

		require.Error(t, err)
		assert.Nil(t, info)
	})
}

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := New(server.URL, zap.NewNop())
	return client, server
}
