package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	repoErr "github.com/sinitsynes/ya_gophermart/internal/err/repository"
)

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no rows: %w", repoErr.ErrNotFound)
	}
	return err
}
