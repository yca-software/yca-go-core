package yca_error_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	yca_error "github.com/yca-software/yca-go-core/error"
)

type ErrorSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorSuite))
}

func (s *ErrorSuite) TestError_string() {
	s.Equal("TEST_CODE", yca_error.NewBadRequestError(nil, "TEST_CODE", nil).Error())

	err := yca_error.NewBadRequestError(errors.New("underlying error"), "TEST_CODE", nil)
	s.Equal("underlying error", err.Error())

	err2 := yca_error.NewInternalServerError(nil, "TEST_CODE", nil)
	s.Equal("TEST_CODE", err2.Error())

	err3 := yca_error.NewNotFoundError(errors.New("db error"), "TEST_CODE", nil)
	s.Equal("db error", err3.Error())

	err4 := yca_error.NewBadRequestError(nil, "", nil)
	s.Empty(err4.Error())
}

func (s *ErrorSuite) TestNewInternalServerError() {
	err := yca_error.NewInternalServerError(nil, "", nil)
	s.Equal(http.StatusInternalServerError, err.StatusCode)
	s.Empty(err.ErrorCode)
	s.Nil(err.Err)
	s.Nil(err.Extra)
}

func (s *ErrorSuite) TestNewInternalServerError_customCode() {
	err := yca_error.NewInternalServerError(nil, "CUSTOM_CODE", nil)
	s.Equal(http.StatusInternalServerError, err.StatusCode)
	s.Equal("CUSTOM_CODE", err.ErrorCode)
}

func (s *ErrorSuite) TestWithExtra() {
	extra := map[string]any{"field": "email", "reason": "invalid format"}
	err := yca_error.NewBadRequestError(nil, "INVALID_EMAIL", extra)
	s.Equal(extra, err.Extra)
}

func (s *ErrorSuite) TestWithUnderlyingError() {
	underlying := errors.New("database error")
	err := yca_error.NewInternalServerError(underlying, "DB_ERROR", nil)
	s.Equal(underlying, err.Err)
}

func (s *ErrorSuite) TestJSONMarshaling() {
	err := yca_error.NewBadRequestError(errors.New("validation failed"), "INVALID_INPUT", map[string]any{"field": "email"})

	data, marshalErr := json.Marshal(err)
	s.Require().NoError(marshalErr)

	var result map[string]any
	s.Require().NoError(json.Unmarshal(data, &result))

	s.Equal("INVALID_INPUT", result["errorCode"])
	s.NotContains(result, "message")
	s.Equal(map[string]any{"field": "email"}, result["extra"])
	s.NotContains(result, "statusCode")
	s.NotContains(result, "err")
}

func (s *ErrorSuite) TestJSONMarshaling_withoutExtra() {
	err := yca_error.NewBadRequestError(nil, "TEST_CODE", nil)

	data, marshalErr := json.Marshal(err)
	s.Require().NoError(marshalErr)

	var result map[string]any
	s.Require().NoError(json.Unmarshal(data, &result))

	s.NotContains(result, "extra")
	s.Equal("TEST_CODE", result["errorCode"])
	s.NotContains(result, "message")
}

func (s *ErrorSuite) TestJSONMarshaling_emptyErrorCode() {
	err := yca_error.NewBadRequestError(nil, "", nil)

	data, marshalErr := json.Marshal(err)
	s.Require().NoError(marshalErr)

	var result map[string]any
	s.Require().NoError(json.Unmarshal(data, &result))

	s.Equal("", result["errorCode"])
	s.NotContains(result, "message")
}

func (s *ErrorSuite) TestJSONMarshaling_messageFromUnderlyingError() {
	err := yca_error.NewNotFoundError(errors.New("user not found"), "", nil)

	data, marshalErr := json.Marshal(err)
	s.Require().NoError(marshalErr)

	var result map[string]any
	s.Require().NoError(json.Unmarshal(data, &result))

	s.Equal("", result["errorCode"])
	s.Equal("user not found", result["message"])
}

func (s *ErrorSuite) TestJSONMarshaling_5xxDoesNotExposeInternalError() {
	err := yca_error.NewInternalServerError(errors.New("database connection failed"), "InternalServerError", nil)

	data, marshalErr := json.Marshal(err)
	s.Require().NoError(marshalErr)

	var result map[string]any
	s.Require().NoError(json.Unmarshal(data, &result))

	s.Equal("InternalServerError", result["errorCode"])
	s.NotContains(result, "message")
}

func (s *ErrorSuite) TestJSONMarshaling_5xxWithoutErrorCode() {
	err := yca_error.NewInternalServerError(errors.New("secret internal detail"), "", nil)

	data, marshalErr := json.Marshal(err)
	s.Require().NoError(marshalErr)

	var result map[string]any
	s.Require().NoError(json.Unmarshal(data, &result))

	s.Equal("", result["errorCode"])
	s.NotContains(result, "message")
}

func (s *ErrorSuite) TestAllParameterCombinations() {
	cases := []struct {
		name      string
		create    func(error, string, any) *yca_error.Error
		err       error
		errorCode string
		extra     any
	}{
		{"NilErr_EmptyCode_NilExtra", yca_error.NewBadRequestError, nil, "", nil},
		{"NilErr_CustomCode_NilExtra", yca_error.NewBadRequestError, nil, "CUSTOM", nil},
		{"NilErr_EmptyCode_WithExtra", yca_error.NewBadRequestError, nil, "", map[string]any{"key": "value"}},
		{"NilErr_CustomCode_WithExtra", yca_error.NewBadRequestError, nil, "CUSTOM", map[string]any{"key": "value"}},
		{"WithErr_EmptyCode_NilExtra", yca_error.NewBadRequestError, errors.New("test"), "", nil},
		{"WithErr_CustomCode_NilExtra", yca_error.NewBadRequestError, errors.New("test"), "CUSTOM", nil},
		{"WithErr_EmptyCode_WithExtra", yca_error.NewBadRequestError, errors.New("test"), "", map[string]any{"key": "value"}},
		{"WithErr_CustomCode_WithExtra", yca_error.NewBadRequestError, errors.New("test"), "CUSTOM", map[string]any{"key": "value"}},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := tc.create(tc.err, tc.errorCode, tc.extra)
			s.NotNil(err)
			s.Equal(tc.err, err.Err)
			s.Equal(tc.extra, err.Extra)
			s.Equal(tc.errorCode, err.ErrorCode)
		})
	}
}

func (s *ErrorSuite) TestStatusCodes() {
	cases := []struct {
		name     string
		create   func(error, string, any) *yca_error.Error
		expected int
	}{
		{"BadRequest", yca_error.NewBadRequestError, http.StatusBadRequest},
		{"Unauthorized", yca_error.NewUnauthorizedError, http.StatusUnauthorized},
		{"PaymentRequired", yca_error.NewPaymentRequiredError, http.StatusPaymentRequired},
		{"Forbidden", yca_error.NewForbiddenError, http.StatusForbidden},
		{"NotFound", yca_error.NewNotFoundError, http.StatusNotFound},
		{"Conflict", yca_error.NewConflictError, http.StatusConflict},
		{"UnprocessableEntity", yca_error.NewUnprocessableEntityError, http.StatusUnprocessableEntity},
		{"TooManyRequests", yca_error.NewTooManyRequestsError, http.StatusTooManyRequests},
		{"InternalServerError", yca_error.NewInternalServerError, http.StatusInternalServerError},
		{"ServiceUnavailable", yca_error.NewServiceUnavailableError, http.StatusServiceUnavailable},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := tc.create(nil, "", nil)
			s.Equal(tc.expected, err.StatusCode)
			s.Empty(err.ErrorCode)
		})
	}
}

func (s *ErrorSuite) TestUnwrap() {
	underlying := errors.New("db failure")
	err := yca_error.NewInternalServerError(underlying, "DB_ERROR", nil)

	s.Equal(underlying, errors.Unwrap(err))
	s.True(errors.Is(err, underlying))
}

func (s *ErrorSuite) TestUnwrap_nilUnderlying() {
	err := yca_error.NewBadRequestError(nil, "", nil)
	s.Nil(errors.Unwrap(err))
}

func (s *ErrorSuite) TestAsError_direct() {
	apiErr := yca_error.NewNotFoundError(errors.New("missing"), "RESOURCE_NOT_FOUND", nil)

	e, ok := yca_error.AsError(apiErr)
	s.True(ok)
	s.Equal(apiErr, e)
	s.Equal(http.StatusNotFound, e.StatusCode)
	s.Equal("RESOURCE_NOT_FOUND", e.ErrorCode)
	s.Equal("missing", e.Error())
}

func (s *ErrorSuite) TestAsError_wrapped() {
	apiErr := yca_error.NewForbiddenError(errors.New("denied"), "FORBIDDEN", nil)
	wrapped := errors.Join(apiErr, errors.New("extra"))

	e, ok := yca_error.AsError(wrapped)
	s.True(ok)
	s.Equal(apiErr, e)
	s.Equal(http.StatusForbidden, e.StatusCode)
}

func (s *ErrorSuite) TestAsError_plainError() {
	e, ok := yca_error.AsError(errors.New("plain error"))
	s.False(ok)
	s.Nil(e)
}
