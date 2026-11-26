package service

import (
	"context"
	"errors"
	"log/slog"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/model"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrNoWithdrawals     = errors.New("no withdrawals")
)

type Repository interface {
	CreateWithdrawal(context.Context, int32, string, float32) error
	GetBalanceByUser(context.Context, int32) (model.Balance, error)
	GetWithdrawalsByUser(context.Context, int32) ([]model.Withdrawal, error)
}

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateWithdrawal(ctx context.Context, userID int32, order string, sum float32) error {
	current, err := s.repo.GetBalanceByUser(ctx, userID)
	if err != nil {
		return err
	}
	if current.Current < sum {
		return ErrInsufficientFunds
	}
	return s.repo.CreateWithdrawal(ctx, userID, order, sum)
}

func (s *Service) GetBalanceByUser(ctx context.Context, userID int32) (model.Balance, error) {
	return s.repo.GetBalanceByUser(ctx, userID)
}

func (s *Service) GetWithdrawalsByUser(ctx context.Context, userID int32) ([]model.Withdrawal, error) {
	withdrawals, err := s.repo.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, repoErr.ErrNotFound) {
			return nil, ErrNoWithdrawals
		}
		return nil, err
	}

	return withdrawals, nil
}
