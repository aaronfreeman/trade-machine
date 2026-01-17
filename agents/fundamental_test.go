package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"trade-machine/models"

	"github.com/shopspring/decimal"
)

func TestFundamentalAnalyst_Name(t *testing.T) {
	analyst := &FundamentalAnalyst{}
	if analyst.Name() != "Fundamental Analyst" {
		t.Errorf("Name() = %v, want 'Fundamental Analyst'", analyst.Name())
	}
}

func TestFundamentalAnalyst_Type(t *testing.T) {
	analyst := &FundamentalAnalyst{}
	if analyst.Type() != models.AgentTypeFundamental {
		t.Errorf("Type() = %v, want AgentTypeFundamental", analyst.Type())
	}
}

func TestNewFundamentalAnalyst(t *testing.T) {
	analyst := NewFundamentalAnalyst(nil, nil)
	if analyst == nil {
		t.Error("NewFundamentalAnalyst should not return nil")
	}
}

func TestFundamentalAnalystResponse_Parsing(t *testing.T) {
	resp := FundamentalAnalystResponse{
		Score:      75.5,
		Confidence: 80.0,
		Reasoning:  "Strong financials",
		KeyFactors: []string{"P/E below industry average", "Growing revenue"},
	}

	if resp.Score != 75.5 {
		t.Errorf("Score = %v, want 75.5", resp.Score)
	}
	if resp.Confidence != 80.0 {
		t.Errorf("Confidence = %v, want 80.0", resp.Confidence)
	}
	if len(resp.KeyFactors) != 2 {
		t.Errorf("KeyFactors length = %v, want 2", len(resp.KeyFactors))
	}
}

func TestFundamentalAnalyst_Analyze_Success(t *testing.T) {
	mockBedrock := &mockBedrockService{
		response: `{
			"score": 65.0,
			"confidence": 75.0,
			"reasoning": "Strong P/E ratio and healthy dividend yield indicate good value",
			"key_factors": ["Low P/E ratio", "Consistent dividend", "Stable earnings"]
		}`,
	}

	mockAlphaVantage := &mockAlphaVantageService{
		fundamentals: &models.Fundamentals{
			Symbol:        "AAPL",
			MarketCap:     decimal.NewFromInt(2500000000000),
			PERatio:       25.5,
			EPS:           decimal.NewFromFloat(6.05),
			DividendYield: 0.005,
			Week52High:    decimal.NewFromFloat(199.62),
			Week52Low:     decimal.NewFromFloat(164.08),
			Beta:          1.25,
			UpdatedAt:     time.Now(),
		},
	}

	analyst := NewFundamentalAnalyst(mockBedrock, mockAlphaVantage)
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if analysis.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want AAPL", analysis.Symbol)
	}
	if analysis.AgentType != models.AgentTypeFundamental {
		t.Errorf("AgentType = %v, want Fundamental", analysis.AgentType)
	}
	if analysis.Score != 65.0 {
		t.Errorf("Score = %v, want 65.0", analysis.Score)
	}
	if analysis.Confidence != 75.0 {
		t.Errorf("Confidence = %v, want 75.0", analysis.Confidence)
	}
	if analysis.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}

	keyFactors, ok := analysis.Data["key_factors"].([]string)
	if !ok {
		t.Error("key_factors should be []string")
	}
	if len(keyFactors) != 3 {
		t.Errorf("Expected 3 key factors, got %d", len(keyFactors))
	}
}

func TestFundamentalAnalyst_Analyze_AlphaVantageError(t *testing.T) {
	mockBedrock := &mockBedrockService{
		response: `{"score": 50, "confidence": 50, "reasoning": "test", "key_factors": []}`,
	}

	mockAlphaVantage := &mockAlphaVantageService{
		err: errors.New("API rate limit exceeded"),
	}

	analyst := NewFundamentalAnalyst(mockBedrock, mockAlphaVantage)
	ctx := context.Background()

	_, err := analyst.Analyze(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when AlphaVantage fails")
	}
}

func TestFundamentalAnalyst_Analyze_BedrockError(t *testing.T) {
	mockBedrock := &mockBedrockService{
		err: errors.New("Bedrock service unavailable"),
	}

	mockAlphaVantage := &mockAlphaVantageService{
		fundamentals: &models.Fundamentals{
			Symbol:    "AAPL",
			MarketCap: decimal.NewFromInt(2500000000000),
			PERatio:   25.5,
		},
	}

	analyst := NewFundamentalAnalyst(mockBedrock, mockAlphaVantage)
	ctx := context.Background()

	_, err := analyst.Analyze(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when Bedrock fails")
	}
}

func TestFundamentalAnalyst_Analyze_InvalidJSON(t *testing.T) {
	mockBedrock := &mockBedrockService{
		response: "This is not valid JSON, just plain text analysis",
	}

	mockAlphaVantage := &mockAlphaVantageService{
		fundamentals: &models.Fundamentals{
			Symbol:    "AAPL",
			MarketCap: decimal.NewFromInt(2500000000000),
			PERatio:   25.5,
		},
	}

	analyst := NewFundamentalAnalyst(mockBedrock, mockAlphaVantage)
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Analyze should not fail with invalid JSON: %v", err)
	}

	if analysis.Score != 0 {
		t.Errorf("Score = %v, want 0 for invalid JSON", analysis.Score)
	}
	if analysis.Confidence != 50 {
		t.Errorf("Confidence = %v, want 50 for invalid JSON", analysis.Confidence)
	}
	if analysis.Reasoning == "" {
		t.Error("Reasoning should contain raw response")
	}
}

func TestFundamentalAnalyst_Analyze_ScoreNormalization(t *testing.T) {
	mockBedrock := &mockBedrockService{
		response: `{
			"score": 150.0,
			"confidence": 120.0,
			"reasoning": "Extreme values test",
			"key_factors": ["test"]
		}`,
	}

	mockAlphaVantage := &mockAlphaVantageService{
		fundamentals: &models.Fundamentals{
			Symbol:  "TEST",
			PERatio: 15.0,
		},
	}

	analyst := NewFundamentalAnalyst(mockBedrock, mockAlphaVantage)
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "TEST")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if analysis.Score > 100 {
		t.Errorf("Score = %v, should be normalized to <= 100", analysis.Score)
	}
	if analysis.Confidence > 100 {
		t.Errorf("Confidence = %v, should be normalized to <= 100", analysis.Confidence)
	}
}
