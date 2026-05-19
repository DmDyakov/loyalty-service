package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestIsRetriableDBError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "connection exception",
			err:       &pgconn.PgError{Code: "08000"},
			wantRetry: true,
		},
		{
			name:      "serialization failure",
			err:       &pgconn.PgError{Code: "40001"},
			wantRetry: true,
		},
		{
			name:      "deadlock detected",
			err:       &pgconn.PgError{Code: "40P01"},
			wantRetry: true,
		},
		{
			name:      "admin shutdown",
			err:       &pgconn.PgError{Code: "57P01"},
			wantRetry: true,
		},
		{
			name:      "too many connections",
			err:       &pgconn.PgError{Code: "53300"},
			wantRetry: true,
		},
		{
			name:      "lock not available",
			err:       &pgconn.PgError{Code: "55P03"},
			wantRetry: true,
		},
		{
			name:      "unique violation - not retriable",
			err:       &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			wantRetry: false,
		},
		{
			name:      "non-pg error",
			err:       errors.New("generic error"),
			wantRetry: false,
		},
		{
			name:      "nil error",
			err:       nil,
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriableDBError(tt.err)
			assert.Equal(t, tt.wantRetry, got)
		})
	}
}

func TestIsUniqueViolation(t *testing.T) {
	t.Run("unique violation", func(t *testing.T) {
		err := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
		assert.True(t, isUniqueViolation(err))
	})

	t.Run("other pg error", func(t *testing.T) {
		err := &pgconn.PgError{Code: "40001"}
		assert.False(t, isUniqueViolation(err))
	})

	t.Run("non-pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.False(t, isUniqueViolation(err))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, isUniqueViolation(nil))
	})
}

func TestDoWithRetry_Success(t *testing.T) {
	logger := zap.NewNop()
	callCount := 0

	fn := func() (int, error) {
		callCount++
		return 42, nil
	}

	result, err := doWithRetry(context.Background(), logger, fn)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 1, callCount)
}

func TestDoWithRetry_NonRetriableError(t *testing.T) {
	logger := zap.NewNop()
	nonRetriable := errors.New("non-retriable error")

	fn := func() (int, error) {
		return 0, nonRetriable
	}

	result, err := doWithRetry(context.Background(), logger, fn)

	assert.Error(t, err)
	assert.Equal(t, nonRetriable, err)
	assert.Equal(t, 0, result)
}

func TestDoWithRetry_ContextCanceled(t *testing.T) {
	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() (int, error) {
		cancel()
		return 0, &pgconn.PgError{Code: "40001"}
	}

	result, err := doWithRetry(ctx, logger, fn)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Equal(t, 0, result)
}
