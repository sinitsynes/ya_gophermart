package balance

import (
	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/handler"
)

func RegisterRoutes(r chi.Router, h *handler.Handler) {
	r.Get("/balance", h.GetBalanceByUser)
	r.Post("/balance/withdraw", h.CreateWithdrawal)
	r.Get("/withdrawals", h.GetWithdrawalsByUser)
}
