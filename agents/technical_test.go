package agents

import (
	"testing"
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
		name      string
		prices    []float64
		period    int
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "uptrending prices - high RSI",
			prices:    []float64{40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55},
			period:    14,
			wantMin:   70.0, // Should be high (bullish)
			wantMax:   100.0,
		},
		{
			name:      "downtrending prices - low RSI",
			prices:    []float64{55, 54, 53, 52, 51, 50, 49, 48, 47, 46, 45, 44, 43, 42, 41, 40},
			period:    14,
			wantMin:   0.0,
			wantMax:   30.0, // Should be low (bearish)
		},
		{
			name:      "insufficient data returns neutral",
			prices:    []float64{100, 101, 102},
			period:    14,
			wantMin:   50.0,
			wantMax:   50.0,
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
