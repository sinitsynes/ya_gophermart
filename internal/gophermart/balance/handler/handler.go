package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/service"
	"github.com/sinitsynes/ya_gophermart/pkg/validation"
)

type Controller interface {
	CreateWithdrawal(context.Context, int32, string, float32) error
	GetBalanceByUser(context.Context, int32) (model.Balance, error)
	GetWithdrawalsByUser(context.Context, int32) ([]model.Withdrawal, error)
}

type Handler struct {
	c      Controller
	logger *slog.Logger
}

func NewHandler(c Controller, logger *slog.Logger) *Handler {
	return &Handler{c: c, logger: logger}
}

func (h *Handler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	input := new(model.Withdrawal)
	if errUnmarshal := json.Unmarshal(bodyBytes, input); errUnmarshal != nil {
		http.Error(w, "Ошибка декодирования запроса", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	currentUserID, ok := ctx.Value(config.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !validation.LuhnChecksum(input.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := h.c.CreateWithdrawal(ctx, currentUserID, input.Order, input.Sum); err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientFunds):
			http.Error(w, "Недостаточно средств", http.StatusPaymentRequired)
		default:
			http.Error(w, "Ошибка создания списания", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetBalanceByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUserID, ok := ctx.Value(config.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	balance, err := h.c.GetBalanceByUser(ctx, currentUserID)
	if err != nil {
		http.Error(w, "Ошибка получения баланса", http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, "Ошибка кодирования ответа", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write((res)) //nolint:errcheck,gosec
}

func (h *Handler) GetWithdrawalsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUserID, ok := ctx.Value(config.UserIDKey).(int32)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	withdrawals, err := h.c.GetWithdrawalsByUser(ctx, currentUserID)
	if err != nil {
		if errors.Is(err, service.ErrNoWithdrawals) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Ошибка получения списаний", http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, "Ошибка кодирования ответа", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write((res)) //nolint:errcheck,gosec
}
