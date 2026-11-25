package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/repository/sqlc"
)

type Storage struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (s *Storage) GetBalanceByUser(ctx context.Context, userID int32) (model.Balance, error) {
	row, err := s.queries.GetBalanceByUser(ctx, userID)
	if err != nil {
		return model.Balance{}, mapError(err)
	}
	return model.Balance{
		Current:   row.Current,
		Withdrawn: row.Withdrawn,
	}, nil
}

func (s *Storage) GetWithdrawalsByUser(ctx context.Context, userID int32) ([]model.Withdrawal, error) {
	rows, err := s.queries.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}
	if len(rows) == 0 {
		return nil, repoErr.ErrNotFound
	}
	result := make([]model.Withdrawal, len(rows))
	for i, row := range rows {
		result[i] = model.Withdrawal{
			Order:       row.Number,
			Sum:         row.Sum,
			ProcessedAt: row.ProcessedAt,
		}
	}
	return result, nil
}

func (s *Storage) CreateWithdrawal(ctx context.Context, userID int32, order string, sum float32) error {
	err := s.queries.CreateWithdrawal(ctx, sqlc.CreateWithdrawalParams{UserID: userID, Number: order, Sum: sum})
	return mapError(err)
}
