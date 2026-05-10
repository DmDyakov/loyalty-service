package handler

import (
	"encoding/json"
	"errors"
	"io"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/handler/middleware"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type OrdersHandler struct {
	srv    OrdersService
	cfg    *config.Config
	logger *zap.Logger
}

func newOrdersHandler(orderService OrdersService, cfg *config.Config, logger *zap.Logger) *OrdersHandler {
	return &OrdersHandler{
		srv:    orderService,
		cfg:    cfg,
		logger: logger,
	}
}

func (h *OrdersHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	const maxBodySize = 64
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	body, err := io.ReadAll(r.Body)
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			h.logger.Error("failed to close request body", zap.Error(closeErr))
		}
	}()
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))

	const StatusUnprocessableEntity = 422
	if err := h.validateOrderNumber(orderNumber); err != nil {

		http.Error(w, err.Error(), StatusUnprocessableEntity)
		return
	}

	if err := h.srv.AddOrder(ctx, userID, orderNumber); err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidOrderNumber):
			http.Error(w, err.Error(), StatusUnprocessableEntity)
			return
		case errors.Is(err, errs.ErrOrderAlreadyExists):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, errs.ErrOrderUploadedByAnother):
			http.Error(w, err.Error(), http.StatusConflict)
			return

		default:
			h.logger.Error("failed to add order", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrdersHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	params, err := parsePaginationParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orders, err := h.srv.GetUserOrders(ctx, userID, params.Limit, params.Offset)
	if err != nil {
		h.logger.Error("failed to get user orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func (h *OrdersHandler) validateOrderNumber(orderNumber string) error {
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
