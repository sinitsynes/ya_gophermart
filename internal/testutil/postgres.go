package testutil

import (
	"context"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sinitsynes/ya_gophermart/internal/infra/db"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	ErrCantRunContainer = errors.New("failed to run postgres container")
	ErrDSN              = errors.New("failed to get dsn")
	ErrSnapshot         = errors.New("failed to snapshot container")
)

func NewPostgresContainer(ctx context.Context, dbName string) (*postgres.PostgresContainer, error) {
	ctr, err := postgres.Run(
		ctx, "postgres:18-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbName),
		postgres.WithPassword(dbName),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCantRunContainer, err)
	}
	dsn, dsnErr := ctr.ConnectionString(ctx, "sslmode=disable")
	if dsnErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrDSN, dsnErr)
	}
	if err := db.RunMigrations(dsn); err != nil {
		return nil, err
	}
	if snapshotErr := ctr.Snapshot(ctx); snapshotErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrSnapshot, snapshotErr)
	}

	return ctr, nil
}
