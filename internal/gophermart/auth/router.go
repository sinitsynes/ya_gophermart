package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	return r
}
