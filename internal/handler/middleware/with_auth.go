package middleware

import (
	"context"
	"errors"
	"loyalty-service/internal/errs"
	"net/http"
	"strings"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthFormat = errors.New("invalid authorization format")
	ErrMissingToken      = errors.New("empty token")
)

type contextKey string

const UserIDKey contextKey = "userID"

func (m *Middleware) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractBearerToken(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		userID, err := m.authService.ValidateToken(r.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrTokenInvalid):
				http.Error(w, errs.ErrTokenInvalid.Error(), http.StatusUnauthorized)
			default:
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", ErrInvalidAuthFormat
	}

	token := strings.TrimPrefix(authHeader, prefix)
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}
