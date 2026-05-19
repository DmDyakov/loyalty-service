package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"loyalty-service/internal/errs"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	logger     *zap.Logger
	httpClient *http.Client
}

type OrderInfo struct {
	Number  string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func New(baseURL string, l *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		logger:  l,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderInfo, error) {
	method := http.MethodGet
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	c.logger.Info("requesting accrual",
		zap.String("method", method),
		zap.String("url", url),
		zap.String("order", orderNumber),
	)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("accrual service unavailable",
			zap.String("url", url),
		)
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error("failed to close response body", zap.Error(err))
		}
	}()

	c.logger.Info("accrual response",
		zap.String("order", orderNumber),
		zap.Int("status", resp.StatusCode),
	)

	switch resp.StatusCode {
	case http.StatusOK:
		var orderInfo OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&orderInfo); err != nil {
			c.logger.Error("failed to decode accrual response",
				zap.String("order", orderNumber),
				zap.Error(err),
			)
			return nil, fmt.Errorf("decode response: %w", err)
		}
		c.logger.Info("accrual info",
			zap.String("order", orderInfo.Number),
			zap.String("status", orderInfo.Status),
			zap.String("accrual", orderInfo.Accrual.String()),
		)
		return &orderInfo, nil

	case http.StatusNoContent:
		c.logger.Info("order not registered in accrual system",
			zap.String("order", orderNumber),
		)
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			seconds, err := strconv.Atoi(retryAfter)
			if err != nil {
				seconds = 0
			}

			delay := time.Duration(seconds) * time.Second
			c.logger.Warn("rate limited by accrual service",
				zap.Duration("retry_after", delay),
				zap.String("order", orderNumber),
			)
			return nil, &errs.ErrRateLimited{RetryAfter: delay}
		}
	default:
		c.logger.Error("unexpected status from accrual",
			zap.String("order", orderNumber),
			zap.Int("status", resp.StatusCode),
		)
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil, nil
}
