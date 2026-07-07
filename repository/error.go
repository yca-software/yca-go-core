package chi_repository

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	chi_error "github.com/yca-software/yca-go-core/error"
)

func WrapSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return chi_error.NewNotFoundError(err, "NotFound", nil)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return chi_error.NewConflictError(err, "Conflict", nil)
		case "23503":
			return chi_error.NewUnprocessableEntityError(err, "UnprocessableEntity", nil)
		}
	}

	return err
}

func ErrNotFoundNoRowsAffected() error {
	return chi_error.NewNotFoundError(errors.New("no rows affected"), "NotFound", nil)
}
