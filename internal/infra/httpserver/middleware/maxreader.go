package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

func LimitMaxRequestSize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, maxSize)
				body, err := io.ReadAll(r.Body)
				if err != nil {
					if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
						w.WriteHeader(http.StatusRequestEntityTooLarge)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_ = r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(body))
				next.ServeHTTP(w, r)
			},
		)
	}
}
