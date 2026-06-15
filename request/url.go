package request

import (
	"encoding/json"
	"net/http"
)

type PathParameter string

const (
	PatientID  PathParameter = "patient_id"
	NoteID     PathParameter = "note_id"
	ConsentID  PathParameter = "consent_id"
	ResourceID PathParameter = "resource_id"
)

func GetPathValue(r *http.Request, key PathParameter) (string, bool) {
	value := r.PathValue(string(key))
	return value, value != ""
}

func MustGetPathValue(r *http.Request, key PathParameter) string {
	value, ok := GetPathValue(r, key)
	if !ok {
		panic("could not retrieve parameter from url")
	}
	return value
}

func GetQueryParamsAs[T any](r *http.Request) (T, error) {
	query := r.URL.Query()
	var t T
	filters := make(map[string]any)

	for key, values := range query {
		filters[key] = values[0]
	}

	m, err := json.Marshal(filters)
	if err != nil {
		return t, err
	}

	if err := json.Unmarshal(m, &t); err != nil {
		return t, err
	}

	return t, nil
}
