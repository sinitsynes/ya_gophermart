package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/sinitsynes/ya_gophermart/config"
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
	secretKey := config.AuthSecretKey(testutil.RandomString(40))
	return service.NewService(repo, secretKey, logger)
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
			err:      service.ErrPasswordTooLong,
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

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

func TestLogin(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	tests := []struct {
		name      string
		user      model.User
		userErr   error
		login     string
		password  string
		wantLogin string
		wantErr   error
	}{
		{
			name:      "OK",
			user:      model.User{ID: 1, Login: "alice", PasswordHash: mustHash(t, "secret")},
			login:     "alice",
			password:  "secret",
			wantLogin: "alice",
		},
		{
			name:     "NOT FOUND",
			userErr:  service.ErrUserNotFound,
			login:    "ghost",
			password: "whatever",
			wantErr:  service.ErrUserNotFound,
		},
		{
			name:     "WRONG PASSWORD",
			user:     model.User{ID: 1, Login: "alice", PasswordHash: mustHash(t, "secret")},
			login:    "alice",
			password: "nope",
			wantErr:  service.ErrInvalidPassword,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			svc := setupController(&mockRepository{
				getFn: func(_ context.Context, _ string) (model.User, error) {
					return test.user, test.userErr
				},
			})

			token, err := svc.Login(ctx, test.login, test.password)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			got, err := svc.ReadToken(token)
			require.NoError(t, err)
			assert.Equal(t, test.wantLogin, got)
		})
	}
}
