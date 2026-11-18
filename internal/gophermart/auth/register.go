package auth

import (
	"log/slog"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
)

func RegisterService(
	userController service.UserController,
	logger *slog.Logger,
	secretKey config.AuthSecretKey,
) *service.Service {
	return service.NewAuthService(userController, logger, secretKey)
}
