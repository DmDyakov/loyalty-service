package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthHandler struct {
	authService AuthService
	cfg         *config.Config
	logger      *zap.Logger
}

func newAuthHandler(
	srv AuthService,
	cfg *config.Config,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: srv,
		cfg:         cfg,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.validateCredentials(creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	token, err := h.authService.Register(ctx, creds.Login, creds.Password)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrLoginTaken):
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		default:
			h.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	h.respondWithToken(w, token)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.validateCredentials(creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	token, err := h.authService.Login(ctx, creds.Login, creds.Password)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrInvalidCredentials):
			http.Error(w, errs.ErrInvalidCredentials.Error(), http.StatusUnauthorized)
		default:
			h.logger.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	h.respondWithToken(w, token)

}

func (h *AuthHandler) respondWithToken(w http.ResponseWriter, token string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token": token,
		"token_type":   "Bearer",
	}); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

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
