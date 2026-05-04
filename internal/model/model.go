package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID           int    `json:"-"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
}

type OrderStatus string

const (
	OrderNew        OrderStatus = "NEW"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderInvalid    OrderStatus = "INVALID"
	OrderProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         int              `json:"-"`
	Number     string           `json:"number"`
	Status     OrderStatus      `json:"status"`
	UserID     int              `json:"-"`
	Accrual    *decimal.Decimal `json:"accrual"`
	UploadedAt time.Time        `json:"uploaded_at"`
}

type Withdrawal struct {
	ID          int             `json:"-"`
	OrderNumber string          `json:"order"`
	UserID      int             `json:"-"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"-"`
}

type Balance struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}
