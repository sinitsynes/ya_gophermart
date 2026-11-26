package repository_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/repository"
	ordersModel "github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	ordersRepo "github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/repository"
	userRepo "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/repository"
	"github.com/sinitsynes/ya_gophermart/internal/testutil"
)

var (
	pgCtr *postgres.PostgresContainer //nolint:gochecknoglobals
	pool  *pgxpool.Pool               //nolint:gochecknoglobals
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	newCtr, err := testutil.NewPostgresContainer(ctx, "gophermart")
	if err != nil {
		log.Fatal(err)
	}
	pgCtr = newCtr
	dsn, err := testutil.ConnectionString(ctx, pgCtr)
	if err != nil {
		log.Fatal(err)
	}
	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	_ = pgCtr.Terminate(ctx)
	os.Exit(code)
}

func createUser(t *testing.T) int32 {
	t.Helper()
	user, err := userRepo.NewStorage(pool).CreateUser(t.Context(), testutil.RandomString(10), testutil.RandomString(10))
	require.NoError(t, err)
	return user.ID
}

func seedAccrual(t *testing.T, userID int32, points float32) {
	t.Helper()
	orders := ordersRepo.NewStorage(pool)
	order, err := orders.CreateOrder(t.Context(), ordersModel.CreateOrder{
		UserID: userID,
		Number: testutil.RandomString(12),
		Status: ordersModel.OrderNew,
	})
	require.NoError(t, err)

	accrual := points
	require.NoError(t, orders.UpdateOrdersAccrual(t.Context(), []ordersModel.UpdateOrder{{
		ID:      order.ID,
		Status:  ordersModel.OrderProcessed,
		Accrual: &accrual,
	}}))
}

func TestGetBalanceByUser(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)

	tests := []struct {
		name string
		setup func(t *testing.T) int32
		want  model.Balance
	}{
		{
			name: "Empty",
			setup: func(t *testing.T) int32 {
				return createUser(t)
			},
			want: model.Balance{Current: 0, Withdrawn: 0},
		},
		{
			name: "WithAccrualAndWithdrawal",
			setup: func(t *testing.T) int32 {
				userID := createUser(t)
				seedAccrual(t, userID, 100)
				require.NoError(t, repo.CreateWithdrawal(ctx, userID, "4561261212345632", 30))
				return userID
			},
			want: model.Balance{Current: 70, Withdrawn: 30},
		},
	}

	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				require.NoError(t, pgCtr.Restore(context.Background()))
				pool.Reset()
			})

			userID := test.setup(t)
			got, err := repo.GetBalanceByUser(ctx, userID)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestCreateWithdrawal(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)

	tests := []struct {
		name  string
		setup func(t *testing.T) (userID int32, order string, sum float32)
		want  model.Withdrawal
	}{
		{
			name: "OK",
			setup: func(t *testing.T) (int32, string, float32) {
				return createUser(t), "4561261212345632", 15.5
			},
			want: model.Withdrawal{
				Order: "4561261212345632",
				Sum:   15.5,
			},
		},
	}

	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				require.NoError(t, pgCtr.Restore(context.Background()))
				pool.Reset()
			})

			userID, order, sum := test.setup(t)
			require.NoError(t, repo.CreateWithdrawal(ctx, userID, order, sum))

			list, err := repo.GetWithdrawalsByUser(ctx, userID)
			require.NoError(t, err)
			require.Len(t, list, 1)
			assert.Equal(t, test.want.Order, list[0].Order)
			assert.Equal(t, test.want.Sum, list[0].Sum)
			assert.False(t, list[0].ProcessedAt.IsZero())
		})
	}
}

func TestGetWithdrawalsByUser(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)

	tests := []struct {
		name    string
		setup   func(t *testing.T) int32
		wantLen int
		wantErr error
		check   func(t *testing.T, got []model.Withdrawal)
	}{
		{
			name: "NotFound",
			setup: func(t *testing.T) int32 {
				return createUser(t)
			},
			wantErr: repoErr.ErrNotFound,
		},
		{
			name: "OrderedByProcessedAtDesc",
			setup: func(t *testing.T) int32 {
				userID := createUser(t)
				require.NoError(t, repo.CreateWithdrawal(ctx, userID, "4561261212345632", 10))
				require.NoError(t, repo.CreateWithdrawal(ctx, userID, "4561261212345640", 20))
				return userID
			},
			wantLen: 2,
			check: func(t *testing.T, got []model.Withdrawal) {
				t.Helper()
				assert.True(t, !got[0].ProcessedAt.Before(got[1].ProcessedAt))
			},
		},
	}

	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				require.NoError(t, pgCtr.Restore(context.Background()))
				pool.Reset()
			})

			userID := test.setup(t)
			got, err := repo.GetWithdrawalsByUser(ctx, userID)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, test.wantLen)
			if test.check != nil {
				test.check(t, got)
			}
		})
	}
}
