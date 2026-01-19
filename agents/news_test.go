package agents

import (
	"context"
	"errors"
	"testing"
	"time"

	"trade-machine/models"
)

func TestNewsAnalyst_Name(t *testing.T) {
	analyst := &NewsAnalyst{}
	if analyst.Name() != "News Sentiment Analyst" {
		t.Errorf("Name() = %v, want 'News Sentiment Analyst'", analyst.Name())
	}
}

func TestNewsAnalyst_Type(t *testing.T) {
	analyst := &NewsAnalyst{}
	if analyst.Type() != models.AgentTypeNews {
		t.Errorf("Type() = %v, want AgentTypeNews", analyst.Type())
	}
}

func TestNewNewsAnalyst(t *testing.T) {
	analyst := NewNewsAnalyst(nil, nil)
	if analyst == nil {
		t.Error("NewNewsAnalyst should not return nil")
	}
}

func TestNewsAnalystResponse_Parsing(t *testing.T) {
	resp := NewsAnalystResponse{
		Score:           45.0,
		Confidence:      70.0,
		Reasoning:       "Mixed sentiment with slight positive bias",
		KeyThemes:       []string{"Earnings beat", "Market volatility", "Product launch"},
		NotableArticles: []string{"AAPL beats Q4 expectations", "Market reacts to Fed news"},
	}

	if resp.Score != 45.0 {
		t.Errorf("Score = %v, want 45.0", resp.Score)
	}
	if resp.Confidence != 70.0 {
		t.Errorf("Confidence = %v, want 70.0", resp.Confidence)
	}
	if len(resp.KeyThemes) != 3 {
		t.Errorf("KeyThemes length = %v, want 3", len(resp.KeyThemes))
	}
	if len(resp.NotableArticles) != 2 {
		t.Errorf("NotableArticles length = %v, want 2", len(resp.NotableArticles))
	}
}

func TestNewsAnalyst_Analyze_Success(t *testing.T) {
	mockLLM := &mockLLMService{
		response: `{
			"score": 70.0,
			"confidence": 80.0,
			"reasoning": "Predominantly positive news coverage with analyst upgrades",
			"key_themes": ["Earnings beat", "Product launch", "Analyst upgrade"],
			"notable_articles": ["Apple Q4 earnings beat expectations", "New product announcement"]
		}`,
	}

	mockNewsAPI := &mockNewsAPIService{
		articles: []models.NewsArticle{
			{
				Title:       "Apple Q4 earnings beat expectations",
				Description: "Apple reported strong quarterly results",
				URL:         "https://example.com/article1",
				Source:      "TechNews",
				Author:      "John Doe",
				PublishedAt: time.Now().Add(-24 * time.Hour),
			},
			{
				Title:       "Apple announces new product line",
				Description: "New product expected to drive growth",
				URL:         "https://example.com/article2",
				Source:      "BusinessInsider",
				Author:      "Jane Smith",
				PublishedAt: time.Now().Add(-48 * time.Hour),
			},
		},
	}

	analyst := NewNewsAnalyst(mockLLM, mockNewsAPI)
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "AAPL")
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if analysis.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want AAPL", analysis.Symbol)
	}
	if analysis.AgentType != models.AgentTypeNews {
		t.Errorf("AgentType = %v, want News", analysis.AgentType)
	}
	if analysis.Score != 70.0 {
		t.Errorf("Score = %v, want 70.0", analysis.Score)
	}
	if analysis.Confidence != 80.0 {
		t.Errorf("Confidence = %v, want 80.0", analysis.Confidence)
	}

	keyThemes, ok := analysis.Data["key_themes"].([]string)
	if !ok {
		t.Error("key_themes should be []string")
	}
	if len(keyThemes) != 3 {
		t.Errorf("Expected 3 key themes, got %d", len(keyThemes))
	}
}

func TestNewsAnalyst_Analyze_NoArticles(t *testing.T) {
	mockLLM := &mockLLMService{
		response: `{"score": 0, "confidence": 20, "reasoning": "No news", "key_themes": [], "notable_articles": []}`,
	}

	mockNewsAPI := &mockNewsAPIService{
		articles: []models.NewsArticle{},
	}

	analyst := NewNewsAnalyst(mockLLM, mockNewsAPI)
	ctx := context.Background()

	analysis, err := analyst.Analyze(ctx, "UNKNOWN")
	if err != nil {
		t.Fatalf("Analyze should not fail with no articles: %v", err)
	}

	if analysis.Score != 0 {
		t.Errorf("Score = %v, want 0 for no articles", analysis.Score)
	}
	if analysis.Confidence != 20 {
		t.Errorf("Confidence = %v, want 20 for no articles", analysis.Confidence)
	}
	if analysis.Reasoning != "No recent news found for this symbol" {
		t.Errorf("Reasoning should indicate no news found")
	}
}

func TestNewsAnalyst_Analyze_NewsAPIError(t *testing.T) {
	mockLLM := &mockLLMService{
		response: `{"score": 0, "confidence": 50, "reasoning": "test", "key_themes": [], "notable_articles": []}`,
	}

	mockNewsAPI := &mockNewsAPIService{
		err: errors.New("API rate limit exceeded"),
	}

	analyst := NewNewsAnalyst(mockLLM, mockNewsAPI)
	ctx := context.Background()

	_, err := analyst.Analyze(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when NewsAPI fails")
	}
}

func TestNewsAnalyst_Analyze_LLMError(t *testing.T) {
	mockLLM := &mockLLMService{
		err: errors.New("LLM service unavailable"),
	}

	mockNewsAPI := &mockNewsAPIService{
		articles: []models.NewsArticle{
			{Title: "Test", Description: "Test", Source: "Test"},
		},
	}

	analyst := NewNewsAnalyst(mockLLM, mockNewsAPI)
	ctx := context.Background()

	_, err := analyst.Analyze(ctx, "AAPL")
	if err == nil {
		t.Error("Expected error when LLM fails")
	}
}

func TestNewsAnalyst_Analyze_InvalidJSON(t *testing.T) {
	mockLLM := &mockLLMService{
		response: "This is plain text analysis, not JSON",
	}

	mockNewsAPI := &mockNewsAPIService{
		articles: []models.NewsArticle{
			{Title: "Test Article", Description: "Test", Source: "Test"},
		},
	}

	analyst := NewNewsAnalyst(mockLLM, mockNewsAPI)
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

func TestNewsAnalyst_IsAvailable_Success(t *testing.T) {
	mockNewsAPI := &mockNewsAPIService{
		articles: []models.NewsArticle{
			{Title: "Test", Description: "Test", Source: "Test"},
		},
	}

	analyst := NewNewsAnalyst(nil, mockNewsAPI)
	ctx := context.Background()

	if !analyst.IsAvailable(ctx) {
		t.Error("IsAvailable should return true when service is healthy")
	}
}

func TestNewsAnalyst_IsAvailable_Failure(t *testing.T) {
	mockNewsAPI := &mockNewsAPIService{
		err: errors.New("service unavailable"),
	}

	analyst := NewNewsAnalyst(nil, mockNewsAPI)
	ctx := context.Background()

	if analyst.IsAvailable(ctx) {
		t.Error("IsAvailable should return false when service fails")
	}
}

func TestNewsAnalyst_GetMetadata(t *testing.T) {
	analyst := &NewsAnalyst{}
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

	// Check that required services include both llm and newsapi
	hasNewsAPI := false
	hasLLM := false
	for _, svc := range metadata.RequiredServices {
		if svc == "newsapi" {
			hasNewsAPI = true
		}
		if svc == "llm" {
			hasLLM = true
		}
	}
	if !hasNewsAPI {
		t.Error("RequiredServices should include newsapi")
	}
	if !hasLLM {
		t.Error("RequiredServices should include llm")
	}
}
