package handler

import (
	"encoding/json"
	"errors"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/handler/middleware"
	"loyalty-service/internal/model"
	"net/http"

	"go.uber.org/zap"
)

type BalanceHandler struct {
	srv    BalanceService
	cfg    *config.Config
	logger *zap.Logger
}

func newBalanceHandler(balanceService BalanceService, cfg *config.Config, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		srv:    balanceService,
		cfg:    cfg,
		logger: logger,
	}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	balance, err := h.srv.GetUserBalance(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get user balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var withdrawal model.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := validateOrderNumber(withdrawal.OrderNumber); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateWithdrawalSum(withdrawal.Sum); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.srv.Withdraw(ctx, userID, withdrawal.OrderNumber, withdrawal.Sum); err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidOrderNumber):
			http.Error(w, err.Error(), errs.StatusUnprocessable)
		case errors.Is(err, errs.ErrInsufficientFunds):
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		default:
			h.logger.Error("failed to withdraw", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	params, err := parsePaginationParams(r, h.cfg.MaxResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	withdrawals, err := h.srv.GetUserWithdrawals(ctx, userID, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("failed to get user withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
