// Package app инициализирует и запускает приложение.
package app

import (
	"context"
	"errors"
	"fmt"
	accrualclient "loyalty-service/internal/client/accrual"
	"loyalty-service/internal/config"
	"loyalty-service/internal/handler"
	"loyalty-service/internal/logger"
	"loyalty-service/internal/repository"
	"loyalty-service/internal/service"
	accrualworker "loyalty-service/internal/worker/accrual"
	"loyalty-service/migrations"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

	if err := migrations.Run(db.DB); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

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

	accrualClient := accrualclient.New(cfg.AccrualBaseURL, logger)

	accrualPoller := accrualworker.NewPoller(
		accrualClient,
		ordersRepo,
		cfg.PollingInterval,
		cfg.RequestInterval,
		logger,
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("server started", zap.String("url", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server listen error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("accrual poller started")
		accrualPoller.Start(gCtx)
		logger.Info("accrual poller stopped")

		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()

		logger.Info("shutdown signal received")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		logger.Info("server stopped gracefully")

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
