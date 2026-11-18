package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sinitsynes/ya_gophermart/config"
	auth "github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
	userStorage "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/repository"
	userService "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
)

func resolveUserLogin(
	r *http.Request,
	pool *pgxpool.Pool,
	userIDKey config.UserIdentifier,
	secretKey config.AuthSecretKey,
	logger *slog.Logger,
) (int32, error) {
	userHeader := r.Header.Get(string(userIDKey))
	if userHeader == "" {
		return 0, errors.New("user header is empty")
	}
	userRepo := userStorage.NewStorage(pool)
	userController := userService.NewUserService(userRepo, logger)
	authController := auth.NewAuthService(userController, logger, secretKey)
	login, err := authController.ReadToken(userHeader, secretKey)
	if err != nil {
		return 0, err
	}
	user, err := userController.GetUserByLogin(r.Context(), login)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func JWTAuth(
	pool *pgxpool.Pool,
	userIDKey config.UserIdentifier,
	secretKey config.AuthSecretKey,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := resolveUserLogin(r, pool, userIDKey, secretKey, logger)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
