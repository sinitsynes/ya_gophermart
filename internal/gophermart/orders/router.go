package orders

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/orders", h.CreateOrder)
	r.Get("/orders", h.GetOrdersByUser)
	return r
}
