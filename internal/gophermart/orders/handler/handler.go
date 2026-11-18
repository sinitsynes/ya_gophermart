package handler

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/service"
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
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			h.logger.Error("Превышен максимальный размер запроса", "error", err)
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		h.logger.Error("Ошибка чтения запроса", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	currentUserID, ok := ctx.Value(config.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !service.LuhnChecksum(string(bodyBytes)) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	toCreate := model.CreateOrder{
		Number: string(bodyBytes),
		UserID: currentUserID,
		Status: model.OrderStatusNew,
	}
	order, err := h.controller.CreateOrder(ctx, toCreate)

	if err != nil {
		if errors.Is(err, service.ErrOrderAlreadyExists) {
			if order.UserID == currentUserID {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.logger.Error("Ошибка создания заказа", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
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
		h.logger.Error("Ошибка получения заказов", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(orders)
	if err != nil {
		h.logger.Error("Ошибка сериализации", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out) //nolint:errcheck,gosec
	w.WriteHeader(http.StatusOK)
}
