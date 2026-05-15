package handler

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

func (h *AuthHandler) validateCredentials(creds Credentials) error {
	const (
		minLoginLen    = 3
		maxLoginLen    = 50
		minPasswordLen = 8
		maxPasswordLen = 100
	)

	if creds.Login == "" && creds.Password == "" {
		return errors.New("login and password are required")
	}

	if creds.Login == "" {
		return errors.New("login is required")
	}

	if creds.Password == "" {
		return errors.New("password is required")
	}

	if len(creds.Login) < minLoginLen {
		return fmt.Errorf("login must be at least %d characters", minLoginLen)
	}
	if len(creds.Login) > maxLoginLen {
		return fmt.Errorf("login must not exceed %d characters", maxLoginLen)
	}

	if len(creds.Password) < minPasswordLen {
		return fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}
	if len(creds.Password) > maxPasswordLen {
		return fmt.Errorf("password must not exceed %d characters", maxPasswordLen)
	}

	return nil
}

func validateOrderNumber(orderNumber string) error {
	if len(orderNumber) == 0 {
		return errors.New("order number is required")
	}
	for _, r := range orderNumber {
		if r < '0' || r > '9' {
			return errors.New("order number must be consist of digits only")
		}
	}

	return nil
}

func validateWithdrawalSum(sum decimal.Decimal) error {
	if sum.LessThanOrEqual(decimal.Zero) {
		return errors.New("sum must be greater than zero")
	}

	if sum.Exponent() < -2 {
		return errors.New("sum must have at most 2 decimal places")
	}

	return nil
}
