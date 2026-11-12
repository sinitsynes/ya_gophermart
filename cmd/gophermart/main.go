package main

import (
	"context"
	"log"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth"
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
	mux := httpserver.MainMux()
	authMux := auth.NewRouter(auth.RegisterService(pool, logger, settings.SecretKey))
	mux.Handle("/api/user", authMux)
	server := httpserver.NewServer(settings.Address, mux)
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
