package service_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/service"
)

type mockRepo struct {
	getBalanceFn     func(context.Context, int32) (model.Balance, error)
	createWithdrawFn func(context.Context, int32, string, float32) error
	getWithdrawalsFn func(context.Context, int32) ([]model.Withdrawal, error)
}

func (m *mockRepo) GetBalanceByUser(ctx context.Context, userID int32) (model.Balance, error) {
	return m.getBalanceFn(ctx, userID)
}

func (m *mockRepo) CreateWithdrawal(ctx context.Context, userID int32, order string, sum float32) error {
	return m.createWithdrawFn(ctx, userID, order, sum)
}

func (m *mockRepo) GetWithdrawalsByUser(ctx context.Context, userID int32) ([]model.Withdrawal, error) {
	return m.getWithdrawalsFn(ctx, userID)
}

func TestCreateWithdrawal(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name    string
		setup   func() *mockRepo
		userID  int32
		order   string
		sum     float32
		wantErr error
	}{
		{
			name: "OK",
			setup: func() *mockRepo {
				return &mockRepo{
					getBalanceFn: func(_ context.Context, userID int32) (model.Balance, error) {
						assert.Equal(t, int32(1), userID)
						return model.Balance{Current: 100, Withdrawn: 0}, nil
					},
					createWithdrawFn: func(_ context.Context, userID int32, order string, sum float32) error {
						assert.Equal(t, int32(1), userID)
						assert.Equal(t, "4561261212345632", order)
						assert.Equal(t, float32(40), sum)
						return nil
					},
				}
			},
			userID: 1,
			order:  "4561261212345632",
			sum:    40,
		},
		{
			name: "InsufficientFunds",
			setup: func() *mockRepo {
				return &mockRepo{
					getBalanceFn: func(context.Context, int32) (model.Balance, error) {
						return model.Balance{Current: 10, Withdrawn: 0}, nil
					},
					createWithdrawFn: func(context.Context, int32, string, float32) error {
						t.Fatal("CreateWithdrawal must not be called")
						return nil
					},
				}
			},
			userID:  1,
			order:   "4561261212345632",
			sum:     40,
			wantErr: service.ErrInsufficientFunds,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			svc := service.NewService(test.setup(), logger)
			err := svc.CreateWithdrawal(ctx, test.userID, test.order, test.sum)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGetBalanceByUser(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	repo := &mockRepo{
		getBalanceFn: func(_ context.Context, userID int32) (model.Balance, error) {
			assert.Equal(t, int32(5), userID)
			return model.Balance{Current: 80.5, Withdrawn: 19.5}, nil
		},
	}
	svc := service.NewService(repo, logger)

	got, err := svc.GetBalanceByUser(ctx, 5)
	require.NoError(t, err)
	assert.Equal(t, model.Balance{Current: 80.5, Withdrawn: 19.5}, got)
}

func TestGetWithdrawalsByUser(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func() *mockRepo
		want    []model.Withdrawal
		wantErr error
	}{
		{
			name: "OK",
			setup: func() *mockRepo {
				return &mockRepo{
					getWithdrawalsFn: func(_ context.Context, userID int32) ([]model.Withdrawal, error) {
						assert.Equal(t, int32(1), userID)
						return []model.Withdrawal{
							{Order: "1", Sum: 10, ProcessedAt: now},
							{Order: "2", Sum: 20, ProcessedAt: now},
						}, nil
					},
				}
			},
			want: []model.Withdrawal{
				{Order: "1", Sum: 10, ProcessedAt: now},
				{Order: "2", Sum: 20, ProcessedAt: now},
			},
		},
		{
			name: "NoWithdrawals",
			setup: func() *mockRepo {
				return &mockRepo{
					getWithdrawalsFn: func(context.Context, int32) ([]model.Withdrawal, error) {
						return nil, repoErr.ErrNotFound
					},
				}
			},
			wantErr: service.ErrNoWithdrawals,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			svc := service.NewService(test.setup(), logger)
			got, err := svc.GetWithdrawalsByUser(ctx, 1)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
