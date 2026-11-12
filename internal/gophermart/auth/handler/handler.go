package handler

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
)

type Controller interface {
	CreateUser(context.Context, model.UserInput) (model.User, error)
	Login(context.Context, string, string) (string, error)
	ReadToken(string, config.AuthSecretKey) (string, error)
}

type Handler struct {
	controller Controller
	logger     *slog.Logger
}

func NewHandler(service Controller, logger *slog.Logger) *Handler {
	return &Handler{controller: service, logger: logger}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "Превышен максимальный размер запроса", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}
	input := new(model.UserInput)
	if errUnmarshal := json.Unmarshal(bodyBytes, input); errUnmarshal != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	token, authErr := h.controller.Login(ctx, input.Login, input.Password)
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
	_, createErr := h.controller.CreateUser(ctx, *input)
	if createErr != nil {
		h.logger.ErrorContext(ctx, "register error", "err", createErr)
		http.Error(w, "Ошибка регистрации", http.StatusInternalServerError)
		return
	}
	token, authErr := h.controller.Login(ctx, input.Login, input.Password)
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
