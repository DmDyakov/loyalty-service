// Package middleware содержит промежуточные слои для авторизации, логирования и т д.
package middleware

import (
	"context"

	"go.uber.org/zap"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (userID int, err error)
}

type Middleware struct {
	authService AuthService
	logger      *zap.Logger
}

func NewMiddleware(authService AuthService, l *zap.Logger) *Middleware {
	return &Middleware{
		authService: authService,
		logger:      l,
	}
}
