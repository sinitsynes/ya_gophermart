package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type Claims struct {
	jwt.RegisteredClaims

	Login string
}

type UserController interface {
	GetUserByLogin(context.Context, string) (model.User, error)
}

type Service struct {
	secretKey      config.AuthSecretKey
	userController UserController
	logger         *slog.Logger
}

func NewAuthService(controller UserController, logger *slog.Logger, secretKey config.AuthSecretKey) *Service {
	return &Service{userController: controller, logger: logger, secretKey: secretKey}
}

func (s *Service) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := s.userController.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}
	err = CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidPassword
	}
	return s.generateToken(user, s.secretKey)
}

func (s *Service) generateToken(user model.User, secretKey config.AuthSecretKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": user.Login,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return tokenString, nil
}

func (s *Service) ReadToken(tokenString string, secretKey config.AuthSecretKey) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return "", ErrInvalidToken
	}
	return claims.Login, nil
}
