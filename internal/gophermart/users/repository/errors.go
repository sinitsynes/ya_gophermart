package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
)

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no rows: %w", repoErr.ErrNotFound)
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "users_login_key" {
				return &repoErr.RepoError{Err: repoErr.ErrAlreadyExists, Message: pgErr.Detail}
			}
			return &repoErr.RepoError{Err: repoErr.ErrConflict, Message: pgErr.Detail}
		default:
			return fmt.Errorf("unknown error: %w", err)
		}
	}
	return err
}
