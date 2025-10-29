package response

import (
	"encoding/json"
	"net/http"

	"github.com/Domedik/trussrod/errors"
)

func WriteError(w http.ResponseWriter, err error) {
	wrapped := errors.Wrap(err)
	http.Error(w, wrapped.Error(), wrapped.HTTPStatus)
}

func WriteHeader(w http.ResponseWriter, key, value string) {
	w.Header().Set(key, value)
}

func WriteStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func JSON(w http.ResponseWriter, status int, body any) {
	WriteHeader(w, "Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

type Paginated[T any] struct {
	Results []T `json:"results"`
	Count   int `json:"count"`
	Page    int `json:"page"`
	Size    int `json:"size"`
}
