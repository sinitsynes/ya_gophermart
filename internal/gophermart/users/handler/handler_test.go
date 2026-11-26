package handler_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/model"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/service"
)

type stubController struct {
	createFn func(context.Context, model.UserInput) (model.User, error)
	loginFn  func(context.Context, string, string) (string, error)
}

func (c stubController) CreateUser(ctx context.Context, input model.UserInput) (model.User, error) {
	if c.createFn == nil {
		return model.User{}, errors.New("CreateUser not stubbed")
	}
	return c.createFn(ctx, input)
}

func (c stubController) Login(ctx context.Context, login, password string) (string, error) {
	if c.loginFn == nil {
		return "", errors.New("Login not stubbed")
	}
	return c.loginFn(ctx, login, password)
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	requestBody := `{"login": "user", "password": "password"}`

	tests := []struct {
		name     string
		body     string
		setup    func() stubController
		wantCode int
		wantBody string
	}{
		{
			name: "OK",
			body: requestBody,
			setup: func() stubController {
				return stubController{
					createFn: func(_ context.Context, _ model.UserInput) (model.User, error) {
						return model.User{ID: 1, Login: "user"}, nil
					},
					loginFn: func(_ context.Context, _, _ string) (string, error) {
						return "test-token", nil
					},
				}
			},
			wantCode: http.StatusOK,
			wantBody: "test-token",
		},
		{
			name: "AlreadyExists",
			body: requestBody,
			setup: func() stubController {
				return stubController{
					createFn: func(_ context.Context, _ model.UserInput) (model.User, error) {
						return model.User{}, service.ErrUserAlreadyExists
					},
				}
			},
			wantCode: http.StatusConflict,
		},
		{
			name:     "InvalidJSON",
			body:     "{",
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			body := []byte(test.body)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/user/register", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.RegisterUser(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	requestBody := `{"login": "user", "password": "password"}`

	tests := []struct {
		name     string
		body     string
		setup    func() stubController
		wantCode int
		wantBody string
	}{
		{
			name: "OK",
			body: requestBody,
			setup: func() stubController {
				return stubController{
					loginFn: func(_ context.Context, login, password string) (string, error) {
						assert.Equal(t, "user", login)
						assert.Equal(t, "password", password)
						return "jwt-token", nil
					},
				}
			},
			wantCode: http.StatusOK,
			wantBody: "jwt-token",
		},
		{
			name: "InvalidPassword",
			body: requestBody,
			setup: func() stubController {
				return stubController{
					loginFn: func(_ context.Context, _, _ string) (string, error) {
						return "", service.ErrInvalidPassword
					},
				}
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "InvalidJSON",
			body:     "not-json",
			setup:    func() stubController { return stubController{} },
			wantCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/user/login", bytes.NewReader([]byte(test.body)))
			rr := httptest.NewRecorder()

			h := handler.NewHandler(test.setup(), logger)
			h.Login(rr, req)

			assert.Equal(t, test.wantCode, rr.Code)
		})
	}
}
