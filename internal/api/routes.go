package api

import (
	"net/http"
	"time"

	"trade-machine/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures a Chi router with all routes
func NewRouter(h *Handler, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(cfg.Agent.TimeoutSeconds) * time.Second))
	r.Use(CORSMiddleware(cfg.HTTP.CORSAllowedOrigins))

	// Root routes
	r.Get("/", h.HandleIndex)
	r.Get("/index.html", h.HandleIndex)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", h.HandleHealth)

		// Portfolio
		r.Get("/portfolio", h.HandleGetPortfolio)
		r.Get("/positions", h.HandleGetPositions)

		// Recommendations
		r.Route("/recommendations", func(r chi.Router) {
			r.Get("/", h.HandleGetRecommendations)
			r.Get("/pending", h.HandleGetPendingRecommendations)
			r.Post("/{id}/approve", h.HandleApproveRecommendation)
			r.Post("/{id}/reject", h.HandleRejectRecommendation)
		})

		// Analysis
		r.Post("/analyze", h.HandleAnalyzeStock)

		// Trades
		r.Get("/trades", h.HandleGetTrades)

		// Agent runs
		r.Get("/agents/runs", h.HandleGetAgentRuns)
	})

	return r
}

// CORSMiddleware returns CORS middleware with the specified allowed origins
func CORSMiddleware(allowedOrigins string) func(next http.Handler) http.Handler {
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
