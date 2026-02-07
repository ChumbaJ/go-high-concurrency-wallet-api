// Package app
package app

import (
	"net/http"

	"test-psql/internal/http/middleware"
)

type handler interface {
	UpdateWalletBalance(w http.ResponseWriter, r *http.Request)
	GetWalletBalance(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	Handler  handler
	Limiter  *middleware.Limiter
}

func NewServer(handler handler, limiter *middleware.Limiter) *Server {
	return &Server{
		Handler: handler,
		Limiter: limiter,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// POST api/v1/wallet
	mux.HandleFunc("POST /api/v1/wallet", s.Handler.UpdateWalletBalance)
	// GET api/v1/wallets/{WALLET_UUID}
	mux.HandleFunc("GET /api/v1/wallets/", s.Handler.GetWalletBalance)

	h := middleware.RecoverMiddleware(mux)
	if s.Limiter != nil {
		h = s.Limiter.Middleware(h)
	}
	return h
}
