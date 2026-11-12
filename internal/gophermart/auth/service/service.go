package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/model"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidToken    = errors.New("invalid token")
	ErrPasswordTooLong = errors.New("password too long")
)

type Claims struct {
	jwt.RegisteredClaims

	Login string
}

type AuthRepository interface {
	CreateUser(context.Context, string, string) (model.User, error)
	GetUserByLogin(context.Context, string) (model.User, error)
}

type Service struct {
	secretKey config.AuthSecretKey
	repo      AuthRepository
	logger    *slog.Logger
}

func NewAuthService(repo AuthRepository, logger *slog.Logger, secretKey config.AuthSecretKey) *Service {
	return &Service{repo: repo, logger: logger, secretKey: secretKey}
}

func (s *Service) CreateUser(ctx context.Context, userIn model.UserInput) (model.User, error) {
	hashPassword, err := s.hashPassword(userIn.Password)
	if err != nil {
		return model.User{}, err
	}
	user, err := s.repo.CreateUser(ctx, userIn.Login, hashPassword)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, login string, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	err = s.checkPassword(user.PasswordHash, password)
	if err != nil {
		return "", err
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

func (s *Service) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", ErrPasswordTooLong
		}
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

func (s *Service) checkPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return ErrInvalidPassword
		case errors.Is(err, bcrypt.ErrHashTooShort):
			return fmt.Errorf("hash too short: %d", len(hashedPassword))
		default:
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}
