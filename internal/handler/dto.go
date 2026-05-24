package handler

import (
	"time"

	"loyalty-service/internal/model"
)

// OrderResponse представляет ответ с информацией о заказе.
type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// WithdrawalResponse представляет ответ с информацией о списании.
type WithdrawalResponse struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// BalanceResponse представляет ответ с информацией о балансе.
type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// toOrdersResponse преобразует модель заказов в формат ответа API.
func toOrdersResponse(orders []model.Order) []OrderResponse {
	result := make([]OrderResponse, len(orders))
	for i, o := range orders {
		resp := OrderResponse{
			Number:     o.Number,
			Status:     string(o.Status),
			UploadedAt: o.UploadedAt,
		}
		if o.Accrual != nil {
			f, _ := o.Accrual.Float64()
			resp.Accrual = &f
		}
		result[i] = resp
	}
	return result
}

// toWithdrawalsResponse преобразует модель списаний в формат ответа API.
func toWithdrawalsResponse(withdrawals []model.Withdrawal) []WithdrawalResponse {
	result := make([]WithdrawalResponse, len(withdrawals))
	for i, w := range withdrawals {
		sum, _ := w.Sum.Float64()
		result[i] = WithdrawalResponse{
			OrderNumber: w.OrderNumber,
			Sum:         sum,
			ProcessedAt: w.ProcessedAt,
		}
	}
	return result
}

// toBalanceResponse преобразует модель баланса в формат ответа API.
func toBalanceResponse(balance *model.Balance) BalanceResponse {
	current, _ := balance.Current.Float64()
	withdrawn, _ := balance.Withdrawn.Float64()
	return BalanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}
}
