// Package handler реализует HTTP-обработчики
package handler

import (
	"context"
	"loyalty-service/internal/config"
	"loyalty-service/internal/handler/middleware"
	"loyalty-service/internal/model"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/shopspring/decimal"
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

//go:generate mockgen -destination=mocks/orders_service.go -package=mocks . OrdersService
type OrdersService interface {
	AddOrder(ctx context.Context, userID int, orderNumber string) error
	GetUserOrders(ctx context.Context, userID int, limit int, offset int) ([]model.Order, error)
}

//go:generate mockgen -destination=mocks/balance_service.go -package=mocks . BalanceService
type BalanceService interface {
	GetUserBalance(ctx context.Context, userID int) (*model.Balance, error)
	Withdraw(ctx context.Context, userID int, orderNumber string, sum decimal.Decimal) error
	GetUserWithdrawals(ctx context.Context, userID int, limit int, offset int) ([]model.Withdrawal, error)
}

type Handler struct {
	authHandler    *AuthHandler
	ordersHandler  *OrdersHandler
	balanceHandler *BalanceHandler
	cfg            *config.Config
	logger         *zap.Logger
	middleware     *middleware.Middleware
}

func NewHandler(
	authService AuthService,
	ordersService OrdersService,
	balanceService BalanceService,
	cfg *config.Config,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		authHandler:    newAuthHandler(authService, cfg, logger),
		ordersHandler:  newOrdersHandler(ordersService, cfg, logger),
		balanceHandler: newBalanceHandler(balanceService, cfg, logger),
		middleware:     middleware.NewMiddleware(authService, logger),
		logger:         logger,
		cfg:            cfg,
	}
}

// RegisterRoutes регистрирует все маршруты API с применением middleware.
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

	r.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(RateLimitAuthRPM, 1*time.Minute))
			r.Post("/register", h.authHandler.Register)
			r.Post("/login", h.authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(RateLimitDefaultRPM, 1*time.Minute))
			r.Use(h.middleware.WithAuth)

			// Управление заказами
			r.Post("/orders", h.ordersHandler.AddOrder)
			r.Get("/orders", h.ordersHandler.GetUserOrders)

			// Управление балансом
			r.Get("/balance", h.balanceHandler.GetUserBalance)
			r.Post("/balance/withdraw", h.balanceHandler.Withdraw)
			r.Get("/withdrawals", h.balanceHandler.GetUserWithdrawals)
		})
	})

	return r
}
