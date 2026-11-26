package users

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/register", h.RegisterUser)
	return r
}
