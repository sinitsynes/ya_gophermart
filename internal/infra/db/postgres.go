package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/sinitsynes/ya_gophermart/config"
	"github.com/sinitsynes/ya_gophermart/migrations"
)

func InitDB(ctx context.Context, dbConfig *config.PostgresConfig) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbConfig.DSN)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.ConnectTimeout = dbConfig.ConnectionTimeout

	ctx, cancel := context.WithTimeout(ctx, dbConfig.ConnectionTimeout)
	defer cancel()
	return pgxpool.NewWithConfig(ctx, config)
}

var (
	ErrDialect        = errors.New("failed to set dialect")
	ErrOpenConnection = errors.New("failed to open database")
	ErrRunMigrations  = errors.New("failed to run migrations")
)

func RunMigrations(connString string) error {
	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%w: %w", ErrDialect, err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenConnection, err)
	}
	defer db.Close()

	if migrationsErr := goose.Up(db, "gophermart"); migrationsErr != nil {
		return fmt.Errorf("%w: %w", ErrRunMigrations, migrationsErr)
	}
	return nil
}
