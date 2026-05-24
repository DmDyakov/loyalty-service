package handler

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestValidateOrderNumber(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid order number",
			orderNumber: "12345678903",
			wantErr:     false,
		},
		{
			name:        "empty order number",
			orderNumber: "",
			wantErr:     true,
			errContains: "order number is required",
		},
		{
			name:        "order number with letters",
			orderNumber: "123abc",
			wantErr:     true,
			errContains: "order number must be consist of digits only",
		},
		{
			name:        "order number with special characters",
			orderNumber: "123-456",
			wantErr:     true,
			errContains: "order number must be consist of digits only",
		},
		{
			name:        "order number with spaces",
			orderNumber: "123 456",
			wantErr:     true,
			errContains: "order number must be consist of digits only",
		},
		{
			name:        "order number with leading zeros",
			orderNumber: "000123",
			wantErr:     false,
		},
		{
			name:        "single digit",
			orderNumber: "5",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOrderNumber(tt.orderNumber)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithdrawalSum(t *testing.T) {
	tests := []struct {
		name        string
		sum         decimal.Decimal
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid sum - integer",
			sum:     decimal.NewFromInt(100),
			wantErr: false,
		},
		{
			name:    "valid sum - one decimal",
			sum:     decimal.NewFromFloat(100.5),
			wantErr: false,
		},
		{
			name:    "valid sum - two decimals",
			sum:     decimal.NewFromFloat(100.55),
			wantErr: false,
		},
		{
			name:        "zero sum",
			sum:         decimal.Zero,
			wantErr:     true,
			errContains: "sum must be greater than zero",
		},
		{
			name:        "negative sum",
			sum:         decimal.NewFromInt(-100),
			wantErr:     true,
			errContains: "sum must be greater than zero",
		},
		{
			name:        "three decimals",
			sum:         decimal.NewFromFloat(100.555),
			wantErr:     true,
			errContains: "sum must have at most 2 decimal places",
		},
		{
			name:    "valid sum - small fraction",
			sum:     decimal.NewFromFloat(0.01),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWithdrawalSum(tt.sum)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
