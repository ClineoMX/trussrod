package response

import (
	"encoding/json"
	"net/http"

	"github.com/clineomx/trussrod/apperr"
	"github.com/clineomx/trussrod/logging"
)

// WithError sends a structured JSON error response
// It automatically includes the trace ID from the request context if available
func WithError(w http.ResponseWriter, err error) {
	wrapped := apperr.Wrap(err)

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(wrapped.HTTPStatus)

	// Store error in response writer if available
	if writer, ok := w.(*logging.ResponseWriter); ok {
		writer.Error = err
	}

	// Encode error as JSON
	json.NewEncoder(w).Encode(wrapped)
}

// WithErrorLegacy is a legacy version that doesn't require a request
// It's kept for backward compatibility but should be migrated to WithError
func WithErrorLegacy(w http.ResponseWriter, err error) {
	wrapped := apperr.Wrap(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(wrapped.HTTPStatus)
	json.NewEncoder(w).Encode(wrapped)
}

func WithHeader(w http.ResponseWriter, key, value string) {
	w.Header().Set(key, value)
}

func WithStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func WithJSON(w http.ResponseWriter, status int, body any) {
	WithHeader(w, "Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func WithMessage(w http.ResponseWriter, status int, message string) {
	m := map[string]string{
		"message": message,
	}
	WithJSON(w, status, m)
}

type Paginated[T any] struct {
	Results []T `json:"results"`
	Count   int `json:"count"`
	Size    int `json:"size"`
}
