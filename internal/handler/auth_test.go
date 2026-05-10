package handler

import (
	"encoding/json"
	"errors"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/handler/mocks"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestHandler_Register(t *testing.T) {
	const (
		method       = "POST"
		endpoint     = "/api/user/register"
		testReqData  = `{"login":"user123","password":"password123"}`
		contentType  = "application/json"
		testJWTToken = "test-jwt-token"
		tokenType    = "Bearer"
	)

	t.Run("success", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)

		mockAuthService.EXPECT().
			Register(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(testJWTToken, nil)

		req, w := doTestRequest(t, method, endpoint, testReqData, contentType)
		h.Register(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, testJWTToken, resp["access_token"])
		assert.Equal(t, tokenType, resp["token_type"])
	})

	t.Run("wrong content-type", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)

		mockAuthService.EXPECT().
			Register(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		req, w := doTestRequest(t, method, endpoint, testReqData, "text/plain")
		h.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Content-Type must be application/json\n", w.Body.String())
	})

	t.Run("validation", func(t *testing.T) {
		h, _ := setupTestAuthHandler(t)
		testAuthValidation(t, method, endpoint, contentType, h.Register)
	})

	t.Run("error login taken", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)
		mockAuthService.EXPECT().
			Register(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", errs.ErrLoginTaken)

		req, w := doTestRequest(t, method, endpoint, testReqData, contentType)
		h.Register(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, "Conflict\n", w.Body.String())
	})

	t.Run("other any error", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)
		mockAuthService.EXPECT().
			Register(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", errors.New("test error"))

		req, w := doTestRequest(t, method, endpoint, testReqData, contentType)
		h.Register(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Internal Server Error\n", w.Body.String())
	})
}

func TestHandler_Login(t *testing.T) {
	const (
		method       = "POST"
		endpoint     = "/api/user/login"
		testReqData  = `{"login":"user123","password":"password123"}`
		contentType  = "application/json"
		testJWTToken = "test-jwt-token"
		tokenType    = "Bearer"
	)

	t.Run("success", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)

		mockAuthService.EXPECT().
			Login(gomock.Any(), "user123", "password123").
			Return(testJWTToken, nil)

		req, w := doTestRequest(t, method, endpoint, testReqData, contentType)
		h.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, testJWTToken, resp["access_token"])
		assert.Equal(t, tokenType, resp["token_type"])
	})

	t.Run("wrong content-type", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)

		mockAuthService.EXPECT().
			Login(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		req, w := doTestRequest(t, method, endpoint, testReqData, "text/plain")
		h.Login(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Content-Type must be application/json\n", w.Body.String())
	})

	t.Run("invalid credentials", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)
		mockAuthService.EXPECT().
			Login(gomock.Any(), "user123", "password123").
			Return("", errs.ErrInvalidCredentials)

		req, w := doTestRequest(t, method, endpoint, testReqData, contentType)
		h.Login(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "invalid login or password\n", w.Body.String())
	})

	t.Run("validation", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)
		mockAuthService.EXPECT().
			Login(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
		testAuthValidation(t, method, endpoint, contentType, h.Login)
	})

	t.Run("other any error", func(t *testing.T) {
		h, mockAuthService := setupTestAuthHandler(t)
		mockAuthService.EXPECT().
			Login(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", errors.New("test error"))

		req, w := doTestRequest(t, method, endpoint, testReqData, contentType)
		h.Login(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Internal Server Error\n", w.Body.String())
	})
}

func testAuthValidation(t *testing.T, method, endpoint, contentType string, handler http.HandlerFunc) {
	t.Helper()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "empty login and password",
			body:       `{"login":"","password":""}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "login and password are required\n",
		},
		{
			name:       "empty login",
			body:       `{"login":"","password":"secret123"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "login is required\n",
		},
		{
			name:       "empty password",
			body:       `{"login":"user","password":""}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "password is required\n",
		},
		{
			name:       "login too short",
			body:       `{"login":"ab","password":"secret123"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "login must be at least 3 characters\n",
		},
		{
			name:       "login too long",
			body:       `{"login":"` + strings.Repeat("a", 51) + `","password":"secret123"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "login must not exceed 50 characters\n",
		},
		{
			name:       "password too short",
			body:       `{"login":"user","password":"1234567"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "password must be at least 8 characters\n",
		},
		{
			name:       "password too long",
			body:       `{"login":"user","password":"` + strings.Repeat("a", 101) + `"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "password must not exceed 100 characters\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, w := doTestRequest(t, method, endpoint, tt.body, contentType)
			handler(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

func setupTestAuthHandler(t *testing.T) (*AuthHandler, *mocks.MockAuthService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAuthService := mocks.NewMockAuthService(ctrl)
	logger := zap.NewNop()
	cfg := &config.Config{RequestTimeout: 10 * time.Second}

	h := newAuthHandler(mockAuthService, cfg, logger)
	return h, mockAuthService
}
