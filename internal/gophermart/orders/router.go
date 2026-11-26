package orders

import (
	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/orders/handler"
)

func RegisterRoutes(r chi.Router, h *handler.Handler) {
	r.Post("/orders", h.CreateOrder)
	r.Get("/orders", h.GetOrdersByUser)
}
