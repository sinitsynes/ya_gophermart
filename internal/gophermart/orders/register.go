package orders

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/service"
	userRepo "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/repository"
)

func RegisterService(pool *pgxpool.Pool, logger *slog.Logger) *service.Service {
	orderStore := repository.NewStorage(pool)
	userStore := userRepo.NewStorage(pool)
	return service.NewService(orderStore, userStore, logger)
}
