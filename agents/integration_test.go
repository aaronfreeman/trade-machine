package agents

import (
	"context"
	"testing"
	"time"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/repository"
)

type mockAgent struct {
	name      string
	agentType models.AgentType
	analysis  *Analysis
	err       error
}

func (m *mockAgent) Analyze(ctx context.Context, symbol string) (*Analysis, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.analysis, nil
}

func (m *mockAgent) Name() string {
	return m.name
}

func (m *mockAgent) Type() models.AgentType {
	return m.agentType
}

func TestPortfolioManager_AnalyzeSymbol_Integration(t *testing.T) {
	ctx := context.Background()
	connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available for integration test")
	}
	defer repo.Close()

	manager := NewPortfolioManager(repo, config.NewTestConfig())

	fundamental := &mockAgent{
		name:      "Mock Fundamental",
		agentType: models.AgentTypeFundamental,
		analysis: &Analysis{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeFundamental,
			Score:      60,
			Confidence: 80,
			Reasoning:  "Strong fundamentals",
			Data:       map[string]interface{}{"test": true},
			Timestamp:  time.Now(),
		},
	}

	technical := &mockAgent{
		name:      "Mock Technical",
		agentType: models.AgentTypeTechnical,
		analysis: &Analysis{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeTechnical,
			Score:      40,
			Confidence: 70,
			Reasoning:  "Bullish technical signals",
			Data:       map[string]interface{}{"test": true},
			Timestamp:  time.Now(),
		},
	}

	news := &mockAgent{
		name:      "Mock News",
		agentType: models.AgentTypeNews,
		analysis: &Analysis{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeNews,
			Score:      50,
			Confidence: 75,
			Reasoning:  "Positive news sentiment",
			Data:       map[string]interface{}{"test": true},
			Timestamp:  time.Now(),
		},
	}

	manager.RegisterAgent(fundamental)
	manager.RegisterAgent(technical)
	manager.RegisterAgent(news)

	t.Run("successful analysis", func(t *testing.T) {
		rec, err := manager.AnalyzeSymbol(ctx, "AAPL")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if rec.Symbol != "AAPL" {
			t.Errorf("expected symbol AAPL, got %s", rec.Symbol)
		}

		if rec.Action != models.RecommendationActionBuy {
			t.Errorf("expected BUY action, got %v", rec.Action)
		}

		if rec.FundamentalScore != 60 {
			t.Errorf("expected fundamental score 60, got %f", rec.FundamentalScore)
		}

		if rec.TechnicalScore != 40 {
			t.Errorf("expected technical score 40, got %f", rec.TechnicalScore)
		}

		if rec.SentimentScore != 50 {
			t.Errorf("expected sentiment score 50, got %f", rec.SentimentScore)
		}

		if rec.Status != models.RecommendationStatusPending {
			t.Errorf("expected pending status, got %v", rec.Status)
		}
	})

	t.Run("bearish recommendation", func(t *testing.T) {
		fundamental.analysis.Score = -60
		technical.analysis.Score = -40
		news.analysis.Score = -50

		rec, err := manager.AnalyzeSymbol(ctx, "TSLA")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if rec.Action != models.RecommendationActionSell {
			t.Errorf("expected SELL action, got %v", rec.Action)
		}
	})

	t.Run("neutral recommendation", func(t *testing.T) {
		fundamental.analysis.Score = 10
		technical.analysis.Score = -5
		news.analysis.Score = 0

		rec, err := manager.AnalyzeSymbol(ctx, "MSFT")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if rec.Action != models.RecommendationActionHold {
			t.Errorf("expected HOLD action, got %v", rec.Action)
		}
	})
}

func TestPortfolioManager_AnalyzeSymbol_PartialFailure(t *testing.T) {
	ctx := context.Background()
	connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available for integration test")
	}
	defer repo.Close()

	manager := NewPortfolioManager(repo, config.NewTestConfig())

	fundamental := &mockAgent{
		name:      "Mock Fundamental",
		agentType: models.AgentTypeFundamental,
		analysis: &Analysis{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeFundamental,
			Score:      60,
			Confidence: 80,
			Reasoning:  "Strong fundamentals",
			Data:       map[string]interface{}{},
			Timestamp:  time.Now(),
		},
	}

	failingAgent := &mockAgent{
		name:      "Failing Agent",
		agentType: models.AgentTypeTechnical,
		err:       context.DeadlineExceeded,
	}

	manager.RegisterAgent(fundamental)
	manager.RegisterAgent(failingAgent)

	rec, err := manager.AnalyzeSymbol(ctx, "AAPL")
	if err != nil {
		t.Fatalf("expected no error with partial failure, got: %v", err)
	}

	if rec.Symbol != "AAPL" {
		t.Errorf("expected symbol AAPL, got %s", rec.Symbol)
	}
}

func TestPortfolioManager_AnalyzeSymbol_AllFail(t *testing.T) {
	ctx := context.Background()
	connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available for integration test")
	}
	defer repo.Close()

	manager := NewPortfolioManager(repo, config.NewTestConfig())

	failingAgent := &mockAgent{
		name:      "Failing Agent",
		agentType: models.AgentTypeFundamental,
		err:       context.DeadlineExceeded,
	}

	manager.RegisterAgent(failingAgent)

	_, err = manager.AnalyzeSymbol(ctx, "AAPL")
	if err == nil {
		t.Error("expected error when all agents fail")
	}
}
