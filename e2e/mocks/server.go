// Package mocks provides HTTP mock servers for external APIs used in E2E tests.
package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// MockServer provides configurable mock responses for all external APIs.
type MockServer struct {
	mu     sync.RWMutex
	server *httptest.Server

	// Response configurations
	bedrockResponses   map[string]BedrockResponse // key: agent type (fundamental, technical, news)
	alpacaBars         []AlpacaBar
	alpacaAccount      AlpacaAccount
	alpacaPositions    []AlpacaPosition
	alpacaOrderResult  *AlpacaOrder
	alphaFundamentals  *AlphaVantageFundamentals
	newsArticles       []NewsArticle
	fmpScreenerResults []FMPScreenerResult

	// Error injection
	bedrockError      error
	alpacaError       error
	alphaVantageError error
	newsAPIError      error
	fmpError          error

	// Request tracking for assertions
	requestLog []RequestLog
}

// RequestLog records incoming requests for test assertions.
type RequestLog struct {
	Method string
	Path   string
	Body   string
}

// NewMockServer creates a new mock server with default responses.
func NewMockServer() *MockServer {
	m := &MockServer{
		bedrockResponses: make(map[string]BedrockResponse),
		requestLog:       make([]RequestLog, 0),
	}
	m.setDefaults()
	m.server = httptest.NewServer(m)
	return m
}

// URL returns the mock server's base URL.
func (m *MockServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server.
func (m *MockServer) Close() {
	m.server.Close()
}

// ServeHTTP implements http.Handler to route requests to appropriate mock handlers.
func (m *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	// Log request
	body := ""
	if r.Body != nil {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		body = string(buf[:n])
	}
	m.requestLog = append(m.requestLog, RequestLog{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   body,
	})
	m.mu.Unlock()

	path := r.URL.Path

	// Route to appropriate handler based on path
	switch {
	case strings.Contains(path, "/model/") && strings.Contains(path, "/invoke"):
		m.handleBedrock(w, r)
	case strings.HasPrefix(path, "/v2/stocks") && strings.Contains(path, "/bars"):
		m.handleAlpacaBars(w, r)
	case path == "/v2/account":
		m.handleAlpacaAccount(w, r)
	case path == "/v2/positions":
		m.handleAlpacaPositions(w, r)
	case path == "/v2/orders":
		m.handleAlpacaOrders(w, r)
	case strings.Contains(path, "OVERVIEW"):
		m.handleAlphaVantage(w, r)
	case strings.Contains(path, "/everything") || strings.Contains(path, "newsapi"):
		m.handleNewsAPI(w, r)
	case strings.Contains(path, "/stock-screener"):
		m.handleFMPScreener(w, r)
	case strings.Contains(path, "/profile"):
		m.handleFMPProfile(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// GetRequestLog returns all logged requests for assertions.
func (m *MockServer) GetRequestLog() []RequestLog {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]RequestLog{}, m.requestLog...)
}

// ClearRequestLog clears the request log.
func (m *MockServer) ClearRequestLog() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestLog = make([]RequestLog, 0)
}

// SetBedrockResponse configures the response for a specific agent type.
func (m *MockServer) SetBedrockResponse(agentType string, resp BedrockResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bedrockResponses[agentType] = resp
}

// SetBedrockError configures Bedrock to return an error.
func (m *MockServer) SetBedrockError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bedrockError = err
}

// SetAlpacaBars configures the bars response for technical analysis.
func (m *MockServer) SetAlpacaBars(bars []AlpacaBar) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alpacaBars = bars
}

// SetAlpacaAccount configures the account response.
func (m *MockServer) SetAlpacaAccount(account AlpacaAccount) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alpacaAccount = account
}

// SetAlpacaPositions configures the positions response.
func (m *MockServer) SetAlpacaPositions(positions []AlpacaPosition) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alpacaPositions = positions
}

// SetAlpacaOrderResult configures the order placement response.
func (m *MockServer) SetAlpacaOrderResult(order *AlpacaOrder) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alpacaOrderResult = order
}

// SetAlpacaError configures Alpaca to return an error.
func (m *MockServer) SetAlpacaError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alpacaError = err
}

// SetAlphaVantageFundamentals configures the fundamentals response.
func (m *MockServer) SetAlphaVantageFundamentals(f *AlphaVantageFundamentals) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alphaFundamentals = f
}

// SetAlphaVantageError configures Alpha Vantage to return an error.
func (m *MockServer) SetAlphaVantageError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alphaVantageError = err
}

// SetNewsArticles configures the news articles response.
func (m *MockServer) SetNewsArticles(articles []NewsArticle) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.newsArticles = articles
}

// SetNewsAPIError configures NewsAPI to return an error.
func (m *MockServer) SetNewsAPIError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.newsAPIError = err
}

// SetFMPScreenerResults configures the FMP screener response.
func (m *MockServer) SetFMPScreenerResults(results []FMPScreenerResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fmpScreenerResults = results
}

// SetFMPError configures FMP to return an error.
func (m *MockServer) SetFMPError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fmpError = err
}

func (m *MockServer) setDefaults() {
	// Default Bedrock responses for each agent
	m.bedrockResponses["fundamental"] = BedrockResponse{
		Score:      45.0,
		Confidence: 75.0,
		Reasoning:  "Strong fundamentals with reasonable valuation.",
		KeyFactors: []string{"Low P/E ratio", "Consistent earnings growth"},
	}
	m.bedrockResponses["technical"] = BedrockResponse{
		Score:      30.0,
		Confidence: 65.0,
		Reasoning:  "Bullish technical signals with price above moving averages.",
		Signals:    []string{"Price above 50-day MA", "RSI neutral at 55"},
	}
	m.bedrockResponses["news"] = BedrockResponse{
		Score:      25.0,
		Confidence: 60.0,
		Reasoning:  "Mixed news sentiment with slight positive bias.",
		KeyThemes:  []string{"Earnings beat expectations", "New product launch"},
	}

	// Default Alpaca bars (100 days of data)
	m.alpacaBars = generateDefaultBars(100)

	// Default Alpaca account
	m.alpacaAccount = AlpacaAccount{
		ID:               "test-account-id",
		AccountNumber:    "TEST123456",
		Status:           "ACTIVE",
		Currency:         "USD",
		Cash:             "100000.00",
		PortfolioValue:   "150000.00",
		BuyingPower:      "200000.00",
		DaytradingBuying: "400000.00",
	}

	// Default empty positions
	m.alpacaPositions = []AlpacaPosition{}

	// Default Alpha Vantage fundamentals
	m.alphaFundamentals = &AlphaVantageFundamentals{
		Symbol:        "AAPL",
		Name:          "Apple Inc",
		Exchange:      "NASDAQ",
		Currency:      "USD",
		Country:       "USA",
		Sector:        "Technology",
		Industry:      "Consumer Electronics",
		MarketCap:     "3000000000000",
		PERatio:       "28.5",
		EPS:           "6.15",
		Beta:          "1.25",
		WeekHigh52:    "199.62",
		WeekLow52:     "143.90",
		DividendYield: "0.52",
	}

	// Default news articles (15 articles)
	m.newsArticles = generateDefaultNewsArticles(15)

	// Default FMP screener results
	m.fmpScreenerResults = []FMPScreenerResult{
		{Symbol: "AAPL", CompanyName: "Apple Inc", MarketCap: 3000000000000, PERatio: 28.5, PBRatio: 45.2, EPS: 6.15, DividendYield: 0.52, Sector: "Technology"},
		{Symbol: "MSFT", CompanyName: "Microsoft Corp", MarketCap: 2800000000000, PERatio: 32.1, PBRatio: 12.3, EPS: 9.21, DividendYield: 0.75, Sector: "Technology"},
		{Symbol: "JNJ", CompanyName: "Johnson & Johnson", MarketCap: 400000000000, PERatio: 15.2, PBRatio: 5.8, EPS: 10.15, DividendYield: 2.95, Sector: "Healthcare"},
	}
}

func (m *MockServer) handleBedrock(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.bedrockError
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine which agent is calling based on the request body
	// The prompt will contain agent-specific keywords
	agentType := "fundamental" // default
	if strings.Contains(r.URL.Path, "technical") || strings.Contains(r.Header.Get("X-Agent-Type"), "technical") {
		agentType = "technical"
	} else if strings.Contains(r.URL.Path, "news") || strings.Contains(r.Header.Get("X-Agent-Type"), "news") {
		agentType = "news"
	}

	m.mu.RLock()
	resp := m.bedrockResponses[agentType]
	m.mu.RUnlock()

	// Format as Claude response
	claudeResp := formatClaudeResponse(resp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claudeResp)
}

func (m *MockServer) handleAlpacaBars(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.alpacaError
	bars := m.alpacaBars
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract symbol from path
	parts := strings.Split(r.URL.Path, "/")
	symbol := "AAPL"
	for i, p := range parts {
		if p == "stocks" && i+1 < len(parts) {
			symbol = parts[i+1]
			break
		}
	}

	resp := map[string]interface{}{
		"bars": map[string]interface{}{
			symbol: bars,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (m *MockServer) handleAlpacaAccount(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.alpacaError
	account := m.alpacaAccount
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (m *MockServer) handleAlpacaPositions(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.alpacaError
	positions := m.alpacaPositions
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func (m *MockServer) handleAlpacaOrders(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.alpacaError
	order := m.alpacaOrderResult
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		order = &AlpacaOrder{
			ID:            "test-order-id",
			ClientOrderID: "test-client-order-id",
			Status:        "filled",
			Symbol:        "AAPL",
			Qty:           "10",
			FilledQty:     "10",
			Side:          "buy",
			Type:          "market",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (m *MockServer) handleAlphaVantage(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.alphaVantageError
	fundamentals := m.alphaFundamentals
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fundamentals)
}

func (m *MockServer) handleNewsAPI(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.newsAPIError
	articles := m.newsArticles
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"status":       "ok",
		"totalResults": len(articles),
		"articles":     articles,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (m *MockServer) handleFMPScreener(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.fmpError
	results := m.fmpScreenerResults
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (m *MockServer) handleFMPProfile(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	err := m.fmpError
	m.mu.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a simple profile
	profile := []map[string]interface{}{
		{
			"symbol":      "AAPL",
			"companyName": "Apple Inc",
			"industry":    "Consumer Electronics",
			"sector":      "Technology",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func formatClaudeResponse(resp BedrockResponse) map[string]interface{} {
	// Build the content JSON that Claude would return
	content := map[string]interface{}{
		"score":      resp.Score,
		"confidence": resp.Confidence,
		"reasoning":  resp.Reasoning,
	}
	if len(resp.KeyFactors) > 0 {
		content["key_factors"] = resp.KeyFactors
	}
	if len(resp.Signals) > 0 {
		content["signals"] = resp.Signals
	}
	if len(resp.KeyThemes) > 0 {
		content["key_themes"] = resp.KeyThemes
	}

	contentJSON, _ := json.Marshal(content)

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(contentJSON),
			},
		},
		"stop_reason": "end_turn",
	}
}

func generateDefaultBars(count int) []AlpacaBar {
	bars := make([]AlpacaBar, count)
	basePrice := 150.0
	for i := 0; i < count; i++ {
		// Generate slightly varying prices
		variance := float64(i%10) - 5
		price := basePrice + variance
		bars[i] = AlpacaBar{
			Timestamp: "2024-01-" + fmt.Sprintf("%02d", (i%28)+1) + "T00:00:00Z",
			Open:      price - 1,
			High:      price + 2,
			Low:       price - 2,
			Close:     price,
			Volume:    1000000 + int64(i*10000),
		}
	}
	return bars
}

func generateDefaultNewsArticles(count int) []NewsArticle {
	articles := make([]NewsArticle, count)
	titles := []string{
		"Company Reports Strong Quarterly Earnings",
		"New Product Launch Expected to Boost Sales",
		"Analyst Upgrades Stock to Buy",
		"Market Share Continues to Grow",
		"Innovation Pipeline Looks Promising",
	}
	for i := 0; i < count; i++ {
		articles[i] = NewsArticle{
			Source:      map[string]string{"name": "Financial Times"},
			Author:      "Test Author",
			Title:       titles[i%len(titles)],
			Description: "This is a test article description for E2E testing.",
			URL:         "https://example.com/article/" + fmt.Sprintf("%d", i),
			PublishedAt: "2024-01-15T12:00:00Z",
		}
	}
	return articles
}
