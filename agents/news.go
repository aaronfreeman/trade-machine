package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"trade-machine/models"
)

const newsSystemPrompt = `You are a financial analyst specializing in news sentiment analysis.
Your job is to analyze recent news articles about a stock and determine market sentiment.

You will be given a list of recent news headlines and descriptions.

Based on this news, provide your analysis in the following JSON format:
{
  "score": <number from -100 to 100, negative=bearish/negative sentiment, positive=bullish/positive sentiment>,
  "confidence": <number from 0 to 100>,
  "reasoning": "<brief explanation of the overall sentiment>",
  "key_themes": ["<theme1>", "<theme2>", "<theme3>"],
  "notable_articles": ["<headline1>", "<headline2>"]
}

Consider:
- Positive news: earnings beats, product launches, partnerships, analyst upgrades
- Negative news: earnings misses, lawsuits, management changes, analyst downgrades
- Neutral news: routine announcements, industry trends

Be objective and focus on how the news might impact stock price.`

// NewsAnalystResponse is the expected response from Claude
type NewsAnalystResponse struct {
	Score           float64  `json:"score"`
	Confidence      float64  `json:"confidence"`
	Reasoning       string   `json:"reasoning"`
	KeyThemes       []string `json:"key_themes"`
	NotableArticles []string `json:"notable_articles"`
}

// NewsAnalyst analyzes news sentiment
type NewsAnalyst struct {
	llm LLMService
	newsAPI     NewsAPIServiceInterface
	healthCache *HealthCache
}

// NewNewsAnalyst creates a new NewsAnalyst
func NewNewsAnalyst(llm LLMService, newsAPI NewsAPIServiceInterface) *NewsAnalyst {
	return &NewsAnalyst{
		llm:     llm,
		newsAPI:     newsAPI,
		healthCache: NewHealthCache(DefaultHealthCacheTTL),
	}
}

// NewNewsAnalystWithCacheTTL creates a new NewsAnalyst with a custom health cache TTL
func NewNewsAnalystWithCacheTTL(llm LLMService, newsAPI NewsAPIServiceInterface, cacheTTL time.Duration) *NewsAnalyst {
	return &NewsAnalyst{
		llm:     llm,
		newsAPI:     newsAPI,
		healthCache: NewHealthCache(cacheTTL),
	}
}

// Analyze performs news sentiment analysis on a stock
func (a *NewsAnalyst) Analyze(ctx context.Context, symbol string) (*Analysis, error) {
	articles, err := a.newsAPI.GetNews(ctx, symbol, 15)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}

	if len(articles) == 0 {
		return &Analysis{
			Symbol:     symbol,
			AgentType:  models.AgentTypeNews,
			Score:      0,
			Confidence: 20,
			Reasoning:  "No recent news found for this symbol",
			Data:       map[string]interface{}{"articles_count": 0},
			Timestamp:  time.Now(),
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Analyze the following recent news about %s:\n\n", symbol))

	for i, article := range articles {
		sb.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, article.Title))
		if article.Description != "" {
			sb.WriteString(fmt.Sprintf("   %s\n", article.Description))
		}
		sb.WriteString(fmt.Sprintf("   Source: %s | Published: %s\n\n",
			article.Source, article.PublishedAt.Format("Jan 2, 2006")))
	}

	sb.WriteString("Provide your sentiment analysis.")

	response, err := a.llm.InvokeWithPrompt(ctx, newsSystemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("failed to invoke bedrock: %w", err)
	}

	var result NewsAnalystResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return &Analysis{
			Symbol:     symbol,
			AgentType:  models.AgentTypeNews,
			Score:      0,
			Confidence: 50,
			Reasoning:  response,
			Data: map[string]interface{}{
				"raw_response":   response,
				"articles_count": len(articles),
			},
			Timestamp: time.Now(),
		}, nil
	}

	return &Analysis{
		Symbol:     symbol,
		AgentType:  models.AgentTypeNews,
		Score:      NormalizeScore(result.Score),
		Confidence: NormalizeConfidence(result.Confidence),
		Reasoning:  result.Reasoning,
		Data: map[string]interface{}{
			"key_themes":       result.KeyThemes,
			"notable_articles": result.NotableArticles,
			"articles_count":   len(articles),
		},
		Timestamp: time.Now(),
	}, nil
}

// Name returns the agent name
func (a *NewsAnalyst) Name() string {
	return "News Sentiment Analyst"
}

// Type returns the agent type
func (a *NewsAnalyst) Type() models.AgentType {
	return models.AgentTypeNews
}

// IsAvailable checks if the agent's dependencies are healthy.
// Results are cached to reduce API calls during frequent availability checks.
func (a *NewsAnalyst) IsAvailable(ctx context.Context) bool {
	if available, valid := a.healthCache.Get(); valid {
		return available
	}

	_, err := a.newsAPI.GetNews(ctx, "AAPL", 1)
	available := err == nil
	a.healthCache.Set(available)
	return available
}

// InvalidateHealthCache clears the health cache, forcing the next check to make a live call.
func (a *NewsAnalyst) InvalidateHealthCache() {
	a.healthCache.Invalidate()
}

// GetMetadata returns information about this agent's capabilities
func (a *NewsAnalyst) GetMetadata() AgentMetadata {
	return AgentMetadata{
		Description:      "Analyzes recent news articles to determine market sentiment",
		Version:          "1.0.0",
		RequiredServices: []string{"llm", "newsapi"},
	}
}
