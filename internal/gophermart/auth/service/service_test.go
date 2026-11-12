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
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
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

func setupController(repo service.AuthRepository) (*service.Service, config.AuthSecretKey) {
	secretKey := config.AuthSecretKey(testutil.RandomString(10))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return service.NewAuthService(repo, logger, secretKey), secretKey
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := []struct {
		name     string
		setup    func() service.AuthRepository
		incoming model.UserInput
		want     model.User
		err      error
	}{
		{
			name: "OK",
			setup: func() service.AuthRepository {
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
			setup: func() service.AuthRepository {
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
			c, _ := setupController(repo)
			got, err := c.CreateUser(ctx, test.incoming)
			if err != nil {
				assert.ErrorIs(t, err, test.err)
				return
			}
			assert.Equal(t, test.want.Login, got.Login)
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := []struct {
		name     string
		setup    func() service.AuthRepository
		incoming model.UserInput
		want     string
		err      error
	}{
		{
			name: "OK",
			setup: func() service.AuthRepository {
				return &mockRepository{
					getFn: func(_ context.Context, _ string) (model.User, error) {
						passwordHash, _ := bcrypt.GenerateFromPassword([]byte("PasswordOK"), bcrypt.DefaultCost)
						return model.User{ID: 1, Login: "LoginOK", PasswordHash: string(passwordHash)}, nil
					},
				}
			},
			incoming: model.UserInput{Login: "LoginOK", Password: "PasswordOK"},
			want:     "LoginOK",
			err:      nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo := test.setup()
			c, secretKey := setupController(repo)
			got, err := c.Login(ctx, test.incoming.Login, test.incoming.Password)
			if err != nil {
				assert.ErrorIs(t, err, test.err)
				return
			}
			login, err := c.ReadToken(got, secretKey)
			require.NoError(t, err)
			assert.Equal(t, test.want, login)
		})
	}
}
