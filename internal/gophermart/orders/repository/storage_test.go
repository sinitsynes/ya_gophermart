package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/repository"
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
		fmt.Println(err) //nolint:forbidigo
		os.Exit(1)
	}
	pgCtr = newCtr
	dsn, err := testutil.ConnectionString(ctx, pgCtr)
	if err != nil {
		fmt.Println(err) //nolint:forbidigo
		os.Exit(1)
	}
	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Println(err) //nolint:forbidigo
		os.Exit(1)
	}

	code := m.Run()
	pgCtr.Terminate(ctx)
	os.Exit(code)
}

func createRandomUser(t *testing.T) int32 {
	t.Helper()
	userStorage := userRepo.NewStorage(pool)
	user, err := userStorage.CreateUser(t.Context(), testutil.RandomString(10), testutil.RandomString(10))
	require.NoError(t, err)
	return user.ID
}

func createRandomOrder(t *testing.T, repo *repository.Storage, userID int32) model.Order {
	t.Helper()
	r := model.CreateOrder{
		UserID: userID,
		Number: testutil.RandomString(10),
		Status: model.OrderNew,
	}
	created, err := repo.CreateOrder(t.Context(), r)
	require.NoError(t, err)
	return created
}

func TestCreateOrder(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)
	tests := []struct {
		name  string
		setup func() model.CreateOrder
		err   error
	}{
		{
			name: "OK",
			setup: func() model.CreateOrder {
				userID := createRandomUser(t)
				return model.CreateOrder{
					UserID: userID,
					Number: testutil.RandomString(10),
					Status: model.OrderNew,
				}
			},
			err: nil,
		},
	}
	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := pgCtr.Restore(ctx)
				require.NoError(t, err)
				pool.Reset()
			})
			order := test.setup()
			created, err := repo.CreateOrder(ctx, order)
			if err != nil {
				require.ErrorIs(t, err, test.err)
			}
			assert.Equal(t, order.UserID, created.UserID)
			assert.Equal(t, order.Number, created.Number)
			assert.Equal(t, order.Status, created.Status)
		})
	}
}

func TestGetOrdersByUserID(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)
	tests := []struct {
		name  string
		setup func(t *testing.T, userID int32) []model.Order
		err   error
	}{
		{
			name: "OK",
			setup: func(t *testing.T, userID int32) []model.Order {
				orders := make([]model.Order, 25)
				for i := range 25 {
					created := createRandomOrder(t, repo, userID)
					orders[i] = model.Order{
						ID:         created.ID,
						UserID:     created.UserID,
						Number:     created.Number,
						Status:     created.Status,
						UploadedAt: created.UploadedAt,
					}
				}
				return orders
			},
			err: nil,
		},
		{
			name: "NOT FOUND",
			setup: func(_ *testing.T, _ int32) []model.Order {
				return nil
			},
			err: repoErr.ErrNotFound,
		},
	}
	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := pgCtr.Restore(ctx)
				require.NoError(t, err)
				pool.Reset()
			})
			userID := createRandomUser(t)
			want := test.setup(t, userID)
			got, err := repo.GetOrdersByUserID(ctx, userID)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, want, got)
		})
	}
}
