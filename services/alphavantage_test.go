package services

import (
	"encoding/json"
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
