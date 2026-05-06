package middleware

import (
	"context"

	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (token string, err error)
	Login(ctx context.Context, login, password string) (token string, err error)
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
