package request

import (
	"net/http"
)

type PathParameter string

const (
	PatientID PathParameter = "patient_id"
	NoteID    PathParameter = "note_id"
	ConsentID PathParameter = "consent_id"
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
