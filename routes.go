package main

import (
	"net/http"
	"time"

	"trade-machine/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures a Chi router with all routes
func NewRouter(h *APIHandler, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(cfg.Agent.TimeoutSeconds) * time.Second))
	r.Use(corsMiddleware(cfg.HTTP.CORSAllowedOrigins))

	// Root routes
	r.Get("/", h.handleIndex)
	r.Get("/index.html", h.handleIndex)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", h.handleHealth)

		// Portfolio
		r.Get("/portfolio", h.handleGetPortfolio)
		r.Get("/positions", h.handleGetPositions)

		// Recommendations
		r.Route("/recommendations", func(r chi.Router) {
			r.Get("/", h.handleGetRecommendations)
			r.Get("/pending", h.handleGetPendingRecommendations)
			r.Post("/{id}/approve", h.handleApproveRecommendation)
			r.Post("/{id}/reject", h.handleRejectRecommendation)
		})

		// Analysis
		r.Post("/analyze", h.handleAnalyzeStock)

		// Trades
		r.Get("/trades", h.handleGetTrades)

		// Agent runs
		r.Get("/agents/runs", h.handleGetAgentRuns)
	})

	return r
}

// corsMiddleware returns CORS middleware with the specified allowed origins
func corsMiddleware(allowedOrigins string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
