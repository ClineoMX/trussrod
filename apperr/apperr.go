package apperr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const ErrMsgInternalError = "Something went wrong, please try again later"

// AppError represents a structured application error
type AppError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	HTTPStatus  int    `json:"-"`
	OriginalErr error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	var parts []string
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Code))
	}
	parts = append(parts, e.Message)
	return strings.Join(parts, " ")
}

func (e *AppError) MarshalJSON() ([]byte, error) {
	type Alias AppError
	aux := &struct {
		*Alias
		HTTPStatus int    `json:"status"`
		Timestamp  string `json:"timestamp"`
	}{
		Alias:      (*Alias)(e),
		HTTPStatus: e.HTTPStatus,
	}
	return json.Marshal(aux)
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.OriginalErr
}

func NotFound(message string) *AppError {
	return &AppError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    message,
		HTTPStatus: http.StatusNotFound,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Code:       "RESOURCE_CONFLICT",
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func Internal(original error) *AppError {
	return &AppError{
		Code:        "INTERNAL_ERROR",
		Message:     ErrMsgInternalError,
		HTTPStatus:  http.StatusInternalServerError,
		OriginalErr: original,
	}
}

// Wrap wraps an existing error with additional context
// It attempts to map common error types to appropriate AppError instances
func Wrap(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return Internal(err)
}
