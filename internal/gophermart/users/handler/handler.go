package handler

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
	users "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
)

type UserController interface {
	CreateUser(context.Context, model.UserInput) (model.User, error)
}
type AuthController interface {
	Login(context.Context, string, string) (string, error)
}

type Handler struct {
	userController UserController
	authController AuthController
	logger         *slog.Logger
}

func NewHandler(userController UserController, authController AuthController, logger *slog.Logger) *Handler {
	return &Handler{
		userController: userController,
		authController: authController,
		logger:         logger,
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	input := new(model.UserInput)
	if errUnmarshal := json.Unmarshal(bodyBytes, input); errUnmarshal != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	_, createErr := h.userController.CreateUser(ctx, *input)
	if createErr != nil {
		if errors.Is(createErr, users.ErrUserAlreadyExists) {
			http.Error(w, "Логин уже занят", http.StatusConflict)
			return
		}
		h.logger.ErrorContext(ctx, "register error", "err", createErr)
		http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
		return
	}
	token, authErr := h.authController.Login(ctx, input.Login, input.Password)
	if authErr != nil {
		switch {
		case errors.Is(authErr, service.ErrInvalidPassword):
			http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
		default:
			h.logger.ErrorContext(ctx, "auth error", "err", authErr)
			http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) //nolint:errcheck,gosec
}
