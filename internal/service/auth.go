package service

import (
	"context"
	"fmt"
	"loyalty-service/internal/config"
	"loyalty-service/internal/errs"
	"loyalty-service/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -destination=mocks/user_repository.go -package=mocks . UserRepository
type UserRepository interface {
	SaveUser(ctx context.Context, login string, passwordHash string) (*model.User, error)
	FindUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type AuthService struct {
	userRepository UserRepository
	cfg            *config.Config
	logger         *zap.Logger
	secretKey      []byte
	tokenExpiry    time.Duration
}

type Claims struct {
	UserID int
	Login  string
	jwt.RegisteredClaims
}

func NewAuthService(repo UserRepository, cfg *config.Config, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepository: repo,
		cfg:            cfg,
		logger:         logger,
		secretKey:      []byte(cfg.JWTSecret),
		tokenExpiry:    cfg.JWTExpiry,
	}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return "", err
	}

	user, err := s.userRepository.SaveUser(ctx, login, passwordHash)
	if err != nil {
		return "", err
	}

	return s.generateToken(user.ID, user.Login)
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepository.FindUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if !checkPassword(password, user.PasswordHash) {
		return "", errs.ErrInvalidCredentials
	}

	return s.generateToken(user.ID, user.Login)
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.logger.Warn("unexpected signing method",
				zap.Any("alg", token.Header["alg"]),
			)
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return 0, errs.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			s.logger.Warn("token has expired")
			return 0, errs.ErrTokenInvalid
		}
		return claims.UserID, nil
	}

	s.logger.Warn("invalid token claims")
	return 0, errs.ErrTokenInvalid
}

func (s *AuthService) generateToken(userID int, login string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "loyalty-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		s.logger.Error("failed to sign token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
