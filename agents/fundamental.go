package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"trade-machine/models"
	"trade-machine/services"
)

const fundamentalSystemPrompt = `You are a financial analyst specializing in fundamental analysis. 
Your job is to analyze company fundamentals and provide a recommendation.

You will be given fundamental data for a stock including:
- P/E ratio, EPS, market cap
- 52-week high/low
- Beta (volatility measure)
- Dividend yield

Based on this data, provide your analysis in the following JSON format:
{
  "score": <number from -100 to 100, negative=bearish, positive=bullish>,
  "confidence": <number from 0 to 100>,
  "reasoning": "<brief explanation of your analysis>",
  "key_factors": ["<factor1>", "<factor2>", "<factor3>"]
}

Be objective and data-driven in your analysis.`

// FundamentalAnalystResponse is the expected response from Claude
type FundamentalAnalystResponse struct {
	Score      float64  `json:"score"`
	Confidence float64  `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
	KeyFactors []string `json:"key_factors"`
}

// FundamentalAnalyst analyzes company fundamentals
type FundamentalAnalyst struct {
	bedrock      *services.BedrockService
	alphaVantage *services.AlphaVantageService
}

// NewFundamentalAnalyst creates a new FundamentalAnalyst
func NewFundamentalAnalyst(bedrock *services.BedrockService, alphaVantage *services.AlphaVantageService) *FundamentalAnalyst {
	return &FundamentalAnalyst{
		bedrock:      bedrock,
		alphaVantage: alphaVantage,
	}
}

// Analyze performs fundamental analysis on a stock
func (a *FundamentalAnalyst) Analyze(ctx context.Context, symbol string) (*Analysis, error) {
	// Fetch fundamentals from Alpha Vantage
	fundamentals, err := a.alphaVantage.GetFundamentals(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fundamentals: %w", err)
	}

	// Build user prompt with fundamental data
	userPrompt := fmt.Sprintf(`Analyze the following fundamental data for %s:

P/E Ratio: %.2f
EPS: %s
Market Cap: %s
52-Week High: %s
52-Week Low: %s
Beta: %.2f
Dividend Yield: %.2f%%

Provide your analysis.`,
		symbol,
		fundamentals.PERatio,
		fundamentals.EPS.String(),
		fundamentals.MarketCap.String(),
		fundamentals.Week52High.String(),
		fundamentals.Week52Low.String(),
		fundamentals.Beta,
		fundamentals.DividendYield*100,
	)

	// Call Claude via Bedrock
	response, err := a.bedrock.InvokeWithPrompt(ctx, fundamentalSystemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke bedrock: %w", err)
	}

	// Parse response
	var result FundamentalAnalystResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If parsing fails, return a basic analysis
		return &Analysis{
			Symbol:     symbol,
			AgentType:  models.AgentTypeFundamental,
			Score:      0,
			Confidence: 50,
			Reasoning:  response,
			Data: map[string]interface{}{
				"raw_response": response,
				"fundamentals": fundamentals,
			},
			Timestamp: time.Now(),
		}, nil
	}

	return &Analysis{
		Symbol:     symbol,
		AgentType:  models.AgentTypeFundamental,
		Score:      NormalizeScore(result.Score),
		Confidence: NormalizeConfidence(result.Confidence),
		Reasoning:  result.Reasoning,
		Data: map[string]interface{}{
			"key_factors":  result.KeyFactors,
			"fundamentals": fundamentals,
		},
		Timestamp: time.Now(),
	}, nil
}

// Name returns the agent name
func (a *FundamentalAnalyst) Name() string {
	return "Fundamental Analyst"
}

// Type returns the agent type
func (a *FundamentalAnalyst) Type() models.AgentType {
	return models.AgentTypeFundamental
}
