package service

import (
	"context"
	"errors"
	"log/slog"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	users "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
)

var (
	ErrOrderAlreadyExists = repoErr.ErrAlreadyExists
	ErrNoOrders           = repoErr.ErrNotFound
)

type UserRepository interface {
	GetUserByLogin(context.Context, string) (users.User, error)
}

type OrderRepository interface {
	GetOrdersByUserID(context.Context, int32) ([]model.Order, error)
	CreateOrder(context.Context, model.CreateOrder) (model.Order, error)
	GetOrderByNumber(context.Context, string) (model.Order, error)
}

type Service struct {
	repo     OrderRepository
	userRepo UserRepository
	logger   *slog.Logger
}

func NewService(orderRepo OrderRepository, userRepo UserRepository, logger *slog.Logger) *Service {
	return &Service{
		repo:     orderRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *Service) CreateOrder(ctx context.Context, order model.CreateOrder) (model.Order, error) {
	created, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, repoErr.ErrAlreadyExists) {
			order, getErr := s.repo.GetOrderByNumber(ctx, order.Number)
			if getErr != nil {
				return model.Order{}, getErr
			}
			return order, ErrOrderAlreadyExists
		}
		return model.Order{}, err
	}
	return created, nil
}

func (s *Service) GetOrdersByUser(ctx context.Context, userID int32) ([]model.Order, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repoErr.ErrNotFound) {
			return nil, ErrNoOrders
		}
		return nil, err
	}
	return orders, nil
}
