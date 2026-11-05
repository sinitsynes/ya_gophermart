package main

import (
	"context"
	"log"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth"
	"github.com/sinitsynes/ya_gophermart/internal/infra/db"
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

	auth.RegisterAuthService(pool, logger, settings.SecretKey)

	return nil

}

func main() {
	ctx := context.Background()
	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
