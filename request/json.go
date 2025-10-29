package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Domedik/trussrod/errors"
	"github.com/go-playground/validator/v10"
)

func JSON[T any](r *http.Request) (T, error) {
	var zero T
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return zero, errors.BadRequest("invalid Content-Type header")
	}
	defer r.Body.Close()

	var v T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		return zero, err
	}

	if err := Validate(v); err != nil {
		return zero, err
	}

	return v, nil
}

func Validate[T any](i T) error {
	var validate = validator.New()
	err := validate.Struct(i)
	var v []string

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			msg := fmt.Sprintf("Field %s failed on the '%s' tag\n", err.Field(), err.Tag())
			v = append(v, msg)
		}
		return errors.ValidationFailed(strings.Join(v, ","))
	}

	return nil
}
