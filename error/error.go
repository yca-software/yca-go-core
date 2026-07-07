package chi_error

import (
	"errors"
	"net/http"
)

type Error struct {
	StatusCode int    `json:"-"`
	Err        error  `json:"-"` // Internal error - not exposed in JSON responses
	ErrorCode  string `json:"errorCode"`
	Message    string `json:"message,omitempty"`
	Extra      any    `json:"extra,omitempty"`
}

// ErrorResponse is the API error body shape used for Swagger/OpenAPI documentation.
type ErrorResponse struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message,omitempty"`
	Extra     any    `json:"extra,omitempty"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.ErrorCode
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// AsError returns *Error if err is or wraps this type; otherwise (nil, false).
// Handlers can use it to respond with StatusCode and JSON body.
func AsError(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}

func newError(statusCode int, err error, errorCode string, extra any) *Error {
	return &Error{
		StatusCode: statusCode,
		Err:        err,
		ErrorCode:  errorCode,
		Message:    errorMessage(statusCode, err, errorCode),
		Extra:      extra,
	}
}

func errorMessage(statusCode int, err error, errorCode string) string {
	if statusCode >= http.StatusInternalServerError {
		return ""
	}
	if errorCode != "" {
		return ""
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func NewInternalServerError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusInternalServerError, err, errorCode, extra)
}

func NewUnprocessableEntityError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusUnprocessableEntity, err, errorCode, extra)
}

func NewConflictError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusConflict, err, errorCode, extra)
}

func NewNotFoundError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusNotFound, err, errorCode, extra)
}

func NewUnauthorizedError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusUnauthorized, err, errorCode, extra)
}

func NewBadRequestError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusBadRequest, err, errorCode, extra)
}

func NewForbiddenError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusForbidden, err, errorCode, extra)
}

func NewPaymentRequiredError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusPaymentRequired, err, errorCode, extra)
}

func NewTooManyRequestsError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusTooManyRequests, err, errorCode, extra)
}

func NewServiceUnavailableError(err error, errorCode string, extra any) *Error {
	return newError(http.StatusServiceUnavailable, err, errorCode, extra)
}
