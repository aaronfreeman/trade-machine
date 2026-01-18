package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"trade-machine/config"
	"trade-machine/internal/app"
	"trade-machine/models"
	"trade-machine/services"
	"trade-machine/templates"

	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP API requests
type Handler struct {
	app *app.App
	cfg *config.Config
}

// NewHandler creates a new Handler
func NewHandler(application *app.App, cfg *config.Config) *Handler {
	return &Handler{app: application, cfg: cfg}
}

// HandleIndex serves the main application page using templ
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Index().Render(r.Context(), w)
}

// HandleHealth returns the health status of the application
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "ok",
		"services": map[string]string{
			"database": "unknown",
		},
	}

	if h.app.Repo() != nil {
		ctx := r.Context()
		if err := h.app.Repo().Health(ctx); err == nil {
			status["services"].(map[string]string)["database"] = "connected"
		} else {
			status["services"].(map[string]string)["database"] = "disconnected"
			status["status"] = "degraded"
		}
	} else {
		status["services"].(map[string]string)["database"] = "not_configured"
	}

	// Add circuit breaker status
	cbStatus := services.GetGlobalRegistry().Status()
	status["circuit_breakers"] = cbStatus

	// Check if any breakers are open (degraded state)
	for _, cb := range cbStatus {
		if cb.State == "open" {
			status["status"] = "degraded"
			break
		}
	}

	h.jsonResponse(w, status)
}

// HandleGetPortfolio returns portfolio summary
func (h *Handler) HandleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	positions, err := h.app.GetPositions()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]interface{}{
		"positions": positions,
		"count":     len(positions),
	})
}

// HandleGetPositions returns all positions
func (h *Handler) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.app.GetPositions()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, positions)
}

// HandleGetRecommendations returns recommendations
func (h *Handler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	limit := h.ParseLimitParam(r, 50)

	recs, err := h.app.GetRecommendations(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, recs)
}

// HandleGetPendingRecommendations returns pending recommendations
func (h *Handler) HandleGetPendingRecommendations(w http.ResponseWriter, r *http.Request) {
	recs, err := h.app.GetPendingRecommendations()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, recs)
}

// HandleApproveRecommendation approves a recommendation
func (h *Handler) HandleApproveRecommendation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.jsonError(w, "Missing recommendation ID", http.StatusBadRequest)
		return
	}

	if err := h.app.ApproveRecommendation(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "approved", "id": id})
}

// HandleRejectRecommendation rejects a recommendation
func (h *Handler) HandleRejectRecommendation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.jsonError(w, "Missing recommendation ID", http.StatusBadRequest)
		return
	}

	if err := h.app.RejectRecommendation(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "rejected", "id": id})
}

// HandleAnalyzeStock triggers analysis of a stock
func (h *Handler) HandleAnalyzeStock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string `json:"symbol"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Try form value
		req.Symbol = r.FormValue("symbol")
	}

	if req.Symbol == "" {
		h.jsonError(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	// Normalize symbol to uppercase
	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))

	if err := h.ValidateSymbol(req.Symbol); err != nil {
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	rec, err := h.app.AnalyzeStock(req.Symbol)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, rec)
}

// HandleGetTrades returns recent trades
func (h *Handler) HandleGetTrades(w http.ResponseWriter, r *http.Request) {
	limit := h.ParseLimitParam(r, 50)

	trades, err := h.app.GetTrades(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, trades)
}

// HandleGetAgentRuns returns recent agent runs
func (h *Handler) HandleGetAgentRuns(w http.ResponseWriter, r *http.Request) {
	limit := h.ParseLimitParam(r, 50)

	runs, err := h.app.GetAgentRuns(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, runs)
}

// Helper functions

// ValidateSymbol validates a stock symbol
func (h *Handler) ValidateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	if len(symbol) > 10 {
		return fmt.Errorf("symbol too long (max 10 characters)")
	}

	matched, _ := regexp.MatchString("^[A-Z0-9.-]+$", symbol)
	if !matched {
		return fmt.Errorf("invalid symbol format (alphanumeric, dots, and dashes only)")
	}

	return nil
}

// ParseLimitParam parses the limit query parameter
func (h *Handler) ParseLimitParam(r *http.Request, defaultLimit int) int {
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			return l
		}
	}
	return defaultLimit
}

func (h *Handler) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// StatusResponse represents a status response
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// AnalyzeRequest represents a stock analysis request
type AnalyzeRequest struct {
	Symbol string `json:"symbol"`
}

// Ensure models are exported for JSON serialization
var _ = models.Position{}
var _ = models.Trade{}
var _ = models.Recommendation{}
var _ = models.AgentRun{}
