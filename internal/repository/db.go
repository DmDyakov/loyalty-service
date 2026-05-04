package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const dbConnectionTimeout = 5 * time.Second

type DB struct {
	*sql.DB
	logger *zap.Logger
}

func NewDB(databaseDSN string, logger *zap.Logger) (*DB, error) {
	sqlDB, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectionTimeout)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{sqlDB, logger}, nil
}

func (db *DB) ExecContextWithRetry(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return doWithRetry(ctx, db.logger, func() (sql.Result, error) {
		return db.ExecContext(ctx, query, args...)
	})
}

func (db *DB) QueryContextWithRetry(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return doWithRetry(ctx, db.logger, func() (*sql.Rows, error) {
		return db.QueryContext(ctx, query, args...)
	})
}

func doWithRetry[T any](ctx context.Context, logger *zap.Logger, fn func() (T, error)) (T, error) {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	const maxAttempts = 4

	var zero T
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		if !isRetriableDBError(err) || attempt == maxAttempts {
			return zero, err
		}

		delay := delays[attempt-1]
		logger.Warn("retrying db operation due to retriable error",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
			if ctx.Err() != nil {
				return zero, ctx.Err()
			}
		}
	}
	return zero, nil
}

func isRetriableDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return true
		}

		switch pgErr.Code {
		case "40001", "40P01", "57P01", "53300", "55P03":
			return true
		}
	}
	return false
}
