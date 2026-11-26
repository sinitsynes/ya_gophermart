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
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/service"
)

const validOrderNumber = "4561261212345632"

type stubController struct {
	createFn     func(context.Context, int32, string, float32) error
	getBalanceFn func(context.Context, int32) (model.Balance, error)
	listFn       func(context.Context, int32) ([]model.Withdrawal, error)
}

func (c stubController) CreateWithdrawal(ctx context.Context, userID int32, order string, sum float32) error {
	if c.createFn == nil {
		return errors.New("CreateWithdrawal not stubbed")
	}
	return c.createFn(ctx, userID, order, sum)
}

func (c stubController) GetBalanceByUser(ctx context.Context, userID int32) (model.Balance, error) {
	if c.getBalanceFn == nil {
		return model.Balance{}, errors.New("GetBalanceByUser not stubbed")
	}
	return c.getBalanceFn(ctx, userID)
}

func (c stubController) GetWithdrawalsByUser(ctx context.Context, userID int32) ([]model.Withdrawal, error) {
	if c.listFn == nil {
		return nil, errors.New("GetWithdrawalsByUser not stubbed")
	}
	return c.listFn(ctx, userID)
}

func withUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, config.UserIDKey, userID)
}

func TestCreateWithdrawal(t *testing.T) {
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
			name:   "OK",
			userID: int32(1),
			body:   `{"order":"4561261212345632","sum":50}`,
			setup: func() stubController {
				return stubController{
					createFn: func(_ context.Context, userID int32, order string, sum float32) error {
						assert.Equal(t, int32(1), userID)
						assert.Equal(t, validOrderNumber, order)
						assert.Equal(t, float32(50), sum)
						return nil
					},
				}
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "InsufficientFunds",
			userID: int32(1),
			body:   `{"order":"4561261212345632","sum":50}`,
			setup: func() stubController {
				return stubController{
					createFn: func(context.Context, int32, string, float32) error {
						return service.ErrInsufficientFunds
					},
				}
			},
			wantCode: http.StatusPaymentRequired,
		},
		{
			name:     "Unauthorized",
			userID:   nil,
			body:     `{"order":"4561261212345632","sum":50}`,
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "InvalidLuhn",
			userID:   int32(1),
			body:     `{"order":"123","sum":50}`,
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:     "InvalidJSON",
			userID:   int32(1),
			body:     `{`,
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusBadRequest,
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
				"/api/user/balance/withdraw",
				bytes.NewBufferString(test.body),
			)
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.CreateWithdrawal(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
		})
	}
}

func TestGetBalanceByUser(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name     string
		userID   any
		setup    func() stubController
		wantCode int
		wantBody *model.Balance
	}{
		{
			name:   "OK",
			userID: int32(1),
			setup: func() stubController {
				return stubController{
					getBalanceFn: func(_ context.Context, userID int32) (model.Balance, error) {
						assert.Equal(t, int32(1), userID)
						return model.Balance{Current: 100.5, Withdrawn: 25}, nil
					},
				}
			},
			wantCode: http.StatusOK,
			wantBody: &model.Balance{Current: 100.5, Withdrawn: 25},
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

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/user/balance", nil)
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.GetBalanceByUser(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
			if test.wantBody != nil {
				wantJSON, err := json.Marshal(test.wantBody)
				require.NoError(t, err)
				assert.JSONEq(t, string(wantJSON), rr.Body.String())
			}
		})
	}
}

func TestGetWithdrawalsByUser(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.DiscardHandler)
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name     string
		userID   any
		setup    func() stubController
		wantCode int
		wantBody []model.Withdrawal
	}{
		{
			name:   "OK",
			userID: int32(1),
			setup: func() stubController {
				return stubController{
					listFn: func(_ context.Context, userID int32) ([]model.Withdrawal, error) {
						assert.Equal(t, int32(1), userID)
						return []model.Withdrawal{
							{Order: "111", Sum: 10, ProcessedAt: now},
							{Order: "222", Sum: 20, ProcessedAt: now},
						}, nil
					},
				}
			},
			wantCode: http.StatusOK,
			wantBody: []model.Withdrawal{
				{Order: "111", Sum: 10, ProcessedAt: now},
				{Order: "222", Sum: 20, ProcessedAt: now},
			},
		},
		{
			name:   "NoContent",
			userID: int32(1),
			setup: func() stubController {
				return stubController{
					listFn: func(context.Context, int32) ([]model.Withdrawal, error) {
						return nil, service.ErrNoWithdrawals
					},
				}
			},
			wantCode: http.StatusNoContent,
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

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/user/withdrawals", nil)
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.GetWithdrawalsByUser(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
			if test.wantBody != nil {
				wantJSON, err := json.Marshal(test.wantBody)
				require.NoError(t, err)
				assert.JSONEq(t, string(wantJSON), rr.Body.String())
			}
		})
	}
}
