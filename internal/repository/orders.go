package repository

import (
	"context"
	"fmt"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type OrdersRepository struct {
	db     *DB
	logger *zap.Logger
}

func NewOrdersRepository(db *DB, l *zap.Logger) *OrdersRepository {
	return &OrdersRepository{
		db:     db,
		logger: l,
	}
}

func (r *OrdersRepository) SaveOrder(ctx context.Context, userID int, orderNumber string) error {
	_, err := r.db.ExecContextWithRetry(
		ctx,
		`INSERT INTO orders (number, user_id)
		VALUES ($1, $2)`,
		orderNumber,
		userID,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return errs.ErrOrderAlreadyExists
		}
		return err
	}

	return nil
}

func (r *OrdersRepository) FindOrdersByUser(ctx context.Context, userID int, limit int, offset int) ([]model.Order, error) {
	query := `SELECT number, status, accrual, uploaded_at FROM orders 
              WHERE user_id = $1 
              ORDER BY uploaded_at DESC`
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

	var orders []model.Order
	for rows.Next() {
		var o model.Order

		if err := rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (r *OrdersRepository) FindOrdersByStatuses(ctx context.Context, statuses []string, limit int, offset int) ([]string, error) {
	if len(statuses) == 0 {
		return nil, errs.ErrOrderStatusesRequired
	}

	query := `SELECT number FROM orders
		WHERE status = ANY($1)
		ORDER BY uploaded_at ASC`
	args := []any{statuses}

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

	var orderNumbers []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		orderNumbers = append(orderNumbers, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orderNumbers, nil
}

func (r *OrdersRepository) UpdateOrderInfo(
	ctx context.Context,
	orderNumber string,
	status model.OrderStatus,
	accrual decimal.Decimal,
) error {
	_, err := r.db.ExecContextWithRetry(
		ctx,
		`UPDATE orders
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE number = $3`,
		status,
		accrual,
		orderNumber,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrdersRepository) FindUserIDByOrderNumber(ctx context.Context, orderNumber string) (int, error) {
	var userID int
	dest := []any{&userID}
	if err := r.db.QueryRowContextWithRetry(
		ctx,
		`SELECT user_id FROM orders WHERE number = $1`,
		dest,
		orderNumber,
	); err != nil {
		return 0, err
	}

	return userID, nil
}
