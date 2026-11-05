package auth

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
)

func RegisterAuthService(pool *pgxpool.Pool, logger *slog.Logger, secretKey config.AuthSecretKey) *handler.Handler {
	repo := repository.NewStorage(pool)
	controller := service.NewAuthService(repo, logger, secretKey)
	handler := handler.NewHandler(controller, logger)
	return handler
}
