package auth

import (
	"net/http"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/auth/handler"
)

func NewRouter(h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/register", h.RegisterUser)
	return mux
}
