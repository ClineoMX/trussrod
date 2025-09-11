package errors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeValidation   = "VALIDATION_ERROR"
	ErrorTypeNotFound     = "NOT_FOUND"
	ErrorTypeUnauthorized = "UNAUTHORIZED"
	ErrorTypeForbidden    = "FORBIDDEN"
	ErrorTypeConflict     = "CONFLICT"
	ErrorTypeInternal     = "INTERNAL_ERROR"
	ErrorTypeBadRequest   = "BAD_REQUEST"
)

// Common error messages
const (
	ErrMsgInvalidPayload   = "The request payload is invalid"
	ErrMsgBadRequest       = "The request is not complete"
	ErrMsgResourceNotFound = "The resource does not exist"
	ErrMsgUnauthorized     = "Unauthorized access"
	ErrMsgForbidden        = "Access forbidden"

	ErrMsgConflictOnCreation = "There is a conflict on the creation of the resource"
	ErrMsgInternalError      = "Something went wrong, please try again later"
)

// AppError represents a structured application error
type AppError struct {
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	HTTPStatus  int       `json:"-"`
	OriginalErr error     `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%d: %s (original: %v)", e.HTTPStatus, e.Message, e.OriginalErr)
	}
	if e.Details != "" {
		return fmt.Sprintf("%d: %s (details: %s)", e.HTTPStatus, e.Message, e.Details)
	}
	return fmt.Sprintf("%d: %s", e.HTTPStatus, e.Message)
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.OriginalErr
}

func ValidationFailed(details string) *AppError {
	err := &AppError{
		Type:       ErrorTypeValidation,
		Message:    ErrMsgInvalidPayload,
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
	return err
}

func NotFound() *AppError {
	err := &AppError{
		Type:       ErrorTypeNotFound,
		Message:    ErrMsgResourceNotFound,
		HTTPStatus: http.StatusNotFound,
	}
	return err
}

func Unauthorized() *AppError {
	err := &AppError{
		Type:       ErrorTypeUnauthorized,
		Message:    ErrMsgUnauthorized,
		HTTPStatus: http.StatusUnauthorized,
	}
	return err
}

func Forbidden() *AppError {
	err := &AppError{
		Type:       ErrorTypeForbidden,
		Message:    ErrMsgForbidden,
		HTTPStatus: http.StatusForbidden,
	}
	return err
}

func Conflict() *AppError {
	err := &AppError{
		Type:       ErrorTypeConflict,
		Message:    ErrMsgConflictOnCreation,
		HTTPStatus: http.StatusConflict,
	}
	return err
}

func BadRequest() *AppError {
	err := &AppError{
		Type:       ErrorTypeBadRequest,
		Message:    ErrMsgBadRequest,
		HTTPStatus: http.StatusBadRequest,
	}
	return err
}

func Internal(original error) *AppError {
	err := &AppError{
		Type:        ErrorTypeInternal,
		Message:     ErrMsgInternalError,
		HTTPStatus:  http.StatusInternalServerError,
		OriginalErr: original,
	}
	return err
}

// Wrap wraps an existing error with additional context
func Wrap(err error) *AppError {
	if err == nil {
		return nil
	}

	var jsonErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		return BadRequest()
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return Conflict()
		case "23502":
			return BadRequest()
		default:
			return Internal(pqErr)
		}
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	if errors.Is(err, sql.ErrNoRows) {
		return NotFound()
	}

	return Internal(err)
}
