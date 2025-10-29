package request

import (
	"net/http"
)

type Header string

const (
	ApiKeyHeader   Header = "X-Domedik-Api-Key"
	AuthHeader     Header = "Authorization"
	IdentityHeader Header = "X-Domedik-Identity"
)

func GetHeader(r *http.Request, h Header) (string, bool) {
	v := r.Header.Get(string(h))
	if v == "" {
		return "", false
	}
	return v, true
}
