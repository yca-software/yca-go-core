package yca_repository_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	yca_error "github.com/yca-software/yca-go-core/error"
	yca_repository "github.com/yca-software/yca-go-core/repository"
)

type ErrorSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorSuite))
}

func (s *ErrorSuite) TestNewRepository_NilDB_Panics() {
	s.Panics(func() {
		yca_repository.NewRepository[struct{}](nil, "products", []string{"id"}, nil)
	})
}

func (s *ErrorSuite) TestNewRepository_EmptyTableName_Panics() {
	s.Panics(func() {
		yca_repository.NewRepository[struct{}](&sqlx.DB{}, "", []string{"id"}, nil)
	})
}

func (s *ErrorSuite) TestNewRepository_EmptyColumns_Panics() {
	db := &sqlx.DB{}
	s.Panics(func() {
		yca_repository.NewRepository[struct{}](db, "products", nil, nil)
	})
	s.Panics(func() {
		yca_repository.NewRepository[struct{}](db, "products", []string{}, nil)
	})
}

func (s *ErrorSuite) TestWrapSQLError_NilReturnsNil() {
	s.Nil(yca_repository.WrapSQLError(nil))
}

func (s *ErrorSuite) TestWrapSQLError_ErrNoRows_ReturnsNotFound() {
	err := yca_repository.WrapSQLError(sql.ErrNoRows)
	s.Require().Error(err)

	var apiErr *yca_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(404, apiErr.StatusCode)
	s.Equal("NotFound", apiErr.ErrorCode)
}

func (s *ErrorSuite) TestWrapSQLError_Conflict_Returns409() {
	err := yca_repository.WrapSQLError(&pgconn.PgError{Code: "23505"})
	s.Require().Error(err)

	var apiErr *yca_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(409, apiErr.StatusCode)
	s.Equal("Conflict", apiErr.ErrorCode)
}

func (s *ErrorSuite) TestWrapSQLError_ForeignKey_Returns422() {
	err := yca_repository.WrapSQLError(&pgconn.PgError{Code: "23503"})
	s.Require().Error(err)

	var apiErr *yca_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(422, apiErr.StatusCode)
	s.Equal("UnprocessableEntity", apiErr.ErrorCode)
}

func (s *ErrorSuite) TestWrapSQLError_GenericErrorPassthrough() {
	inner := errors.New("connection reset")
	err := yca_repository.WrapSQLError(inner)
	s.Require().Error(err)
	s.ErrorIs(err, inner)
}

func (s *ErrorSuite) TestErrNotFoundNoRowsAffected() {
	err := yca_repository.ErrNotFoundNoRowsAffected()
	s.Require().Error(err)

	var apiErr *yca_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(404, apiErr.StatusCode)
	s.Equal("NotFound", apiErr.ErrorCode)
}
