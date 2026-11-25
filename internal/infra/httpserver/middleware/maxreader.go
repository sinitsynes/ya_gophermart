package middleware

import (
	"errors"
	"io"
	"net/http"
)

func LimitMaxRequestSize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, maxSize)
				_, err := io.ReadAll(r.Body)
				if err != nil {
					if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
						w.WriteHeader(http.StatusRequestEntityTooLarge)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
			},
		)
	}
}
