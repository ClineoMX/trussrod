package response

import (
	"encoding/json"
	"net/http"

	"github.com/Domedik/trussrod/errors"
	"github.com/Domedik/trussrod/logging"
)

func WithError(w http.ResponseWriter, err error) {
	wrapped := errors.Wrap(err)
	if writer, ok := w.(*logging.ResponseWriter); ok {
		writer.Error = err
		http.Error(writer, wrapped.Error(), wrapped.HTTPStatus)
	} else {
		http.Error(w, wrapped.Error(), wrapped.HTTPStatus)
	}
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
	Page    int `json:"page"`
	Size    int `json:"size"`
}
