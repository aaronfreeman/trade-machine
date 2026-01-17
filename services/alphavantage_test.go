package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAlphaVantageService(t *testing.T) {
	service := NewAlphaVantageService("test-api-key")
	if service == nil {
		t.Error("NewAlphaVantageService should not return nil")
	}
	if service.apiKey != "test-api-key" {
		t.Errorf("apiKey = %v, want 'test-api-key'", service.apiKey)
	}
	if service.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if service.baseURL != "https://www.alphavantage.co/query" {
		t.Errorf("baseURL = %v, want 'https://www.alphavantage.co/query'", service.baseURL)
	}
}

func TestOverviewResponse_Deserialization(t *testing.T) {
	jsonResponse := `{
		"Symbol": "AAPL",
		"Name": "Apple Inc",
		"Description": "Apple Inc. designs, manufactures, and markets smartphones...",
		"Exchange": "NASDAQ",
		"Currency": "USD",
		"Country": "USA",
		"Sector": "TECHNOLOGY",
		"Industry": "CONSUMER ELECTRONICS",
		"MarketCapitalization": "2500000000000",
		"PERatio": "28.5",
		"PEGRatio": "2.3",
		"BookValue": "4.25",
		"DividendPerShare": "0.96",
		"DividendYield": "0.005",
		"EPS": "6.05",
		"RevenuePerShareTTM": "24.32",
		"ProfitMargin": "0.255",
		"Beta": "1.25",
		"52WeekHigh": "199.62",
		"52WeekLow": "164.08",
		"AnalystTargetPrice": "190.50"
	}`

	var resp OverviewResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal OverviewResponse: %v", err)
	}

	if resp.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", resp.Symbol)
	}
	if resp.Name != "Apple Inc" {
		t.Errorf("Name = %v, want 'Apple Inc'", resp.Name)
	}
	if resp.PERatio != "28.5" {
		t.Errorf("PERatio = %v, want '28.5'", resp.PERatio)
	}
	if resp.MarketCap != "2500000000000" {
		t.Errorf("MarketCap = %v, want '2500000000000'", resp.MarketCap)
	}
	if resp.Week52High != "199.62" {
		t.Errorf("Week52High = %v, want '199.62'", resp.Week52High)
	}
	if resp.Week52Low != "164.08" {
		t.Errorf("Week52Low = %v, want '164.08'", resp.Week52Low)
	}
}

func TestNewsResponse_Deserialization(t *testing.T) {
	jsonResponse := `{
		"items": "10",
		"feed": [
			{
				"title": "Apple announces new product",
				"url": "https://example.com/news/1",
				"summary": "Apple has announced a revolutionary new product...",
				"source": "TechNews",
				"time_published": "20240115T120000",
				"authors": ["John Doe"],
				"overall_sentiment_label": "Bullish",
				"overall_sentiment_score": 0.75
			}
		]
	}`

	var resp NewsResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal NewsResponse: %v", err)
	}

	if resp.Items != "10" {
		t.Errorf("Items = %v, want '10'", resp.Items)
	}
	if len(resp.Feed) != 1 {
		t.Errorf("Feed length = %v, want 1", len(resp.Feed))
	}
	if resp.Feed[0].Title != "Apple announces new product" {
		t.Errorf("Feed[0].Title = %v, want 'Apple announces new product'", resp.Feed[0].Title)
	}
	if resp.Feed[0].SentimentScore != 0.75 {
		t.Errorf("Feed[0].SentimentScore = %v, want 0.75", resp.Feed[0].SentimentScore)
	}
}

func TestQuoteResponse_Deserialization(t *testing.T) {
	jsonResponse := `{
		"Global Quote": {
			"01. symbol": "AAPL",
			"02. open": "185.50",
			"03. high": "188.00",
			"04. low": "184.00",
			"05. price": "187.50",
			"06. volume": "50000000",
			"07. latest trading day": "2024-01-15",
			"08. previous close": "185.00",
			"09. change": "2.50",
			"10. change percent": "1.35%"
		}
	}`

	var resp QuoteResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal QuoteResponse: %v", err)
	}

	if resp.GlobalQuote.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", resp.GlobalQuote.Symbol)
	}
	if resp.GlobalQuote.Price != "187.50" {
		t.Errorf("Price = %v, want '187.50'", resp.GlobalQuote.Price)
	}
	if resp.GlobalQuote.Volume != "50000000" {
		t.Errorf("Volume = %v, want '50000000'", resp.GlobalQuote.Volume)
	}
	if resp.GlobalQuote.Change != "2.50" {
		t.Errorf("Change = %v, want '2.50'", resp.GlobalQuote.Change)
	}
}

func TestAlphaVantageService_GetFundamentals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OverviewResponse{
			Symbol:        "AAPL",
			Name:          "Apple Inc",
			MarketCap:     "2500000000000",
			PERatio:       "28.5",
			EPS:           "6.05",
			DividendYield: "0.005",
			Week52High:    "199.62",
			Week52Low:     "164.08",
			Beta:          "1.25",
		})
	}))
	defer server.Close()

	service := NewAlphaVantageService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	fundamentals, err := service.GetFundamentals(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetFundamentals failed: %v", err)
	}

	if fundamentals.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", fundamentals.Symbol)
	}
	if fundamentals.PERatio != 28.5 {
		t.Errorf("PERatio = %v, want 28.5", fundamentals.PERatio)
	}
}

func TestAlphaVantageService_GetNews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := NewsResponse{
			Items: "2",
			Feed: []struct {
				Title            string   `json:"title"`
				URL              string   `json:"url"`
				Summary          string   `json:"summary"`
				Source           string   `json:"source"`
				TimePublished    string   `json:"time_published"`
				Authors          []string `json:"authors"`
				OverallSentiment string   `json:"overall_sentiment_label"`
				SentimentScore   float64  `json:"overall_sentiment_score"`
			}{
				{
					Title:          "Apple News 1",
					URL:            "https://example.com/1",
					Summary:        "Summary 1",
					Source:         "TechNews",
					TimePublished:  "20240115T120000",
					Authors:        []string{"John Doe"},
					SentimentScore: 0.8,
				},
				{
					Title:          "Apple News 2",
					URL:            "https://example.com/2",
					Summary:        "Summary 2",
					Source:         "BusinessNews",
					TimePublished:  "20240115T130000",
					Authors:        []string{"Jane Smith"},
					SentimentScore: 0.6,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewAlphaVantageService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	articles, err := service.GetNews(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetNews failed: %v", err)
	}

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}
	if articles[0].Title != "Apple News 1" {
		t.Errorf("First article title = %v, want 'Apple News 1'", articles[0].Title)
	}
}

func TestAlphaVantageService_GetQuote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := QuoteResponse{}
		response.GlobalQuote.Symbol = "AAPL"
		response.GlobalQuote.Price = "187.50"
		response.GlobalQuote.Volume = "50000000"
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewAlphaVantageService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	quote, err := service.GetQuote(ctx, "AAPL")
	if err != nil {
		t.Fatalf("GetQuote failed: %v", err)
	}

	if quote.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", quote.Symbol)
	}
	if quote.Volume != 50000000 {
		t.Errorf("Volume = %v, want 50000000", quote.Volume)
	}
}
