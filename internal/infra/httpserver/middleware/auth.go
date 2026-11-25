package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/config"
	users "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
)

func resolveUserLogin(
	r *http.Request,
	userService *users.Service,
) (int32, error) {
	userHeader := r.Header.Get("Authorization")
	if userHeader == "" {
		return 0, errors.New("user header is empty")
	}
	login, err := userService.ReadToken(userHeader)
	if err != nil {
		return 0, err
	}
	user, err := userService.GetUserByLogin(r.Context(), login)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func JWTAuth(
	userService *users.Service,
	userIDKey config.UserIdentifier,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := resolveUserLogin(r, userService)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
