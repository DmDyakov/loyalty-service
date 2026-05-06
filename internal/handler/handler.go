package handler

import (
	"context"
	"loyalty-service/internal/config"
	"loyalty-service/internal/handler/middleware"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
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
	middleware  *middleware.Middleware
}

func NewHandler(
	authService AuthService,
	cfg *config.Config,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		authHandler: newAuthHandler(authService, cfg, logger),
		middleware:  middleware.NewMiddleware(authService, logger),
		logger:      logger,
		cfg:         cfg,
	}
}

func (h *Handler) RegisterRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(h.middleware.WithLogging)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Compress(5,
		"application/json",
		"text/plain",
	))
	r.Use(chimw.Timeout(h.cfg.RequestTimeout))
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
		r.Use(h.middleware.WithAuth)
		// r.Route("/api/user", func(r chi.Router) {
		// 	// r.Get("/orders", h.orderHandler.OrdersByUserId)
		// })

	})

	return r
}
