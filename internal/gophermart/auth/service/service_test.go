package service_test

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/service"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
	users "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sinitsynes/ya_gophermart/internal/testutil"
)

type mockUserController struct {
	getUserByLoginFn func(ctx context.Context, username string) (model.User, error)
}

func (m *mockUserController) GetUserByLogin(ctx context.Context, username string) (model.User, error) {
	return m.getUserByLoginFn(ctx, username)
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost) // MinCost keeps tests fast
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
			userErr:  users.ErrUserNotFound,
			login:    "ghost",
			password: "whatever",
			wantErr:  users.ErrUserNotFound,
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
			secretKey := config.AuthSecretKey(testutil.RandomString(40))
			svc := service.NewAuthService(&mockUserController{
				getUserByLoginFn: func(_ context.Context, _ string) (model.User, error) {
					return test.user, test.userErr
				},
			}, nil, secretKey)

			token, err := svc.Login(ctx, test.login, test.password)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				assert.Empty(t, token)
				return
			}
			require.NoError(t, err)
			got, err := svc.ReadToken(token, secretKey)
			require.NoError(t, err)
			assert.Equal(t, test.wantLogin, got)
		})
	}
}
