package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"trade-machine/config"
	"trade-machine/models"

	marketdata "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

func TestTechnicalAnalyst_CalculateSMA(t *testing.T) {
	analyst := &TechnicalAnalyst{}

	tests := []struct {
		name   string
		prices []float64
		period int
		want   float64
	}{
		{
			name:   "simple 5-day SMA",
			prices: []float64{10, 20, 30, 40, 50},
			period: 5,
			want:   30.0, // (10+20+30+40+50)/5
		},
		{
			name:   "3-day SMA from longer series",
			prices: []float64{10, 20, 30, 40, 50},
			period: 3,
			want:   40.0, // (30+40+50)/3
		},
		{
			name:   "period equals length",
			prices: []float64{100, 200, 300},
			period: 3,
			want:   200.0, // (100+200+300)/3
		},
		{
			name:   "period too long returns zero",
			prices: []float64{10, 20},
			period: 5,
			want:   0.0,
		},
		{
			name:   "single value",
			prices: []float64{100},
			period: 1,
			want:   100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyst.calculateSMA(tt.prices, tt.period)
			if got != tt.want {
				t.Errorf("calculateSMA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTechnicalAnalyst_CalculateRSI(t *testing.T) {
	analyst := &TechnicalAnalyst{}

	tests := []struct {
		name    string
		prices  []float64
		period  int
		wantMin float64
		wantMax float64
	}{
		{
			name:    "uptrending prices - high RSI",
			prices:  []float64{40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55},
			period:  14,
			wantMin: 70.0, // Should be high (bullish)
			wantMax: 100.0,
		},
		{
			name:    "downtrending prices - low RSI",
			prices:  []float64{55, 54, 53, 52, 51, 50, 49, 48, 47, 46, 45, 44, 43, 42, 41, 40},
			period:  14,
			wantMin: 0.0,
			wantMax: 30.0, // Should be low (bearish)
		},
		{
			name:    "insufficient data returns neutral",
			prices:  []float64{100, 101, 102},
			period:  14,
			wantMin: 50.0,
			wantMax: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyst.calculateRSI(tt.prices, tt.period)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateRSI() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTechnicalAnalyst_CalculateEMA(t *testing.T) {
	analyst := &TechnicalAnalyst{}

	tests := []struct {
		name   string
		prices []float64
		period int
	}{
		{
			name:   "simple EMA calculation",
			prices: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			period: 5,
		},
		{
			name:   "short series",
			prices: []float64{100, 110, 120},
			period: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyst.calculateEMA(tt.prices, tt.period)

			// EMA should have same length as input
			if len(got) != len(tt.prices) {
				t.Errorf("EMA length = %v, want %v", len(got), len(tt.prices))
			}

			// Last EMA value should be reasonable (close to recent prices)
			if len(got) > 0 {
				lastEMA := got[len(got)-1]
				lastPrice := tt.prices[len(tt.prices)-1]
				// EMA should be within 50% of last price for trending data
				if lastEMA < lastPrice*0.5 || lastEMA > lastPrice*1.5 {
					t.Errorf("Last EMA %v seems unreasonable for last price %v", lastEMA, lastPrice)
				}
			}
		})
	}
}

func TestTechnicalAnalyst_CalculateIndicators(t *testing.T) {
	analyst := &TechnicalAnalyst{}

	// Generate test price data (100 days)
	prices := make([]float64, 100)
	for i := 0; i < 100; i++ {
		prices[i] = 100 + float64(i)*0.5 // Uptrending
	}

	indicators := analyst.calculateIndicators(prices)

	// Check that all expected keys exist
	expectedKeys := []string{"rsi", "sma20", "sma50", "macd", "macd_signal", "macd_histogram", "high", "low"}
	for _, key := range expectedKeys {
		if _, ok := indicators[key]; !ok {
			t.Errorf("Missing indicator key: %s", key)
		}
	}

	// RSI should be between 0 and 100
	rsi := indicators["rsi"].(float64)
	if rsi < 0 || rsi > 100 {
		t.Errorf("RSI = %v, should be between 0 and 100", rsi)
	}

	// SMAs should be positive
	sma20 := indicators["sma20"].(float64)
	sma50 := indicators["sma50"].(float64)
	if sma20 <= 0 {
		t.Errorf("SMA20 = %v, should be positive", sma20)
	}
	if sma50 <= 0 {
		t.Errorf("SMA50 = %v, should be positive", sma50)
	}

	// For uptrending data, SMA20 should be > SMA50
	if sma20 <= sma50 {
		t.Errorf("For uptrend, SMA20 (%v) should be > SMA50 (%v)", sma20, sma50)
	}

	// High should be >= Low
	high := indicators["high"].(float64)
	low := indicators["low"].(float64)
	if high < low {
		t.Errorf("High (%v) should be >= Low (%v)", high, low)
	}
}

func TestTechnicalAnalyst_Name(t *testing.T) {
	analyst := &TechnicalAnalyst{}
	if analyst.Name() != "Technical Analyst" {
		t.Errorf("Name() = %v, want 'Technical Analyst'", analyst.Name())
	}
}

func TestTechnicalAnalyst_Type(t *testing.T) {
	analyst := &TechnicalAnalyst{}
	if string(analyst.Type()) != "technical" {
		t.Errorf("Type() = %v, want 'technical'", analyst.Type())
	}
}

func TestNewTechnicalAnalyst(t *testing.T) {
	analyst := NewTechnicalAnalyst(nil, nil, config.NewTestConfig())
	if analyst == nil {
		t.Error("NewTechnicalAnalyst should not return nil")
	}
}

func TestTechnicalAnalyst_Analyze_Success(t *testing.T) {
	mockLLM := &mockLLMService{
		response: `{
			"score": 55.0,
			"confidence": 75.0,
			"reasoning": "Bullish MACD crossover with RSI in neutral territory",
			"signals": ["MACD bullish crossover", "Price above SMA20", "RSI neutral"]
		}`,
	}

	bars := make([]marketdata.Bar, 100)
	basePrice := 100.0
	for i := 0; i < 100; i++ {
		bars[i] = marketdata.Bar{
			Timestamp: time.Now().AddDate(0, 0, -100+i),
			Open:      basePrice + float64(i)*0.5,
			High:      basePrice + float64(i)*0.5 + 1.0,
			Low:       basePrice + float64(i)*0.5 - 1.0,
			Close:     basePrice + float64(i)*0.5,
			Volume:    1000000,
		}
	}

	mockAlpaca := &mockAlpacaService{
		bars: bars,
	}

	analyst := NewTechnicalAnalyst(mockLLM, mockAlpaca, config.NewTestConfig())
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if analysis.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want AAPL", analysis.Symbol)
	}
	if analysis.AgentType != models.AgentTypeTechnical {
		t.Errorf("AgentType = %v, want Technical", analysis.AgentType)
	}
	if analysis.Score != 55.0 {
		t.Errorf("Score = %v, want 55.0", analysis.Score)
	}
	if analysis.Confidence != 75.0 {
		t.Errorf("Confidence = %v, want 75.0", analysis.Confidence)
	}

	signals, ok := analysis.Data["signals"].([]string)
	if !ok {
		t.Error("signals should be []string")
	}
	if len(signals) != 3 {
		t.Errorf("Expected 3 signals, got %d", len(signals))
	}
}

func TestTechnicalAnalyst_Analyze_InsufficientData(t *testing.T) {
	mockLLM := &mockLLMService{
		response: `{"score": 0, "confidence": 20, "reasoning": "test", "signals": []}`,
	}

	bars := make([]marketdata.Bar, 30)
	for i := 0; i < 30; i++ {
		bars[i] = marketdata.Bar{
			Close:  100.0 + float64(i),
			Volume: 1000000,
		}
	}

	mockAlpaca := &mockAlpacaService{
		bars: bars,
	}

	analyst := NewTechnicalAnalyst(mockLLM, mockAlpaca, config.NewTestConfig())
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "NEWSTOCK")
	if err != nil {
		t.Fatalf("Analyze should not fail with insufficient data: %v", err)
	}

	if analysis.Score != 0 {
		t.Errorf("Score = %v, want 0 for insufficient data", analysis.Score)
	}
	if analysis.Confidence != 20 {
		t.Errorf("Confidence = %v, want 20 for insufficient data", analysis.Confidence)
	}
	if analysis.Reasoning != "Insufficient price history for technical analysis" {
		t.Errorf("Reasoning should indicate insufficient data")
	}
}

func TestTechnicalAnalyst_Analyze_AlpacaError(t *testing.T) {
	mockLLM := &mockLLMService{
		response: `{"score": 0, "confidence": 50, "reasoning": "test", "signals": []}`,
	}

	mockAlpaca := &mockAlpacaService{
		err: errors.New("Alpaca API unavailable"),
	}

	analyst := NewTechnicalAnalyst(mockLLM, mockAlpaca, config.NewTestConfig())
	ctx := context.Background()

	_, err := analyst.Analyze(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when Alpaca fails")
	}
}

func TestTechnicalAnalyst_Analyze_LLMError(t *testing.T) {
	mockLLM := &mockLLMService{
		err: errors.New("LLM service unavailable"),
	}

	bars := make([]marketdata.Bar, 100)
	for i := 0; i < 100; i++ {
		bars[i] = marketdata.Bar{
			Close:  100.0 + float64(i)*0.5,
			Volume: 1000000,
		}
	}

	mockAlpaca := &mockAlpacaService{
		bars: bars,
	}

	analyst := NewTechnicalAnalyst(mockLLM, mockAlpaca, config.NewTestConfig())
	ctx := context.Background()

	_, err := analyst.Analyze(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when LLM fails")
	}
}

func TestTechnicalAnalyst_Analyze_InvalidJSON(t *testing.T) {
	mockLLM := &mockLLMService{
		response: "Plain text technical analysis, not JSON",
	}

	bars := make([]marketdata.Bar, 100)
	for i := 0; i < 100; i++ {
		bars[i] = marketdata.Bar{
			Close:  100.0 + float64(i)*0.5,
			Volume: 1000000,
		}
	}

	mockAlpaca := &mockAlpacaService{
		bars: bars,
	}

	analyst := NewTechnicalAnalyst(mockLLM, mockAlpaca, config.NewTestConfig())
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
}

func TestTechnicalAnalyst_IsAvailable_Success(t *testing.T) {
	bars := make([]marketdata.Bar, 1)
	bars[0] = marketdata.Bar{Close: 100.0, Volume: 1000000}

	mockAlpaca := &mockAlpacaService{
		bars: bars,
	}

	analyst := NewTechnicalAnalyst(nil, mockAlpaca, config.NewTestConfig())
	ctx := context.Background()

	if !analyst.IsAvailable(ctx) {
		t.Error("IsAvailable should return true when service is healthy")
	}
}

func TestTechnicalAnalyst_IsAvailable_Failure(t *testing.T) {
	mockAlpaca := &mockAlpacaService{
		err: errors.New("service unavailable"),
	}

	analyst := NewTechnicalAnalyst(nil, mockAlpaca, config.NewTestConfig())
	ctx := context.Background()

	if analyst.IsAvailable(ctx) {
		t.Error("IsAvailable should return false when service fails")
	}
}

func TestTechnicalAnalyst_GetMetadata(t *testing.T) {
	analyst := &TechnicalAnalyst{}
	metadata := analyst.GetMetadata()

	if metadata.Description == "" {
		t.Error("Description should not be empty")
	}
	if metadata.Version == "" {
		t.Error("Version should not be empty")
	}
	if len(metadata.RequiredServices) == 0 {
		t.Error("RequiredServices should not be empty")
	}

	// Check that required services include both llm and alpaca
	hasAlpaca := false
	hasLLM := false
	for _, svc := range metadata.RequiredServices {
		if svc == "alpaca" {
			hasAlpaca = true
		}
		if svc == "llm" {
			hasLLM = true
		}
	}
	if !hasAlpaca {
		t.Error("RequiredServices should include alpaca")
	}
	if !hasLLM {
		t.Error("RequiredServices should include llm")
	}
}
