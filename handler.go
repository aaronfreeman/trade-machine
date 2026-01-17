package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/templates"

	"github.com/go-chi/chi/v5"
)

// APIHandler handles HTTP API requests
type APIHandler struct {
	app *App
	cfg *config.Config
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(app *App, cfg *config.Config) *APIHandler {
	return &APIHandler{app: app, cfg: cfg}
}

// handleIndex serves the main application page using templ
func (h *APIHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Index().Render(r.Context(), w)
}

// handleHealth returns the health status of the application
func (h *APIHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "ok",
		"services": map[string]string{
			"database": "unknown",
		},
	}

	if h.app.repo != nil {
		ctx := r.Context()
		if err := h.app.repo.Health(ctx); err == nil {
			status["services"].(map[string]string)["database"] = "connected"
		} else {
			status["services"].(map[string]string)["database"] = "disconnected"
			status["status"] = "degraded"
		}
	} else {
		status["services"].(map[string]string)["database"] = "not_configured"
	}

	h.jsonResponse(w, status)
}

// handleGetPortfolio returns portfolio summary
func (h *APIHandler) handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
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

// handleGetPositions returns all positions
func (h *APIHandler) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.app.GetPositions()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, positions)
}

// handleGetRecommendations returns recommendations
func (h *APIHandler) handleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	limit := h.parseLimitParam(r, 50)

	recs, err := h.app.GetRecommendations(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, recs)
}

// handleGetPendingRecommendations returns pending recommendations
func (h *APIHandler) handleGetPendingRecommendations(w http.ResponseWriter, r *http.Request) {
	recs, err := h.app.GetPendingRecommendations()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, recs)
}

// handleApproveRecommendation approves a recommendation
func (h *APIHandler) handleApproveRecommendation(w http.ResponseWriter, r *http.Request) {
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

// handleRejectRecommendation rejects a recommendation
func (h *APIHandler) handleRejectRecommendation(w http.ResponseWriter, r *http.Request) {
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

// handleAnalyzeStock triggers analysis of a stock
func (h *APIHandler) handleAnalyzeStock(w http.ResponseWriter, r *http.Request) {
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

	if err := h.validateSymbol(req.Symbol); err != nil {
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

// handleGetTrades returns recent trades
func (h *APIHandler) handleGetTrades(w http.ResponseWriter, r *http.Request) {
	limit := h.parseLimitParam(r, 50)

	trades, err := h.app.GetTrades(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, trades)
}

// handleGetAgentRuns returns recent agent runs
func (h *APIHandler) handleGetAgentRuns(w http.ResponseWriter, r *http.Request) {
	limit := h.parseLimitParam(r, 50)

	runs, err := h.app.GetAgentRuns(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, runs)
}

// Helper functions

func (h *APIHandler) validateSymbol(symbol string) error {
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

func (h *APIHandler) parseLimitParam(r *http.Request, defaultLimit int) int {
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			return l
		}
	}
	return defaultLimit
}

func (h *APIHandler) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *APIHandler) jsonError(w http.ResponseWriter, message string, status int) {
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
