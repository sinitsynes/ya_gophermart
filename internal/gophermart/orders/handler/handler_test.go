package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/service"
)

const validOrderNumber = "4561261212345632"

type stubController struct {
	createFn func(context.Context, model.CreateOrder) (model.Order, error)
	listFn   func(context.Context, int32) ([]model.Order, error)
}

func (c stubController) CreateOrder(ctx context.Context, order model.CreateOrder) (model.Order, error) {
	if c.createFn == nil {
		return model.Order{}, errors.New("CreateOrder not stubbed")
	}
	return c.createFn(ctx, order)
}

func (c stubController) GetOrdersByUser(ctx context.Context, userID int32) ([]model.Order, error) {
	if c.listFn == nil {
		return nil, errors.New("GetOrdersByUser not stubbed")
	}
	return c.listFn(ctx, userID)
}

func withUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, config.UserIDKey, userID)
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name     string
		userID   any
		body     string
		setup    func() stubController
		wantCode int
	}{
		{
			name:   "Created",
			userID: int32(1),
			body:   validOrderNumber,
			setup: func() stubController {
				return stubController{
					createFn: func(_ context.Context, order model.CreateOrder) (model.Order, error) {
						return model.Order{ID: 10, UserID: 1, Number: order.Number}, nil
					},
				}
			},
			wantCode: http.StatusCreated,
		},
		{
			name:   "AlreadyExistsSameUser",
			userID: int32(1),
			body:   validOrderNumber,
			setup: func() stubController {
				return stubController{
					createFn: func(context.Context, model.CreateOrder) (model.Order, error) {
						return model.Order{UserID: 1, Number: validOrderNumber}, service.ErrOrderAlreadyExists
					},
				}
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "AlreadyExistsOtherUser",
			userID: int32(1),
			body:   validOrderNumber,
			setup: func() stubController {
				return stubController{
					createFn: func(context.Context, model.CreateOrder) (model.Order, error) {
						return model.Order{UserID: 2, Number: validOrderNumber}, service.ErrOrderAlreadyExists
					},
				}
			},
			wantCode: http.StatusConflict,
		},
		{
			name:     "Unauthorized",
			userID:   nil,
			body:     validOrderNumber,
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "InvalidLuhn",
			userID:   int32(1),
			body:     "123",
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if test.userID != nil {
				ctx = withUserID(ctx, test.userID.(int32))
			}

			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/api/user/orders",
				bytes.NewBufferString(test.body),
			)
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.CreateOrder(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
		})
	}
}

func TestGetOrdersByUser(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name     string
		userID   any
		setup    func() stubController
		wantCode int
		wantBody []model.Order
	}{
		{
			name:   "OK",
			userID: int32(1),
			setup: func() stubController {
				return stubController{
					listFn: func(_ context.Context, userID int32) ([]model.Order, error) {
						assert.Equal(t, int32(1), userID)
						return []model.Order{
							{Number: "111", Status: model.OrderNew, UploadedAt: now},
							{Number: "222", Status: model.OrderProcessed, UploadedAt: now},
						}, nil
					},
				}
			},
			wantCode: http.StatusOK,
			wantBody: []model.Order{
				{Number: "111", Status: model.OrderNew, UploadedAt: now},
				{Number: "222", Status: model.OrderProcessed, UploadedAt: now},
			},
		},
		{
			name:   "NoContent",
			userID: int32(1),
			setup: func() stubController {
				return stubController{
					listFn: func(context.Context, int32) ([]model.Order, error) {
						return nil, service.ErrNoOrders
					},
				}
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:   "InternalError",
			userID: int32(1),
			setup: func() stubController {
				return stubController{
					listFn: func(context.Context, int32) ([]model.Order, error) {
						return nil, errors.New("db down")
					},
				}
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Unauthorized",
			userID:   nil,
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if test.userID != nil {
				ctx = withUserID(ctx, test.userID.(int32))
			}

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/user/orders", nil)
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.GetOrdersByUser(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
			if test.wantBody != nil {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				wantJSON, err := json.Marshal(test.wantBody)
				require.NoError(t, err)
				assert.JSONEq(t, string(wantJSON), rr.Body.String())
			}
		})
	}
}
