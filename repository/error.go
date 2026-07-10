package yca_repository

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	yca_error "github.com/yca-software/yca-go-core/error"
)

func WrapSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return yca_error.NewNotFoundError(err, "NotFound", nil)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return yca_error.NewConflictError(err, "Conflict", nil)
		case "23503":
			return yca_error.NewUnprocessableEntityError(err, "UnprocessableEntity", nil)
		}
	}

	return err
}

func ErrNotFoundNoRowsAffected() error {
	return yca_error.NewNotFoundError(errors.New("no rows affected"), "NotFound", nil)
}
