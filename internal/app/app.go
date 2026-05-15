// Package app инициализирует и запускает приложение.
package app

import (
	"context"
	"errors"
	"fmt"
	"loyalty-service/internal/config"
	"loyalty-service/internal/handler"
	"loyalty-service/internal/logger"
	"loyalty-service/internal/repository"
	"loyalty-service/internal/service"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

func Run(ctx context.Context, args []string) error {
	cfg, err := config.New(args)
	if err != nil {
		return fmt.Errorf("failed to create app config: %s", err)
	}

	if cfg == nil {
		return fmt.Errorf("app config is nil")
	}

	logger, err := logger.NewZapLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
		}
	}()

	db, err := repository.NewDB(cfg.DatabaseDSN, logger)
	if err != nil {
		return fmt.Errorf("failed to create db: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", zap.Error(err))
		}
	}()

	userRepo := repository.NewUserRepository(db, logger)
	ordersRepo := repository.NewOrdersRepository(db, logger)
	balanceRepo := repository.NewBalanceRepository(db, logger)

	authSrv := service.NewAuthService(userRepo, cfg, logger)
	ordersSrv := service.NewOrdersService(ordersRepo, cfg, logger)
	balanceSrv := service.NewBalanceService(balanceRepo, cfg, logger)

	handler := handler.NewHandler(
		authSrv,
		ordersSrv,
		balanceSrv,
		cfg,
		logger,
	)
	r := handler.RegisterRoutes()
	server := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	appCtx, appCancel := context.WithCancel(ctx)
	defer appCancel()

	go func() {
		logger.Info("server started", zap.String("url", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen error", zap.Error(err))
			appCancel()
		}
	}()

	<-appCtx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("server stopped gracefully")

	return nil
}
