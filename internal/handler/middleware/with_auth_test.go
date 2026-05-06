package middleware

import (
	"context"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/handler/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantToken  string
		wantErr    error
	}{
		{
			name:       "valid token",
			authHeader: "Bearer valid-token-123",
			wantToken:  "valid-token-123",
			wantErr:    nil,
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			wantToken:  "",
			wantErr:    ErrMissingAuthHeader,
		},
		{
			name:       "invalid format - no Bearer prefix",
			authHeader: "Basic some-token",
			wantToken:  "",
			wantErr:    ErrInvalidAuthFormat,
		},
		{
			name:       "invalid format - only Bearer without space",
			authHeader: "Bearer",
			wantToken:  "",
			wantErr:    ErrInvalidAuthFormat,
		},
		{
			name:       "invalid format - Bearer with space only",
			authHeader: "Bearer ",
			wantToken:  "",
			wantErr:    ErrMissingToken,
		},
		{
			name:       "lowercase bearer",
			authHeader: "bearer token123",
			wantToken:  "",
			wantErr:    ErrInvalidAuthFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token, err := extractBearerToken(req)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestWithAuth(t *testing.T) {
	validToken := "valid-token"
	invalidToken := "invalid-token"
	userID := 123

	t.Run("valid token sets userID in context", func(t *testing.T) {
		m, mockAuthService := setupTestAuthMiddleware(t)
		mockAuthService.EXPECT().
			ValidateToken(gomock.Any(), validToken).
			Return(userID, nil)

		var capturedUserID int
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := GetUserIDFromContext(r.Context())
			assert.True(t, ok)
			capturedUserID = id
			w.WriteHeader(http.StatusOK)
		})

		handler := m.WithAuth(nextHandler)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, userID, capturedUserID)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		m := &Middleware{
			authService: nil,
			logger:      zap.NewNop(),
		}

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		handler := m.WithAuth(nextHandler)

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		m, mockAuthService := setupTestAuthMiddleware(t)
		mockAuthService.EXPECT().
			ValidateToken(gomock.Any(), invalidToken).
			Return(0, errs.ErrTokenInvalid)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		handler := m.WithAuth(nextHandler)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+invalidToken)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid auth format returns 401", func(t *testing.T) {
		m, _ := setupTestAuthMiddleware(t)

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called")
		})

		handler := m.WithAuth(nextHandler)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Basic some-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Run("userID exists in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, 42)

		userID, ok := GetUserIDFromContext(ctx)

		assert.True(t, ok)
		assert.Equal(t, 42, userID)
	})

	t.Run("userID not in context", func(t *testing.T) {
		ctx := context.Background()

		userID, ok := GetUserIDFromContext(ctx)

		assert.False(t, ok)
		assert.Equal(t, 0, userID)
	})

	t.Run("wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "not-an-int")

		userID, ok := GetUserIDFromContext(ctx)

		assert.False(t, ok)
		assert.Equal(t, 0, userID)
	})
}

func setupTestAuthMiddleware(t *testing.T) (*Middleware, *mocks.MockAuthService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthService := mocks.NewMockAuthService(ctrl)

	m := &Middleware{
		authService: mockAuthService,
		logger:      zap.NewNop(),
	}
	return m, mockAuthService
}
