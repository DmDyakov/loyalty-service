package handler

import (
	"time"

	"loyalty-service/internal/model"
)

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type WithdrawalResponse struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

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

func toBalanceResponse(balance *model.Balance) BalanceResponse {
	current, _ := balance.Current.Float64()
	withdrawn, _ := balance.Withdrawn.Float64()
	return BalanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}
}
