package accrual

import (
	"context"
	"errors"
	"loyalty-service/internal/client/accrual"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const defaultBatchSize = 50

type AccrualClient interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*accrual.OrderInfo, error)
}

type OrdersRepository interface {
	FindOrdersByStatuses(ctx context.Context, statuses []string, limit, offset int) ([]string, error)
	UpdateOrderInfo(ctx context.Context, orderNumber string, status model.OrderStatus, accrual decimal.Decimal) error
}

type Poller struct {
	client          AccrualClient
	repo            OrdersRepository
	pollingInterval time.Duration
	requestInterval time.Duration
	logger          *zap.Logger
	batchSize       int
}

func NewPoller(
	client AccrualClient,
	repo OrdersRepository,
	pollingInterval time.Duration,
	requestInterval time.Duration,
	l *zap.Logger,
) *Poller {
	return &Poller{
		client:          client,
		repo:            repo,
		pollingInterval: pollingInterval,
		requestInterval: requestInterval,
		logger:          l,
		batchSize:       defaultBatchSize,
	}
}

func (p *Poller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *Poller) processBatch(ctx context.Context) {
	orders, err := p.repo.FindOrdersByStatuses(
		ctx,
		[]string{string(model.OrderNew), string(model.OrderProcessing)},
		p.batchSize,
		0,
	)
	if err != nil {
		p.logger.Error("failed to get orders for polling", zap.Error(err))
		return
	}

	for _, orderNumber := range orders {
		select {
		case <-ctx.Done():
			return
		default:
		}

		orderInfo, err := p.client.GetOrderInfo(ctx, orderNumber)
		if err != nil {
			var rateLimitErr *errs.ErrRateLimited
			if errors.As(err, &rateLimitErr) {
				p.logger.Warn("rate limited", zap.Duration("retry_after", rateLimitErr.RetryAfter))
				select {
				case <-ctx.Done():
					return
				case <-time.After(rateLimitErr.RetryAfter):
				}
			} else {
				p.logger.Error("failed to get order info", zap.String("order", orderNumber), zap.Error(err))
			}
			continue
		}

		if orderInfo == nil {
			continue
		}

		status := model.OrderStatus(orderInfo.Status)
		if status == model.OrderProcessed || status == model.OrderInvalid {
			if err := p.repo.UpdateOrderInfo(ctx, orderNumber, status, orderInfo.Accrual); err != nil {
				p.logger.Error("failed to update order info",
					zap.String("order", orderNumber),
					zap.Error(err),
				)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(p.requestInterval):
		}
	}
}
