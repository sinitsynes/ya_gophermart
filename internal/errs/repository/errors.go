package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
)

func MapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no rows: %w", ErrNotFound)
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("unique violation: %w at %s", ErrUniqueViolation, pgErr.ColumnName)
		default:
			return fmt.Errorf("unknown error: %w", err)
		}
	}
	return err
}
