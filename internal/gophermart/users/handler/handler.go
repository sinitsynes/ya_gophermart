package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
)

type UserController interface {
	CreateUser(context.Context, model.UserInput) (model.User, error)
	Login(context.Context, string, string) (string, error)
}

type Handler struct {
	userController UserController
	logger         *slog.Logger
}

func NewHandler(userController UserController, logger *slog.Logger) *Handler {
	return &Handler{
		userController: userController,
		logger:         logger,
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	input := new(model.UserInput)
	if errUnmarshal := json.Unmarshal(bodyBytes, input); errUnmarshal != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	_, createErr := h.userController.CreateUser(ctx, *input)
	if createErr != nil {
		if errors.Is(createErr, service.ErrUserAlreadyExists) {
			http.Error(w, "Логин уже занят", http.StatusConflict)
			return
		}
		h.logger.ErrorContext(ctx, "register error", "err", createErr)
		http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
		return
	}
	token, authErr := h.userController.Login(ctx, input.Login, input.Password)
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	input := new(model.UserInput)
	if errUnmarshal := json.Unmarshal(bodyBytes, input); errUnmarshal != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	token, authErr := h.userController.Login(ctx, input.Login, input.Password)
	if authErr != nil {
		switch {
		case errors.Is(authErr, service.ErrInvalidPassword):
			http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
		default:
			h.logger.ErrorContext(ctx, "auth error", "err", authErr)
			http.Error(w, "Ошибка аутентификации", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) //nolint:errcheck,gosec
}
