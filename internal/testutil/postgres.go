package testutil

import (
	"context"
	"errors"
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/sinitsynes/ya_gophermart/internal/infra/db"
)

var (
	ErrCantRunContainer = errors.New("failed to run postgres container")
	ErrDSN              = errors.New("failed to get dsn")
	ErrSnapshot         = errors.New("failed to snapshot container")
)

// NewPostgresContainer создает новый контейнер Postgres,
// применяет миграции и делает снэпшот для быстрого восстановления состояния БД между тестами.
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

	dsn, err := ConnectionString(ctx, ctr)
	if err != nil {
		return nil, err
	}
	if migrationsErr := db.RunMigrations(dsn); migrationsErr != nil {
		return nil, fmt.Errorf("%w", migrationsErr)
	}
	if snapshotErr := ctr.Snapshot(ctx); snapshotErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrSnapshot, snapshotErr)
	}

	return ctr, nil
}

func ConnectionString(ctx context.Context, ctr *postgres.PostgresContainer) (string, error) {
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDSN, err)
	}
	return dsn, nil
}
