// Package repository содержит слой доступа к базе данных.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// UserRepository управляет данными пользователей в базе данных.
type UserRepository struct {
	db     *DB
	logger *zap.Logger
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db *DB, l *zap.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: l,
	}
}

// SaveUser сохраняет нового пользователя в базе данных.
func (r *UserRepository) SaveUser(ctx context.Context, login string, passwordHash string) (*model.User, error) {
	user := &model.User{}
	dest := []any{&user.ID, &user.Login}
	err := r.db.QueryRowContextWithRetry(
		ctx,
		`INSERT INTO users (login, password_hash) 
			VALUES ($1, $2) 
			RETURNING id, login`,
		dest,
		login,
		passwordHash,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, errs.ErrLoginTaken
		}

		return nil, err
	}

	return user, nil
}

// FindUserByLogin ищет пользователя по логину.
func (r *UserRepository) FindUserByLogin(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{
		Login: login,
	}
	dest := []any{&user.ID, &user.PasswordHash}
	err := r.db.QueryRowContextWithRetry(
		ctx,
		`SELECT id, password_hash FROM users 
		WHERE login = $1`,
		dest,
		login,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, err
	}
	return user, nil
}
