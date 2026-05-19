// Package middleware содержит промежуточные слои для авторизации, логирования и т д.
package middleware

import (
	"context"

	"go.uber.org/zap"
)

// AuthService — интерфейс сервиса аутентификации для middleware.
type AuthService interface {
	ValidateToken(ctx context.Context, token string) (userID int, err error)
}

// Middleware — структура, содержащая зависимости для middleware-слоёв.
type Middleware struct {
	authService AuthService
	logger      *zap.Logger
}

// NewMiddleware создает новый экземпляр Middleware.
func NewMiddleware(authService AuthService, l *zap.Logger) *Middleware {
	return &Middleware{
		authService: authService,
		logger:      l,
	}
}
