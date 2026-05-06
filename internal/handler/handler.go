package handler

import (
	"context"
	"loyalty-service/internal/config"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"go.uber.org/zap"
)

const (
	RateLimitDefaultRPH = 1000
	RateLimitDefaultRPM = 100

	RateLimitAuthRPM = 5
)

//go:generate mockgen -destination=mocks/auth_service.go -package=mocks . AuthService
type AuthService interface {
	Register(ctx context.Context, login, password string) (token string, err error)
	Login(ctx context.Context, login, password string) (token string, err error)
	ValidateToken(ctx context.Context, token string) (userID int, err error)
}

type Handler struct {
	authHandler *AuthHandler
	cfg         *config.Config
	logger      *zap.Logger
	middleware  *Middleware
}

func NewHandler(
	authService AuthService,
	cfg *config.Config,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		authHandler: newAuthHandler(authService, cfg, logger),
		middleware:  newMiddleware(authService, logger),
		logger:      logger,
		cfg:         cfg,
	}
}

func (h *Handler) RegisterRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(h.middleware.withLogging)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5,
		"application/json",
		"text/plain",
	))
	r.Use(middleware.Timeout(h.cfg.RequestTimeout))
	r.Use(httprate.LimitByIP(RateLimitDefaultRPH, 1*time.Hour))

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(RateLimitAuthRPM, 1*time.Minute))

		r.Route("/api/user", func(r chi.Router) {
			r.Post("/register", h.authHandler.Register)
			r.Post("/login", h.authHandler.Login)
		})

	})

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(RateLimitDefaultRPM, 1*time.Minute))
		r.Use(h.middleware.withAuth)
		// TODO: реализовать приватные роуты

	})

	return r
}
