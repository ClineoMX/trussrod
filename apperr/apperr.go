package apperr

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aws/smithy-go"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
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
	Code        string          `json:"code"`
	Message     string          `json:"message"`
	Details     string          `json:"details,omitempty"`
	Fields      map[string]any  `json:"fields,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
	TraceID     string          `json:"trace_id,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
	HTTPStatus  int             `json:"-"`
	OriginalErr error           `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	var parts []string
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Code))
	}
	parts = append(parts, e.Message)
	if e.Details != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Details))
	}
	return strings.Join(parts, " ")
}

// MarshalJSON implements custom JSON marshaling for AppError
// This ensures we only expose safe information in API responses
func (e *AppError) MarshalJSON() ([]byte, error) {
	type Alias AppError
	aux := &struct {
		*Alias
		HTTPStatus int    `json:"status"`
		Timestamp  string `json:"timestamp"`
	}{
		Alias:      (*Alias)(e),
		HTTPStatus: e.HTTPStatus,
		Timestamp:  e.Timestamp.Format(time.RFC3339),
	}
	return json.Marshal(aux)
}

// LogError returns string for logging with additional context
func (e *AppError) LogError() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s] %d: %s", e.Code, e.HTTPStatus, e.Message))

	if e.Details != "" {
		parts = append(parts, fmt.Sprintf("details: %s", e.Details))
	}

	if e.OriginalErr != nil {
		parts = append(parts, fmt.Sprintf("original: %v", e.OriginalErr))
	}

	return strings.Join(parts, " | ")
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.OriginalErr
}

// ValidationFailed creates a validation error with details
// details should contain field-specific validation errors
func ValidationFailed(details string) *AppError {
	return &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    ErrMsgInvalidPayload,
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
		Timestamp:  time.Now().UTC(),
	}
}

// ValidationFailedWithFields creates a validation error with structured field errors
func ValidationFailedWithFields(fieldErrors map[string]string) *AppError {
	fields := make(map[string]any)
	for k, v := range fieldErrors {
		fields[k] = v
	}

	var details []string
	for field, err := range fieldErrors {
		details = append(details, fmt.Sprintf("%s: %s", field, err))
	}

	return &AppError{
		Code:       "VALIDATION_FAILED",
		Message:    ErrMsgInvalidPayload,
		HTTPStatus: http.StatusBadRequest,
		Details:    strings.Join(details, "; "),
		Fields:     fields,
		Timestamp:  time.Now().UTC(),
	}
}

// NotFound creates a not found error
func NotFound() *AppError {
	return &AppError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    ErrMsgResourceNotFound,
		HTTPStatus: http.StatusNotFound,
		Timestamp:  time.Now().UTC(),
	}
}

// NotFoundWithResource creates a not found error with resource context
func NotFoundWithResource(resourceType, resourceID string) *AppError {
	return &AppError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    fmt.Sprintf("The %s with identifier '%s' does not exist", resourceType, resourceID),
		HTTPStatus: http.StatusNotFound,
		Fields: map[string]any{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		},
		Timestamp: time.Now().UTC(),
	}
}

// Unauthorized creates an unauthorized error
func Unauthorized() *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    ErrMsgUnauthorized,
		HTTPStatus: http.StatusUnauthorized,
		Timestamp:  time.Now().UTC(),
	}
}

// UnauthorizedWithReason creates an unauthorized error with a specific reason
func UnauthorizedWithReason(reason string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    ErrMsgUnauthorized,
		Details:    reason,
		HTTPStatus: http.StatusUnauthorized,
		Timestamp:  time.Now().UTC(),
	}
}

// Forbidden creates a forbidden error
func Forbidden() *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    ErrMsgForbidden,
		HTTPStatus: http.StatusForbidden,
		Timestamp:  time.Now().UTC(),
	}
}

// ForbiddenWithReason creates a forbidden error with a specific reason
func ForbiddenWithReason(reason string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    ErrMsgForbidden,
		Details:    reason,
		HTTPStatus: http.StatusForbidden,
		Timestamp:  time.Now().UTC(),
	}
}

// Conflict creates a conflict error
func Conflict() *AppError {
	return &AppError{
		Code:       "RESOURCE_CONFLICT",
		Message:    ErrMsgConflictOnCreation,
		HTTPStatus: http.StatusConflict,
		Timestamp:  time.Now().UTC(),
	}
}

// ConflictWithDetails creates a conflict error with specific details about what conflicted
func ConflictWithDetails(resourceType, conflictingField string) *AppError {
	return &AppError{
		Code:       "RESOURCE_CONFLICT",
		Message:    fmt.Sprintf("The %s already exists with the provided %s", resourceType, conflictingField),
		HTTPStatus: http.StatusConflict,
		Fields: map[string]any{
			"resource_type":     resourceType,
			"conflicting_field": conflictingField,
		},
		Timestamp: time.Now().UTC(),
	}
}

func ConflictWithFields(msg string, fields any) *AppError {
	data, err := json.Marshal(fields)
	if err != nil {
		return Internal(err)
	}

	return &AppError{
		Code:       "RESOURCE_CONFLICT",
		Message:    msg,
		HTTPStatus: http.StatusConflict,
		Data:       json.RawMessage(data),
		Timestamp:  time.Now().UTC(),
	}
}

// BadRequest creates a bad request error with details
// Accepts string or fmt.Stringer for details
func BadRequest(detail any) *AppError {
	var dets string
	switch v := detail.(type) {
	case string:
		dets = v
	case fmt.Stringer:
		dets = v.String()
	case nil:
		dets = ""
	default:
		dets = fmt.Sprintf("%v", v)
	}

	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    ErrMsgBadRequest,
		HTTPStatus: http.StatusBadRequest,
		Details:    dets,
		Timestamp:  time.Now().UTC(),
	}
}

// BadRequestWithCode creates a bad request error with a specific error code
func BadRequestWithCode(code, message string, details any) *AppError {
	var dets string
	switch v := details.(type) {
	case string:
		dets = v
	case fmt.Stringer:
		dets = v.String()
	case nil:
		dets = ""
	default:
		dets = fmt.Sprintf("%v", v)
	}

	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Details:    dets,
		Timestamp:  time.Now().UTC(),
	}
}

// Internal creates an internal server error
// The details field is left empty to avoid exposing internal error messages
func Internal(original error) *AppError {
	return &AppError{
		Code:        "INTERNAL_ERROR",
		Message:     ErrMsgInternalError,
		HTTPStatus:  http.StatusInternalServerError,
		OriginalErr: original,
		Timestamp:   time.Now().UTC(),
		// Details intentionally left empty for security
	}
}

// Wrap wraps an existing error with additional context
// It attempts to map common error types to appropriate AppError instances
func Wrap(err error) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Handle context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return BadRequestWithCode(
			"REQUEST_TIMEOUT",
			"The request exceeded the maximum allowed time",
			"Operation timed out",
		)
	}
	if errors.Is(err, context.Canceled) {
		return BadRequestWithCode(
			"REQUEST_CANCELLED",
			"The request was cancelled",
			"Operation was cancelled",
		)
	}

	// Handle AWS/Smithy API errors
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		message := apiErr.ErrorMessage()
		switch code {
		case "NotAuthorizedException":
			return UnauthorizedWithReason(message)
		case "UserNotFoundException":
			return NotFound()
		case "UserNotConfirmedException":
			return ForbiddenWithReason("User account is not confirmed")
		case "InvalidParameterException":
			return BadRequestWithCode("INVALID_PARAMETER", message, nil)
		case "ResourceNotFoundException":
			return NotFound()
		case "InvalidPasswordException":
			return BadRequestWithCode("INVALID_PASSWORD", "The provided password does not meet requirements", message)
		case "PasswordResetRequiredException":
			return ForbiddenWithReason("Password reset is required")
		case "TooManyRequestsException":
			return BadRequestWithCode("RATE_LIMIT_EXCEEDED", "Too many requests", "Please try again later")
		case "LimitExceededException":
			return BadRequestWithCode("LIMIT_EXCEEDED", "Service limit exceeded", message)
		case "AccessDeniedException":
			return ForbiddenWithReason(message)
		case "InvalidUserPoolConfigurationException":
			return Internal(apiErr)
		case "NoSuchKey", "NoSuchBucket":
			return NotFound()
		case "AccessDenied", "Forbidden":
			return ForbiddenWithReason(message)
		case "InvalidKeyId", "InvalidCiphertextException", "IncorrectKeyException":
			return BadRequestWithCode("INVALID_ENCRYPTION_KEY", "Invalid encryption key", message)
		case "KMSInvalidStateException":
			return BadRequestWithCode("KMS_KEY_UNAVAILABLE", "Encryption key is unavailable", message)
		default:
			return BadRequestWithCode("AWS_ERROR", fmt.Sprintf("AWS service error: %s", code), message)
		}
	}

	// Handle JSON syntax errors
	var jsonErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		return BadRequestWithCode(
			"INVALID_JSON",
			"The request body contains invalid JSON",
			fmt.Sprintf("JSON syntax error at offset %d", jsonErr.Offset),
		)
	}

	// Handle JSON unmarshal type errors
	var jsonTypeErr *json.UnmarshalTypeError
	if errors.As(err, &jsonTypeErr) {
		return BadRequestWithCode(
			"INVALID_JSON_TYPE",
			"The request body contains a value with an incorrect type",
			fmt.Sprintf("Field '%s' expects type %s but got %s", jsonTypeErr.Field, jsonTypeErr.Type, jsonTypeErr.Value),
		)
	}

	// Handle PostgreSQL errors (lib/pq)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505": // Unique constraint violation
			constraint := pqErr.Constraint
			if constraint != "" {
				return ConflictWithDetails("resource", constraint)
			}
			return Conflict()
		case "23502": // Not null constraint violation
			column := pqErr.Column
			if column != "" {
				return BadRequestWithCode(
					"REQUIRED_FIELD_MISSING",
					"Required field is missing",
					fmt.Sprintf("Field '%s' is required but was not provided", column),
				)
			}
			return BadRequest("One or more required fields are missing")
		case "23503": // Foreign key violation
			return BadRequestWithCode(
				"INVALID_REFERENCE",
				"The request references a resource that does not exist",
				pqErr.Message,
			)
		case "23514": // Check constraint violation
			return BadRequestWithCode(
				"CONSTRAINT_VIOLATION",
				"The provided data violates a validation constraint",
				pqErr.Message,
			)
		case "08003": // Connection does not exist
			return BadRequestWithCode(
				"DATABASE_CONNECTION_ERROR",
				"Database connection error",
				"Unable to establish database connection",
			)
		case "08006": // Connection failure
			return BadRequestWithCode(
				"DATABASE_CONNECTION_FAILED",
				"Database connection failed",
				"Lost connection to database",
			)
		case "42P01": // Undefined table
			return Internal(pqErr)
		case "42P07": // Duplicate table
			return Conflict()
		default:
			return Internal(pqErr)
		}
	}

	// Handle pgx errors (pgx/v5) - pgx uses pgconn.PgError
	// Check for pgx connection errors and constraint violations via error message
	errStr := err.Error()
	if strings.Contains(errStr, "duplicate key value") || strings.Contains(errStr, "unique constraint") {
		return Conflict()
	}
	if strings.Contains(errStr, "violates not-null constraint") || strings.Contains(errStr, "null value in column") {
		return BadRequest("One or more required fields are missing")
	}
	if strings.Contains(errStr, "violates foreign key constraint") {
		return BadRequestWithCode(
			"INVALID_REFERENCE",
			"The request references a resource that does not exist",
			errStr,
		)
	}
	if strings.Contains(errStr, "violates check constraint") {
		return BadRequestWithCode(
			"CONSTRAINT_VIOLATION",
			"The provided data violates a validation constraint",
			errStr,
		)
	}

	// Handle SQL no rows error
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return NotFound()
	}

	// Handle Redis errors
	if errors.Is(err, redis.Nil) {
		return NotFound()
	}
	var redisErr redis.Error
	if errors.As(err, &redisErr) {
		msg := redisErr.Error()
		if strings.Contains(msg, "connection refused") || strings.Contains(msg, "timeout") {
			return BadRequestWithCode(
				"CACHE_UNAVAILABLE",
				"Cache service is temporarily unavailable",
				"Please try again later",
			)
		}
		return Internal(redisErr)
	}

	// Handle JWT errors (jwt/v5)
	// jwt/v5 uses different error types - check error messages and use errors.Is
	if errors.Is(err, jwt.ErrTokenMalformed) {
		return UnauthorizedWithReason("Token is malformed")
	}
	if errors.Is(err, jwt.ErrTokenExpired) {
		return UnauthorizedWithReason("Token has expired")
	}
	if errors.Is(err, jwt.ErrTokenNotValidYet) {
		return UnauthorizedWithReason("Token is not valid yet")
	}
	if errors.Is(err, jwt.ErrTokenInvalidAudience) {
		return UnauthorizedWithReason("Token audience is invalid")
	}
	if errors.Is(err, jwt.ErrTokenInvalidIssuer) {
		return UnauthorizedWithReason("Token issuer is invalid")
	}
	if errors.Is(err, jwt.ErrTokenInvalidId) {
		return UnauthorizedWithReason("Token ID is invalid")
	}
	if errors.Is(err, jwt.ErrTokenInvalidClaims) {
		return UnauthorizedWithReason("Token claims are invalid")
	}
	// Check for JWT-related error messages (signature errors, etc.)
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "token") {
		if strings.Contains(errMsg, "signature") || strings.Contains(errMsg, "invalid") {
			return UnauthorizedWithReason("Token validation failed")
		}
	}

	// Handle network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return BadRequestWithCode(
				"NETWORK_TIMEOUT",
				"Network operation timed out",
				"The connection to the service timed out",
			)
		}
		if netOpErr, ok := netErr.(*net.OpError); ok {
			if netOpErr.Op == "dial" {
				return BadRequestWithCode(
					"CONNECTION_REFUSED",
					"Unable to connect to service",
					"Service is temporarily unavailable",
				)
			}
		}
		return Internal(netErr)
	}

	// Handle DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return BadRequestWithCode(
			"DNS_ERROR",
			"DNS resolution failed",
			fmt.Sprintf("Unable to resolve hostname: %s", dnsErr.Name),
		)
	}

	// Handle IO errors
	if errors.Is(err, io.EOF) {
		return BadRequestWithCode(
			"UNEXPECTED_EOF",
			"Unexpected end of input",
			"The request body was incomplete",
		)
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return BadRequestWithCode(
			"INCOMPLETE_DATA",
			"Incomplete data received",
			"The request body was truncated",
		)
	}

	// Handle base64 encoding errors
	var base64Err base64.CorruptInputError
	if errors.As(err, &base64Err) {
		return BadRequestWithCode(
			"INVALID_BASE64",
			"Invalid base64 encoding",
			"The provided data is not valid base64",
		)
	}

	// Default to internal error
	return Internal(err)
}

// WithTraceID adds a trace ID to the error for request correlation
func (e *AppError) WithTraceID(traceID string) *AppError {
	e.TraceID = traceID
	return e
}

// WithField adds a field to the error's Fields map
func (e *AppError) WithField(key string, value any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields[key] = value
	return e
}

// WithFields adds multiple fields to the error's Fields map
func (e *AppError) WithFields(fields map[string]any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}
