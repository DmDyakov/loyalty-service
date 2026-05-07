package service

import (
	"context"
	"errors"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"loyalty-service/internal/service/mocks"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := setupTestAuthService(t)

		repo.EXPECT().
			SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&model.User{ID: 1, Login: "test_login"}, nil)

		token, err := s.Register(context.Background(), "test_login", "test_password")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
	t.Run("user already exists", func(t *testing.T) {
		s, repo := setupTestAuthService(t)

		repo.EXPECT().
			SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errs.ErrLoginTaken)

		token, err := s.Register(context.Background(), "test_login", "test_password")

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrLoginTaken)
	})
}

func TestService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := setupTestAuthService(t)
		password := "password123"

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "test_login").
			Return(&model.User{
				ID:           1,
				Login:        "test_login",
				PasswordHash: hashForTest(password),
			}, nil)

		token, err := s.Login(context.Background(), "test_login", password)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("user not found", func(t *testing.T) {
		s, repo := setupTestAuthService(t)

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "nonexistent").
			Return(nil, errs.ErrInvalidCredentials)

		token, err := s.Login(context.Background(), "nonexistent", "password123")

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
	})

	t.Run("invalid password", func(t *testing.T) {
		s, repo := setupTestAuthService(t)

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "test_login").
			Return(&model.User{
				ID:           1,
				Login:        "test_login",
				PasswordHash: hashForTest("correct_password"),
			}, nil)

		token, err := s.Login(context.Background(), "test_login", "wrong_password")

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, errs.ErrInvalidCredentials)
	})

	t.Run("repository error", func(t *testing.T) {
		s, repo := setupTestAuthService(t)

		repo.EXPECT().
			FindUserByLogin(gomock.Any(), "test_login").
			Return(nil, errors.New("db connection error"))

		token, err := s.Login(context.Background(), "test_login", "password123")

		require.Error(t, err)
		assert.Empty(t, token)
	})
}

func TestService_ValidateToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, _ := setupTestAuthService(t)

		token, err := s.generateToken(42, "test_login")
		require.NoError(t, err)

		userID, err := s.ValidateToken(context.Background(), token)

		require.NoError(t, err)
		assert.Equal(t, 42, userID)
	})

	t.Run("invalid token", func(t *testing.T) {
		s, _ := setupTestAuthService(t)

		userID, err := s.ValidateToken(context.Background(), "invalid.token.here")

		require.Error(t, err)
		assert.Equal(t, 0, userID)
		assert.ErrorIs(t, err, errs.ErrTokenInvalid)
	})

	t.Run("expired token", func(t *testing.T) {
		s, _ := setupTestAuthService(t)
		// Устанавливаем токен с истекшим сроком
		s.tokenExpiry = -1 * time.Hour

		token, err := s.generateToken(42, "test_login")
		require.NoError(t, err)

		userID, err := s.ValidateToken(context.Background(), token)

		require.Error(t, err)
		assert.Equal(t, 0, userID)
		assert.ErrorIs(t, err, errs.ErrTokenInvalid)
	})

	t.Run("wrong signing method", func(t *testing.T) {
		s, _ := setupTestAuthService(t)

		// Создаём токен с другим алгоритмом (не HMAC)
		wrongToken := jwt.NewWithClaims(jwt.SigningMethodRS256, &Claims{
			UserID: 42,
			Login:  "test_login",
		})
		// Используем случайный ключ, так как нам важно только создать токен с другим алгоритмом
		tokenString, _ := wrongToken.SignedString([]byte("some-key"))

		userID, err := s.ValidateToken(context.Background(), tokenString)

		require.Error(t, err)
		assert.Equal(t, 0, userID)
		assert.ErrorIs(t, err, errs.ErrTokenInvalid)
	})

	t.Run("token with different secret", func(t *testing.T) {
		s, _ := setupTestAuthService(t)

		claims := &Claims{
			UserID: 42,
			Login:  "test_login",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		wrongToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := wrongToken.SignedString([]byte("different-secret"))

		userID, err := s.ValidateToken(context.Background(), tokenString)

		require.Error(t, err)
		assert.Equal(t, 0, userID)
		assert.ErrorIs(t, err, errs.ErrTokenInvalid)
	})
}

func hashForTest(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func setupTestAuthService(t *testing.T) (*AuthService, *mocks.MockUserRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUserRepository := mocks.NewMockUserRepository(ctrl)
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
		JWTExpiry: 1,
	}
	logger := zap.NewNop()
	s := NewAuthService(mockUserRepository, cfg, logger)
	return s, mockUserRepository
}
