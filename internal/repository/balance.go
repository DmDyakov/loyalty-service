package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type BalanceRepository struct {
	db     *DB
	logger *zap.Logger
}

func NewBalanceRepository(db *DB, l *zap.Logger) *BalanceRepository {
	return &BalanceRepository{
		db:     db,
		logger: l,
	}
}

func (r *BalanceRepository) GetAccrualSumByUser(ctx context.Context, userID int) (*decimal.Decimal, error) {
	var sum *decimal.Decimal
	dest := []any{&sum}
	err := r.db.QueryRowContextWithRetry(
		ctx,
		`SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1`,
		dest,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return sum, nil
}

func (r *BalanceRepository) GetWithdrawnSumByUser(ctx context.Context, userID int) (*decimal.Decimal, error) {
	var sum *decimal.Decimal
	dest := []any{&sum}
	err := r.db.QueryRowContextWithRetry(
		ctx,
		`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`,
		dest,
		userID,
	)
	if err != nil {
		return nil, err
	}

	return sum, nil
}

func (r *BalanceRepository) SaveWithdrawal(ctx context.Context, userID int, orderNumber string, sum decimal.Decimal) error {
	_, err := doWithRetry(ctx, r.db.logger, func() (any, error) {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}

		defer func() {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				r.db.logger.Error("rollback failed", zap.Error(rbErr))
			}
		}()

		_, err = tx.ExecContext(ctx,
			`SELECT 1 FROM orders WHERE user_id = $1 FOR UPDATE`, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("lock orders: %w", err)
		}

		var accrualSum, withdrawnSum decimal.Decimal

		if err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1`, userID,
		).Scan(&accrualSum); err != nil {
			return nil, fmt.Errorf("get accrualSum: %w", err)
		}

		if err := tx.QueryRowContext(ctx,
			`SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`, userID,
		).Scan(&withdrawnSum); err != nil {
			return nil, fmt.Errorf("get withdrawnSum: %w", err)
		}

		currentBalance := decimal.Zero.Add(accrualSum).Sub(withdrawnSum)
		if currentBalance.LessThan(sum) {
			return nil, errs.ErrInsufficientFunds
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO withdrawals (user_id, order_number, sum)
			VALUES ($1, $2, $3)`,
			userID,
			orderNumber,
			sum,
		)

		if err != nil {
			return nil, err
		}

		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit tx: %w", err)
		}

		return nil, nil
	})

	return err
}

func (r *BalanceRepository) FindWithdrawalsByUser(ctx context.Context, userID int, limit int, offset int) ([]model.Withdrawal, error) {
	query := `SELECT order_number, sum, processed_at FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC`
	args := []any{userID}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", len(args)+1)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContextWithRetry(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []model.Withdrawal
	for rows.Next() {
		var w model.Withdrawal
		if err := rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
