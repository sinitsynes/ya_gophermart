package balance

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/service"
)

func RegisterService(pool *pgxpool.Pool, logger *slog.Logger) *service.Service {
	store := repository.NewStorage(pool)
	return service.NewService(store, logger)
}
