package main

import (
	"context"
	"log"

	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth"
	authHandler "github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders"
	ordersHandler "github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/handler"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users"
	userHandler "github.com/sinitsynes/ya_gophermart/internal/gophermart/users/handler"

	"github.com/sinitsynes/ya_gophermart/internal/infra/db"
	"github.com/sinitsynes/ya_gophermart/internal/infra/httpserver"
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

	userSvc := users.RegisterService(pool, logger)
	authSvc := auth.RegisterService(userSvc, logger, settings.SecretKey)

	usersH := userHandler.NewHandler(userSvc, authSvc, logger)
	authH := authHandler.NewHandler(authSvc, logger)

	ordersSvc := orders.RegisterService(pool, logger)
	ordersH := ordersHandler.NewHandler(ordersSvc, logger)

	mainRouter := gophermart.MainRouter(logger)
	mainRouter.Route("/api/user/", func(r chi.Router) {
		r.Mount("/", users.NewRouter(usersH))
		r.Mount("/", auth.NewRouter(authH))
		r.Mount("/", orders.NewRouter(ordersH))
	})

	server := httpserver.NewServer(settings.Address, mainRouter)
	err = server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
