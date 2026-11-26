package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/repository/sqlc"
)

type Storage struct {
	queries *sqlc.Queries
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{
		queries: sqlc.New(pool),
	}
}

func (s *Storage) CreateUser(ctx context.Context, login, passwordHash string) (model.User, error) {
	params := sqlc.CreateUserParams{
		Login:        login,
		PasswordHash: passwordHash,
	}
	row, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return model.User{}, mapError(err)
	}
	return model.User{ID: row.ID, Login: row.Login}, nil
}

func (s *Storage) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	row, err := s.queries.GetUserByLogin(ctx, login)
	if err != nil {
		return model.User{}, mapError(err)
	}
	return model.User{ID: row.ID, Login: row.Login, PasswordHash: row.PasswordHash}, nil
}
