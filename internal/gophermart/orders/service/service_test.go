package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/service"
	users "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
)

type mockOrderRepo struct {
	createFn    func(context.Context, model.CreateOrder) (model.Order, error)
	getByNumFn  func(context.Context, string) (model.Order, error)
	getByUserFn func(context.Context, int32) ([]model.Order, error)
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order model.CreateOrder) (model.Order, error) {
	return m.createFn(ctx, order)
}

func (m *mockOrderRepo) GetOrderByNumber(ctx context.Context, number string) (model.Order, error) {
	return m.getByNumFn(ctx, number)
}

func (m *mockOrderRepo) GetOrdersByUserID(ctx context.Context, userID int32) ([]model.Order, error) {
	return m.getByUserFn(ctx, userID)
}

type mockUserRepo struct{}

func (mockUserRepo) GetUserByLogin(context.Context, string) (users.User, error) {
	return users.User{}, errors.New("not used")
}

func newTestService(t *testing.T, repo service.Repository) *service.Service {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	accrual := service.NewAccrualService(nil, "http://localhost")
	return service.NewService(repo, mockUserRepo{}, accrual, logger)
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	validLuhn := "4561261212345631"
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func() service.Repository
		input   model.CreateOrder
		want    model.Order
		wantErr error
	}{
		{
			name: "OK",
			setup: func() service.Repository {
				return &mockOrderRepo{
					createFn: func(_ context.Context, order model.CreateOrder) (model.Order, error) {
						return model.Order{
							ID:         10,
							UserID:     order.UserID,
							Number:     order.Number,
							Status:     order.Status,
							UploadedAt: now,
						}, nil
					},
				}
			},
			input: model.CreateOrder{
				UserID: 1,
				Number: validLuhn,
				Status: model.OrderNew,
			},
			want: model.Order{
				ID:         10,
				UserID:     1,
				Number:     validLuhn,
				Status:     model.OrderNew,
				UploadedAt: now,
			},
		},
		{
			name: "AlreadyExists",
			setup: func() service.Repository {
				existing := model.Order{
					ID:         7,
					UserID:     2,
					Number:     validLuhn,
					Status:     model.OrderProcessing,
					UploadedAt: now,
				}
				return &mockOrderRepo{
					createFn: func(context.Context, model.CreateOrder) (model.Order, error) {
						return model.Order{}, &repoErr.RepoError{Err: repoErr.ErrAlreadyExists, Message: "duplicate"}
					},
					getByNumFn: func(_ context.Context, number string) (model.Order, error) {
						require.Equal(t, existing.Number, number)
						return existing, nil
					},
				}
			},
			input: model.CreateOrder{
				UserID: 1,
				Number: validLuhn,
				Status: model.OrderNew,
			},
			want: model.Order{
				ID:         7,
				UserID:     2,
				Number:     validLuhn,
				Status:     model.OrderProcessing,
				UploadedAt: now,
			},
			wantErr: service.ErrOrderAlreadyExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(t, test.setup())
			got, err := svc.CreateOrder(ctx, test.input)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, test.wantErr)
				if errors.Is(test.wantErr, service.ErrOrderAlreadyExists) {
					assert.Equal(t, test.want, got)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGetOrdersByUser(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func() service.Repository
		userID  int32
		want    []model.Order
		wantErr error
	}{
		{
			name: "OK",
			setup: func() service.Repository {
				return &mockOrderRepo{
					getByUserFn: func(_ context.Context, userID int32) ([]model.Order, error) {
						assert.Equal(t, int32(1), userID)
						return []model.Order{
							{ID: 1, UserID: 1, Number: "111", Status: model.OrderNew, UploadedAt: now},
							{ID: 2, UserID: 1, Number: "222", Status: model.OrderProcessed, UploadedAt: now},
						}, nil
					},
				}
			},
			userID: 1,
			want: []model.Order{
				{ID: 1, UserID: 1, Number: "111", Status: model.OrderNew, UploadedAt: now},
				{ID: 2, UserID: 1, Number: "222", Status: model.OrderProcessed, UploadedAt: now},
			},
		},
		{
			name: "NoOrders",
			setup: func() service.Repository {
				return &mockOrderRepo{
					getByUserFn: func(context.Context, int32) ([]model.Order, error) {
						return nil, repoErr.ErrNotFound
					},
				}
			},
			userID:  1,
			wantErr: service.ErrNoOrders,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			svc := newTestService(t, test.setup())
			got, err := svc.GetOrdersByUser(ctx, test.userID)
			if test.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
