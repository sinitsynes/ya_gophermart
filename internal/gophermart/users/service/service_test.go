package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	auth "github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
	"github.com/sinitsynes/ya_gophermart/internal/testutil"
)

const (
	bcryptMaxLength = 72
)

type mockRepository struct {
	createFn func(ctx context.Context, username, password string) (model.User, error)
	getFn    func(ctx context.Context, username string) (model.User, error)
}

func (m *mockRepository) CreateUser(ctx context.Context, username, password string) (model.User, error) {
	return m.createFn(ctx, username, password)
}

func (m *mockRepository) GetUserByLogin(ctx context.Context, username string) (model.User, error) {
	return m.getFn(ctx, username)
}

func setupController(repo service.UserRepository) *service.Service {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return service.NewUserService(repo, logger)
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := []struct {
		name     string
		setup    func() service.UserRepository
		incoming model.UserInput
		want     model.User
		err      error
	}{
		{
			name: "OK",
			setup: func() service.UserRepository {
				return &mockRepository{
					createFn: func(_ context.Context, login, passwordHash string) (model.User, error) {
						return model.User{ID: 1, Login: login, PasswordHash: passwordHash}, nil
					},
				}
			},
			incoming: model.UserInput{Login: "LoginOK", Password: "PasswordOK"},
			want:     model.User{ID: 1, Login: "LoginOK"},
			err:      nil,
		},
		{
			name: "PasswordTooLong",
			setup: func() service.UserRepository {
				return &mockRepository{
					createFn: func(_ context.Context, _, _ string) (model.User, error) {
						return model.User{}, nil
					},
				}
			},
			incoming: model.UserInput{Login: "LoginOK", Password: testutil.RandomString(bcryptMaxLength + 1)},
			want:     model.User{},
			err:      auth.ErrPasswordTooLong,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo := test.setup()
			c := setupController(repo)
			got, err := c.CreateUser(ctx, test.incoming)
			if err != nil {
				assert.ErrorIs(t, err, test.err)
				return
			}
			assert.Equal(t, test.want.Login, got.Login)
		})
	}
}
