package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/errs/repository"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

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

func createRandomUser(t *testing.T, repo *repository.Storage) model.User {
	t.Helper()
	login := testutil.RandomString(10)
	password := testutil.RandomString(10)
	user, err := repo.CreateUser(t.Context(), login, password)
	require.NoError(t, err)
	return user
}

func TestCreateUser(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)
	tests := []struct {
		name  string
		setup func(t *testing.T) (login, password string)
		err   error
	}{
		{
			name: "OK",
			setup: func(_ *testing.T) (string, string) {
				return testutil.RandomString(10), testutil.RandomString(10)
			},
			err: nil,
		},
		{
			name: "LoginNotUnique",
			setup: func(t *testing.T) (string, string) {
				existing := createRandomUser(t, repo)
				return existing.Login, testutil.RandomString(10)
			},
			err: repoErr.ErrUniqueViolation,
		},
	}
	for _, test := range tests { //nolint:paralleltest
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := pgCtr.Restore(ctx)
				require.NoError(t, err)
				pool.Reset()
			})
			login, password := test.setup(t)
			got, err := repo.CreateUser(ctx, login, password)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				return
			}
			assert.Equal(t, login, got.Login)
		})
	}
}

func TestGetUser(t *testing.T) { //nolint:paralleltest
	ctx := t.Context()
	repo := repository.NewStorage(pool)
	tests := []struct {
		name  string
		setup func(t *testing.T) model.User
		err   error
	}{
		{
			name: "OK",
			setup: func(t *testing.T) model.User {
				return createRandomUser(t, repo)
			},
			err: nil,
		},
		{
			name: "NotFound",
			setup: func(_ *testing.T) model.User {
				return model.User{}
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
			user := test.setup(t)
			got, err := repo.GetUserByLogin(ctx, user.Login)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				return
			}
			assert.Equal(t, user.Login, got.Login)
		})
	}
}
