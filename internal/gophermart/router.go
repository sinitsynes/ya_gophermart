package gophermart

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	gophermartMiddleware "github.com/sinitsynes/ya_gophermart/internal/infra/httpserver/middleware"
)

const (
	requestTimeout time.Duration = 3 * time.Second
	maxRequestSize int64         = 20 * (1 << 10) //nolint: mnd // 20KB
)

func MainRouter(logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(gophermartMiddleware.Logger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(gophermartMiddleware.GzipCompressor)
	r.Use(gophermartMiddleware.GzipDecompressor)
	r.Use(gophermartMiddleware.LimitMaxRequestSize(maxRequestSize))

	return r
}
