package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"trade-machine/config"
	"trade-machine/models"

	marketdata "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

const technicalSystemPrompt = `You are a financial analyst specializing in technical analysis.
Your job is to analyze price action and technical indicators to predict short-term price movements.

You will be given technical indicators including:
- RSI (Relative Strength Index): <30 oversold, >70 overbought
- MACD (Moving Average Convergence Divergence) and Signal line
- SMA (Simple Moving Averages): 20-day and 50-day
- Recent price action

Based on these indicators, provide your analysis in the following JSON format:
{
  "score": <number from -100 to 100, negative=bearish, positive=bullish>,
  "confidence": <number from 0 to 100>,
  "reasoning": "<brief explanation of your technical analysis>",
  "signals": ["<signal1>", "<signal2>", "<signal3>"]
}

Consider:
- RSI divergences and overbought/oversold conditions
- MACD crossovers and histogram trends
- Price relative to moving averages (support/resistance)
- Overall trend direction

Be objective and focus on actionable technical signals.`

// TechnicalAnalystResponse is the expected response from Claude
type TechnicalAnalystResponse struct {
	Score      float64  `json:"score"`
	Confidence float64  `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
	Signals    []string `json:"signals"`
}

// TechnicalAnalyst analyzes price action and technical indicators
type TechnicalAnalyst struct {
	bedrock      BedrockServiceInterface
	alpaca       AlpacaServiceInterface
	lookbackDays int
}

// NewTechnicalAnalyst creates a new TechnicalAnalyst
func NewTechnicalAnalyst(bedrock BedrockServiceInterface, alpaca AlpacaServiceInterface, cfg *config.Config) *TechnicalAnalyst {
	return &TechnicalAnalyst{
		bedrock:      bedrock,
		alpaca:       alpaca,
		lookbackDays: cfg.Agent.TechnicalLookbackDays,
	}
}

// Analyze performs technical analysis on a stock
func (a *TechnicalAnalyst) Analyze(ctx context.Context, symbol string) (*Analysis, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -a.lookbackDays)

	bars, err := a.alpaca.GetBars(ctx, symbol, start, end, marketdata.OneDay)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price data: %w", err)
	}

	if len(bars) < 50 {
		return &Analysis{
			Symbol:     symbol,
			AgentType:  models.AgentTypeTechnical,
			Score:      0,
			Confidence: 20,
			Reasoning:  "Insufficient price history for technical analysis",
			Data:       map[string]interface{}{"bars_count": len(bars)},
			Timestamp:  time.Now(),
		}, nil
	}

	// Extract close prices for indicator calculation
	closePrices := make([]float64, len(bars))
	for i, bar := range bars {
		closePrices[i] = bar.Close
	}

	// Calculate indicators
	indicators := a.calculateIndicators(closePrices)

	// Build user prompt with technical data
	latestBar := bars[len(bars)-1]
	userPrompt := fmt.Sprintf(`Analyze the following technical indicators for %s:

Current Price: $%.2f
52-Day High: $%.2f
52-Day Low: $%.2f

RSI (14-period): %.2f
MACD: %.4f
MACD Signal: %.4f
MACD Histogram: %.4f

SMA 20: $%.2f
SMA 50: $%.2f

Price vs SMA20: %.2f%%
Price vs SMA50: %.2f%%

Provide your technical analysis.`,
		symbol,
		latestBar.Close,
		indicators["high"].(float64),
		indicators["low"].(float64),
		indicators["rsi"].(float64),
		indicators["macd"].(float64),
		indicators["macd_signal"].(float64),
		indicators["macd_histogram"].(float64),
		indicators["sma20"].(float64),
		indicators["sma50"].(float64),
		(latestBar.Close/indicators["sma20"].(float64)-1)*100,
		(latestBar.Close/indicators["sma50"].(float64)-1)*100,
	)

	// Call Claude via Bedrock
	response, err := a.bedrock.InvokeWithPrompt(ctx, technicalSystemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke bedrock: %w", err)
	}

	// Parse response
	var result TechnicalAnalystResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return &Analysis{
			Symbol:     symbol,
			AgentType:  models.AgentTypeTechnical,
			Score:      0,
			Confidence: 50,
			Reasoning:  response,
			Data: map[string]interface{}{
				"raw_response": response,
				"indicators":   indicators,
			},
			Timestamp: time.Now(),
		}, nil
	}

	return &Analysis{
		Symbol:     symbol,
		AgentType:  models.AgentTypeTechnical,
		Score:      NormalizeScore(result.Score),
		Confidence: NormalizeConfidence(result.Confidence),
		Reasoning:  result.Reasoning,
		Data: map[string]interface{}{
			"signals":    result.Signals,
			"indicators": indicators,
		},
		Timestamp: time.Now(),
	}, nil
}

// calculateIndicators computes technical indicators from price data
func (a *TechnicalAnalyst) calculateIndicators(prices []float64) map[string]interface{} {
	result := make(map[string]interface{})

	// Calculate RSI manually (14-period)
	result["rsi"] = a.calculateRSI(prices, 14)

	// Calculate SMA 20 and 50
	sma20 := a.calculateSMA(prices, 20)
	sma50 := a.calculateSMA(prices, 50)
	result["sma20"] = sma20
	result["sma50"] = sma50

	// Calculate EMA for MACD
	ema12 := a.calculateEMA(prices, 12)
	ema26 := a.calculateEMA(prices, 26)

	// MACD = EMA12 - EMA26
	macdLine := make([]float64, len(prices))
	for i := range prices {
		macdLine[i] = ema12[i] - ema26[i]
	}

	// Signal line = 9-period EMA of MACD
	signalLine := a.calculateEMA(macdLine, 9)

	if len(macdLine) > 0 && len(signalLine) > 0 {
		macd := macdLine[len(macdLine)-1]
		signal := signalLine[len(signalLine)-1]
		result["macd"] = macd
		result["macd_signal"] = signal
		result["macd_histogram"] = macd - signal
	} else {
		result["macd"] = 0.0
		result["macd_signal"] = 0.0
		result["macd_histogram"] = 0.0
	}

	// Calculate high/low
	high, low := prices[0], prices[0]
	for _, p := range prices {
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
	}
	result["high"] = high
	result["low"] = low

	return result
}

// calculateRSI computes Relative Strength Index
func (a *TechnicalAnalyst) calculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0 // neutral
	}

	var gains, losses float64
	for i := 1; i <= period; i++ {
		change := prices[len(prices)-i] - prices[len(prices)-i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

// calculateSMA computes Simple Moving Average
func (a *TechnicalAnalyst) calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

// calculateEMA computes Exponential Moving Average
func (a *TechnicalAnalyst) calculateEMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return prices
	}

	ema := make([]float64, len(prices))
	multiplier := 2.0 / float64(period+1)

	// Start with SMA for first EMA value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
		ema[i] = prices[i] // placeholder
	}
	ema[period-1] = sum / float64(period)

	// Calculate EMA for rest
	for i := period; i < len(prices); i++ {
		ema[i] = (prices[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

// Name returns the agent name
func (a *TechnicalAnalyst) Name() string {
	return "Technical Analyst"
}

// Type returns the agent type
func (a *TechnicalAnalyst) Type() models.AgentType {
	return models.AgentTypeTechnical
}
