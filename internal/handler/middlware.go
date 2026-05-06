package handler

import "go.uber.org/zap"

type Middleware struct {
	authService AuthService
	logger      *zap.Logger
}

func newMiddleware(authService AuthService, l *zap.Logger) *Middleware {
	return &Middleware{
		authService: authService,
		logger:      l,
	}
}
