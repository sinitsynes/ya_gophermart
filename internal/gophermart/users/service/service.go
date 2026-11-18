package service

import (
	"context"
	"log/slog"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	auth "github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
)

var (
	ErrUserNotFound      = repoErr.ErrNotFound
	ErrUserAlreadyExists = repoErr.ErrAlreadyExists
)

type UserRepository interface {
	CreateUser(context.Context, string, string) (model.User, error)
	GetUserByLogin(context.Context, string) (model.User, error)
}

type Service struct {
	repo   UserRepository
	logger *slog.Logger
}

func NewUserService(repo UserRepository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateUser(ctx context.Context, userIn model.UserInput) (model.User, error) {
	hashPassword, err := auth.HashPassword(userIn.Password)
	if err != nil {
		return model.User{}, err
	}
	return s.repo.CreateUser(ctx, userIn.Login, hashPassword)
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}
