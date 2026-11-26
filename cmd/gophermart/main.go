package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance"
	balanceHandler "github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders"
	ordersHandler "github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users"
	userHandler "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/handler"

	"github.com/sinitsynes/ya_gophermart/internal/infra/db"
	"github.com/sinitsynes/ya_gophermart/internal/infra/httpserver"
	"github.com/sinitsynes/ya_gophermart/internal/infra/httpserver/middleware"
)

func run(ctx context.Context) error {
	settings, err := config.Load()
	if err != nil {
		return err
	}
	logger := config.NewLogger()

	migrationsErr := db.RunMigrations(settings.Database.DSN)
	if migrationsErr != nil {
		return migrationsErr
	}
	pool, err := db.InitDB(ctx, &settings.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	userSvc := users.RegisterService(pool, settings.SecretKey, logger)
	usersH := userHandler.NewHandler(userSvc, logger)

	ordersSvc := orders.RegisterService(ctx, pool, logger, settings.AccrualSystemAddress)
	ordersH := ordersHandler.NewHandler(ordersSvc, logger)

	balanceSvc := balance.RegisterService(pool, logger)
	balanceH := balanceHandler.NewHandler(balanceSvc, logger)

	r := gophermart.MainRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Mount("/", users.NewRouter(usersH))

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(userSvc, config.UserIDKey))
			r.Mount("/", orders.NewRouter(ordersH))
			r.Mount("/", balance.NewRouter(balanceH))
		})
	})

	server := httpserver.NewServer(settings.Address, r)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return nil
	case channeledErr := <-errCh:
		if errors.Is(channeledErr, http.ErrServerClosed) {
			return nil
		}
		return channeledErr
	}
}

func main() {
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	err := run(appCtx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}
