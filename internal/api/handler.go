package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"trade-machine/config"
	"trade-machine/internal/app"
	"trade-machine/internal/settings"
	"trade-machine/models"
	"trade-machine/services"
	"trade-machine/templates"
	"trade-machine/templates/components"
	"trade-machine/templates/partials"

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
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.PositionsList(positions), r)
		return
	}

	h.jsonResponse(w, positions)
}

// HandleGetRecommendations returns recommendations
func (h *Handler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	limit := h.ParseLimitParam(r, 50)

	recs, err := h.app.GetRecommendations(limit)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.RecommendationsList(recs), r)
		return
	}

	h.jsonResponse(w, recs)
}

// HandleGetPendingRecommendations returns pending recommendations
func (h *Handler) HandleGetPendingRecommendations(w http.ResponseWriter, r *http.Request) {
	recs, err := h.app.GetPendingRecommendations()
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.RecommendationsList(recs), r)
		return
	}

	h.jsonResponse(w, recs)
}

// HandleApproveRecommendation approves a recommendation
func (h *Handler) HandleApproveRecommendation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Missing recommendation ID", r)
			return
		}
		h.jsonError(w, "Missing recommendation ID", http.StatusBadRequest)
		return
	}

	if err := h.app.ApproveRecommendation(id); err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		// Return the updated recommendation card
		rec, err := h.app.GetRecommendationByID(id)
		if err != nil {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.htmlResponse(w, partials.RecommendationCardUpdated(*rec), r)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "approved", "id": id})
}

// HandleRejectRecommendation rejects a recommendation
func (h *Handler) HandleRejectRecommendation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Missing recommendation ID", r)
			return
		}
		h.jsonError(w, "Missing recommendation ID", http.StatusBadRequest)
		return
	}

	if err := h.app.RejectRecommendation(id); err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		// Return the updated recommendation card
		rec, err := h.app.GetRecommendationByID(id)
		if err != nil {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.htmlResponse(w, partials.RecommendationCardUpdated(*rec), r)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "rejected", "id": id})
}

// HandleAnalyzeStock triggers analysis of a stock
func (h *Handler) HandleAnalyzeStock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string `json:"symbol"`
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&req)
	} else {
		_ = r.ParseForm()
		req.Symbol = r.FormValue("symbol")
	}

	if req.Symbol == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Symbol is required", r)
			return
		}
		h.jsonError(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))

	if err := h.ValidateSymbol(req.Symbol); err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	rec, err := h.app.AnalyzeStock(req.Symbol)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.AnalyzeResult(rec), r)
		return
	}

	h.jsonResponse(w, rec)
}

// HandleGetTrades returns recent trades
func (h *Handler) HandleGetTrades(w http.ResponseWriter, r *http.Request) {
	limit := h.ParseLimitParam(r, 50)

	trades, err := h.app.GetTrades(limit)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.TradesList(trades), r)
		return
	}

	h.jsonResponse(w, trades)
}

// HandleGetAgentRuns returns recent agent runs
func (h *Handler) HandleGetAgentRuns(w http.ResponseWriter, r *http.Request) {
	limit := h.ParseLimitParam(r, 50)

	runs, err := h.app.GetAgentRuns(limit)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.AgentRunsList(runs), r)
		return
	}

	h.jsonResponse(w, runs)
}

// Helper functions

// isHTMXRequest checks if the request is from HTMX
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// templComponent matches the templ.Component interface
type templComponent interface {
	Render(ctx context.Context, w io.Writer) error
}

// htmlResponse renders a templ component as HTML
func (h *Handler) htmlResponse(w http.ResponseWriter, component templComponent, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

// htmlError renders an error state as HTML
func (h *Handler) htmlError(w http.ResponseWriter, message string, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	components.ErrorState(message).Render(r.Context(), w)
}

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

// HandleRunScreener triggers a full screener run
func (h *Handler) HandleRunScreener(w http.ResponseWriter, r *http.Request) {
	if h.app.Screener() == nil {
		if isHTMXRequest(r) {
			h.htmlResponse(w, partials.ScreenerNotConfigured(), r)
			return
		}
		h.jsonError(w, "Screener not configured", http.StatusServiceUnavailable)
		return
	}

	run, err := h.app.RunScreener()
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		picks, _ := h.app.GetTopPicks()
		h.htmlResponse(w, partials.TodaysPicks(run, picks), r)
		return
	}

	h.jsonResponse(w, run)
}

// HandleGetLatestScreenerRun returns the most recent screener run
func (h *Handler) HandleGetLatestScreenerRun(w http.ResponseWriter, r *http.Request) {
	if h.app.Screener() == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Screener not configured", r)
			return
		}
		h.jsonError(w, "Screener not configured", http.StatusServiceUnavailable)
		return
	}

	run, err := h.app.GetLatestScreenerRun()
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if run == nil {
		if isHTMXRequest(r) {
			h.htmlResponse(w, partials.ScreenerEmpty(), r)
			return
		}
		h.jsonResponse(w, map[string]interface{}{"run": nil})
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.ScreenerRunResult(run), r)
		return
	}

	h.jsonResponse(w, run)
}

// HandleGetScreenerRuns returns screener run history
func (h *Handler) HandleGetScreenerRuns(w http.ResponseWriter, r *http.Request) {
	if h.app.Screener() == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Screener not configured", r)
			return
		}
		h.jsonError(w, "Screener not configured", http.StatusServiceUnavailable)
		return
	}

	limit := h.ParseLimitParam(r, 10)

	runs, err := h.app.GetScreenerRunHistory(limit)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.ScreenerRunsList(runs), r)
		return
	}

	h.jsonResponse(w, runs)
}

// HandleGetScreenerRun returns a specific screener run by ID
func (h *Handler) HandleGetScreenerRun(w http.ResponseWriter, r *http.Request) {
	if h.app.Screener() == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Screener not configured", r)
			return
		}
		h.jsonError(w, "Screener not configured", http.StatusServiceUnavailable)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Missing screener run ID", r)
			return
		}
		h.jsonError(w, "Missing screener run ID", http.StatusBadRequest)
		return
	}

	run, err := h.app.GetScreenerRun(id)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if run == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Screener run not found", r)
			return
		}
		h.jsonError(w, "Screener run not found", http.StatusNotFound)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.ScreenerRunResult(run), r)
		return
	}

	h.jsonResponse(w, run)
}

// HandleGetTopPicks returns the top picks from the latest completed screener run
func (h *Handler) HandleGetTopPicks(w http.ResponseWriter, r *http.Request) {
	if h.app.Screener() == nil {
		if isHTMXRequest(r) {
			h.htmlResponse(w, partials.ScreenerNotConfigured(), r)
			return
		}
		h.jsonError(w, "Screener not configured", http.StatusServiceUnavailable)
		return
	}

	run, err := h.app.GetLatestScreenerRun()
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	picks, err := h.app.GetTopPicks()
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.TodaysPicks(run, picks), r)
		return
	}

	h.jsonResponse(w, picks)
}

// Ensure models are exported for JSON serialization
var _ = models.Position{}
var _ = models.Trade{}
var _ = models.Recommendation{}
var _ = models.AgentRun{}
var _ = models.ScreenerRun{}

// HandleGetSettings returns masked API key settings
func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settingsStore := h.app.Settings()
	if settingsStore == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Settings not available", r)
			return
		}
		h.jsonError(w, "Settings not available", http.StatusServiceUnavailable)
		return
	}

	masked := settingsStore.GetMaskedSettings()

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.SettingsForm(masked), r)
		return
	}

	h.jsonResponse(w, masked)
}

// HandleUpdateAPIKey updates a single API key configuration
func (h *Handler) HandleUpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	settingsStore := h.app.Settings()
	if settingsStore == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Settings not available", r)
			return
		}
		h.jsonError(w, "Settings not available", http.StatusServiceUnavailable)
		return
	}

	var req settings.APIKeyConfig
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if isHTMXRequest(r) {
				h.htmlError(w, "Invalid JSON request", r)
				return
			}
			h.jsonError(w, "Invalid JSON request", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			if isHTMXRequest(r) {
				h.htmlError(w, "Failed to parse form", r)
				return
			}
			h.jsonError(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		req.ServiceName = settings.ServiceName(r.FormValue("service_name"))
		req.APIKey = r.FormValue("api_key")
		req.APISecret = r.FormValue("api_secret")
		req.BaseURL = r.FormValue("base_url")
		req.Region = r.FormValue("region")
		req.ModelID = r.FormValue("model_id")
	}

	if req.ServiceName == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Service name is required", r)
			return
		}
		h.jsonError(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// Check if at least one field has a value to update
	hasUpdate := req.APIKey != "" || req.APISecret != "" || req.BaseURL != "" || req.Region != "" || req.ModelID != ""
	if !hasUpdate {
		// No fields to update - just return current state
		if isHTMXRequest(r) {
			masked := settingsStore.GetMaskedSettings()
			h.htmlResponse(w, partials.SettingsForm(masked), r)
			return
		}
		h.jsonResponse(w, map[string]string{"status": "no changes", "service": string(req.ServiceName)})
		return
	}

	// Merge with existing config to preserve fields not being updated
	existingConfig := settingsStore.GetAPIKey(req.ServiceName)
	if existingConfig != nil {
		if req.APIKey == "" {
			req.APIKey = existingConfig.APIKey
		}
		if req.APISecret == "" {
			req.APISecret = existingConfig.APISecret
		}
		if req.BaseURL == "" {
			req.BaseURL = existingConfig.BaseURL
		}
		if req.Region == "" {
			req.Region = existingConfig.Region
		}
		if req.ModelID == "" {
			req.ModelID = existingConfig.ModelID
		}
	}

	if err := settingsStore.SetAPIKey(&req); err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		masked := settingsStore.GetMaskedSettings()
		h.htmlResponse(w, partials.SettingsForm(masked), r)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "saved", "service": string(req.ServiceName)})
}

// HandleTestAPIKey tests if an API key is valid
func (h *Handler) HandleTestAPIKey(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")
	if service == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Service name is required", r)
			return
		}
		h.jsonError(w, "Service name is required", http.StatusBadRequest)
		return
	}

	settingsStore := h.app.Settings()
	if settingsStore == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Settings not available", r)
			return
		}
		h.jsonError(w, "Settings not available", http.StatusServiceUnavailable)
		return
	}

	serviceName := settings.ServiceName(service)
	config := settingsStore.GetAPIKey(serviceName)
	if config == nil {
		if isHTMXRequest(r) {
			h.htmlResponse(w, partials.ServiceStatus(serviceName, false, "Not configured"), r)
			return
		}
		h.jsonError(w, "Service not configured", http.StatusNotFound)
		return
	}

	validator := settings.NewValidator()
	result, err := validator.ValidateAPIKey(r.Context(), config)
	if err != nil {
		if isHTMXRequest(r) {
			h.htmlResponse(w, partials.ServiceStatus(serviceName, false, err.Error()), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.ServiceStatus(serviceName, result.Valid, result.Message), r)
		return
	}

	h.jsonResponse(w, result)
}

// HandleDeleteAPIKey removes an API key configuration
func (h *Handler) HandleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "service")
	if service == "" {
		if isHTMXRequest(r) {
			h.htmlError(w, "Service name is required", r)
			return
		}
		h.jsonError(w, "Service name is required", http.StatusBadRequest)
		return
	}

	settingsStore := h.app.Settings()
	if settingsStore == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Settings not available", r)
			return
		}
		h.jsonError(w, "Settings not available", http.StatusServiceUnavailable)
		return
	}

	serviceName := settings.ServiceName(service)
	if err := settingsStore.DeleteAPIKey(serviceName); err != nil {
		if isHTMXRequest(r) {
			h.htmlError(w, err.Error(), r)
			return
		}
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMXRequest(r) {
		masked := settingsStore.GetMaskedSettings()
		h.htmlResponse(w, partials.SettingsForm(masked), r)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "deleted", "service": service})
}

// HandleResetSettings removes all API key configurations (for E2E testing)
func (h *Handler) HandleResetSettings(w http.ResponseWriter, r *http.Request) {
	settingsStore := h.app.Settings()
	if settingsStore == nil {
		h.jsonError(w, "Settings not available", http.StatusServiceUnavailable)
		return
	}

	if err := settingsStore.ResetAll(); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonResponse(w, map[string]string{"status": "reset"})
}

// HandleSettingsPage renders the settings page
func (h *Handler) HandleSettingsPage(w http.ResponseWriter, r *http.Request) {
	settingsStore := h.app.Settings()
	if settingsStore == nil {
		if isHTMXRequest(r) {
			h.htmlError(w, "Settings not available", r)
			return
		}
		h.jsonError(w, "Settings not available", http.StatusServiceUnavailable)
		return
	}

	masked := settingsStore.GetMaskedSettings()

	if isHTMXRequest(r) {
		h.htmlResponse(w, partials.SettingsForm(masked), r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Index().Render(r.Context(), w)
}
