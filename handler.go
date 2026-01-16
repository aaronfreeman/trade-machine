package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"trade-machine/models"
)

// APIHandler handles HTTP API requests
type APIHandler struct {
	app *App
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(app *App) *APIHandler {
	return &APIHandler{app: app}
}

// ServeHTTP routes requests to appropriate handlers
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path

	switch {
	// Legacy greet endpoint
	case path == "/api/greet":
		h.handleGreet(w, r)

	// Portfolio endpoints
	case path == "/api/portfolio":
		h.handleGetPortfolio(w, r)
	case path == "/api/positions":
		h.handleGetPositions(w, r)

	// Recommendation endpoints
	case path == "/api/recommendations":
		h.handleGetRecommendations(w, r)
	case path == "/api/recommendations/pending":
		h.handleGetPendingRecommendations(w, r)
	case strings.HasPrefix(path, "/api/recommendations/") && strings.HasSuffix(path, "/approve"):
		h.handleApproveRecommendation(w, r)
	case strings.HasPrefix(path, "/api/recommendations/") && strings.HasSuffix(path, "/reject"):
		h.handleRejectRecommendation(w, r)

	// Analysis endpoints
	case path == "/api/analyze":
		h.handleAnalyzeStock(w, r)

	// Trade endpoints
	case path == "/api/trades":
		h.handleGetTrades(w, r)

	// Agent endpoints
	case path == "/api/agents/runs":
		h.handleGetAgentRuns(w, r)

	default:
		http.NotFound(w, r)
	}
}

// handleGreet handles the legacy greet endpoint
func (h *APIHandler) handleGreet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = "World"
	}

	greeting := h.app.Greet(name)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<span class="text-success">%s</span>`, greeting)
}

// handleGetPortfolio returns portfolio summary
func (h *APIHandler) handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positions, err := h.app.GetPositions()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, positions)
}

// handleGetRecommendations returns recommendations
func (h *APIHandler) handleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	recs, err := h.app.GetRecommendations(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, recs)
}

// handleGetPendingRecommendations returns pending recommendations
func (h *APIHandler) handleGetPendingRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recs, err := h.app.GetPendingRecommendations()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, recs)
}

// handleApproveRecommendation approves a recommendation
func (h *APIHandler) handleApproveRecommendation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /api/recommendations/{id}/approve
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		h.jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[3]

	if err := h.app.ApproveRecommendation(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "approved", "id": id})
}

// handleRejectRecommendation rejects a recommendation
func (h *APIHandler) handleRejectRecommendation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /api/recommendations/{id}/reject
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		h.jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id := parts[3]

	if err := h.app.RejectRecommendation(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "rejected", "id": id})
}

// handleAnalyzeStock triggers analysis of a stock
func (h *APIHandler) handleAnalyzeStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	rec, err := h.app.AnalyzeStock(req.Symbol)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, rec)
}

// handleGetTrades returns recent trades
func (h *APIHandler) handleGetTrades(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	trades, err := h.app.GetTrades(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, trades)
}

// handleGetAgentRuns returns recent agent runs
func (h *APIHandler) handleGetAgentRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	runs, err := h.app.GetAgentRuns(limit)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, runs)
}

// Helper functions

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
