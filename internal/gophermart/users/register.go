package users

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
)

func RegisterService(pool *pgxpool.Pool, secretKey config.AuthSecretKey, logger *slog.Logger) *service.Service {
	repo := repository.NewStorage(pool)
	return service.NewService(repo, secretKey, logger)
}
