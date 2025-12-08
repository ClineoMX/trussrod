// Package middleware holds all middleware functions to run upon requests.
package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/Domedik/trussrod/errors"
	"github.com/Domedik/trussrod/logging"
	"github.com/Domedik/trussrod/request"
	"github.com/Domedik/trussrod/response"
)

// Middleware type alias for http.Handler.
type Middleware func(next http.Handler) http.Handler

// HasApiKey is a middleware to check if an authorized client has
// a secret api key and matches the one in settings.
func HasApiKey(key string) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			h, ok := request.GetHeader(r, request.ApiKeyHeader)
			if !ok || h != key {
				response.WithError(w, errors.Unauthorized())
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func WithTimeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// RecoveryMiddleware recovers from panics and logs them
func Recovery(logger *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					rid := request.MustGetTraceID(r)
					reqLogger := logger.WithTraceID(rid)

					// Log the panic
					fields := map[string]any{
						"panic":       err,
						"method":      r.Method,
						"path":        r.URL.Path,
						"remote_addr": r.RemoteAddr,
					}

					reqLogger.HTTP(
						r.Method,
						r.URL.Path,
						http.StatusInternalServerError,
						0,
						fields,
					)
					response.WithError(w, errors.Internal(nil))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func Logging(logger *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := request.MustGetTraceID(r)
			wrapped := logging.NewResponseWriter(w)

			reqLogger := logger.WithTraceID(rid)

			// Record start time
			start := time.Now()

			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			fields := map[string]any{
				"remote_addr":    r.RemoteAddr,
				"user_agent":     r.UserAgent(),
				"content_length": r.ContentLength,
			}

			if len(r.URL.Query()) > 0 {
				fields["query_params"] = r.URL.Query()
			}

			if user, ok := request.GetUser(r); ok {
				fields["user_id"] = user.ID
			}

			fields["response_size"] = wrapped.Body.Len()

			// Log request
			reqLogger.HTTP(
				r.Method,
				r.URL.Path,
				wrapped.StatusCode,
				duration,
				fields,
			)
		})
	}
}

func SetTraceID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid, ok := request.GetHeader(r, "X-Domedik-Trace-ID")
			if !ok {
				b := make([]byte, 16)
				rand.Read(b)
				rid = hex.EncodeToString(b)
			}
			next.ServeHTTP(w, request.WithTraceID(r, rid))
		})
	}
}
