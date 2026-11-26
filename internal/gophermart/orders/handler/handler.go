package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/service"
	"github.com/sinitsynes/ya_gophermart/pkg/validation"
)

type Controller interface {
	CreateOrder(context.Context, model.CreateOrder) (model.Order, error)
	GetOrdersByUser(context.Context, int32) ([]model.Order, error)
}

type Handler struct {
	controller Controller
	logger     *slog.Logger
}

func NewHandler(controller Controller, logger *slog.Logger) *Handler {
	return &Handler{
		controller: controller,
		logger:     logger,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	ctx := r.Context()
	currentUserID, ok := ctx.Value(config.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if string(bodyBytes) == "" || !validation.LuhnChecksum(string(bodyBytes)) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	toCreate := model.CreateOrder{
		Number: string(bodyBytes),
		UserID: currentUserID,
		Status: model.OrderNew,
	}
	order, err := h.controller.CreateOrder(ctx, toCreate)

	if err != nil {
		if errors.Is(err, service.ErrOrderAlreadyExists) {
			if order.UserID == currentUserID {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.logger.ErrorContext(ctx, "Ошибка создания заказа", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetOrdersByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(config.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	orders, err := h.controller.GetOrdersByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoOrders) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.logger.ErrorContext(ctx, "Ошибка получения заказов", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(orders)
	if err != nil {
		h.logger.ErrorContext(ctx, "Ошибка сериализации", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(out) //nolint:errcheck,gosec
}
