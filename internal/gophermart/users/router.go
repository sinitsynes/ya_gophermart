package users

import (
	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/users/handler"
)

func RegisterRoutes(r chi.Router, h *handler.Handler) {
	r.Post("/login", h.Login)
	r.Post("/register", h.RegisterUser)
}
