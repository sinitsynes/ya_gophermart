package balance

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sinitsynes/ya_gophermart/internal/gophermart/balance/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/balance", h.GetBalanceByUser)
	r.Post("/balance/withdraw", h.CreateWithdrawal)
	r.Get("/withdrawals", h.GetWithdrawalsByUser)
	return r
}
