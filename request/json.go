package request

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Domedik/trussrod/errors"
)

func JSON[T any](r *http.Request) (T, error) {
	var zero T
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return zero, errors.BadRequest()
	}
	defer r.Body.Close()

	var v T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		return zero, err
	}
	return v, nil
}
