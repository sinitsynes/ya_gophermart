package orders

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/service"
	userRepo "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/repository"
)

func RegisterService(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger, accrualURL string) *service.Service {
	orderStore := repository.NewStorage(pool)
	userStore := userRepo.NewStorage(pool)
	accrualService := service.NewAccrualService(orderStore, accrualURL)
	accrualService.Start(ctx)
	return service.NewService(orderStore, userStore, accrualService, logger)
}
