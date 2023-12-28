package server

import (
	"github.com/JuliyaMS/gophermart/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	h *Handlers
	r *chi.Mux
}

func NewRouter(h *Handlers) *Router {
	router := chi.NewRouter()

	router.Post("/api/user/register", middleware.CompressionGzip(h.registration))
	router.Post("/api/user/login", middleware.CompressionGzip(h.login))
	router.Post("/api/user/orders", middleware.CompressionGzip(h.loadOrders))
	router.Post("/api/user/balance/withdraw", middleware.CompressionGzip(h.balanceWithdraw))
	router.Get("/api/user/orders", middleware.CompressionGzip(h.getOrders))
	router.Get("/api/user/balance", middleware.CompressionGzip(h.getBalance))
	router.Get("/api/user/withdrawals", middleware.CompressionGzip(h.infoWithdraw))

	return &Router{
		h: h,
		r: router,
	}
}

func (ro *Router) GetRouter() *chi.Mux {
	return ro.r
}
