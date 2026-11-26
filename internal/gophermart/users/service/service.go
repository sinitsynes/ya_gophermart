package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"

	"github.com/sinitsynes/ya_gophermart/config"
	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
)

var (
	ErrUserNotFound      = repoErr.ErrNotFound
	ErrUserAlreadyExists = repoErr.ErrAlreadyExists
	ErrInvalidToken      = errors.New("invalid token")
)

type Claims struct {
	jwt.RegisteredClaims

	Login string `json:"login"`
}

type UserRepository interface {
	CreateUser(context.Context, string, string) (model.User, error)
	GetUserByLogin(context.Context, string) (model.User, error)
}

type Service struct {
	repo      UserRepository
	secretKey config.AuthSecretKey
	logger    *slog.Logger
}

func NewService(repo UserRepository, secretKey config.AuthSecretKey, logger *slog.Logger) *Service {
	return &Service{repo: repo, secretKey: secretKey, logger: logger}
}

func (s *Service) CreateUser(ctx context.Context, userIn model.UserInput) (model.User, error) {
	hashPassword, err := hashPassword(userIn.Password)
	if err != nil {
		return model.User{}, err
	}
	return s.repo.CreateUser(ctx, userIn.Login, hashPassword)
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

func (s *Service) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}
	err = checkPassword(user.PasswordHash, password)
	if err != nil {
		return "", ErrInvalidPassword
	}
	return s.generateToken(user, s.secretKey)
}

func (s *Service) generateToken(user model.User, secretKey config.AuthSecretKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Login: user.Login,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return tokenString, nil
}

func (s *Service) ReadToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return "", ErrInvalidToken
	}
	return claims.Login, nil
}
