package agents

import (
	"context"
	"testing"
	"time"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/repository"

	"github.com/shopspring/decimal"
)

type mockAgent struct {
	name        string
	agentType   models.AgentType
	analysis    *Analysis
	err         error
	isAvailable bool
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

func (m *mockAgent) IsAvailable(ctx context.Context) bool {
	return m.isAvailable
}

func (m *mockAgent) GetMetadata() AgentMetadata {
	return AgentMetadata{
		Description:      "Mock agent for testing",
		Version:          "1.0.0",
		RequiredServices: []string{"mock"},
	}
}

// integrationMockAccountProvider for integration tests
type integrationMockAccountProvider struct{}

func (m *integrationMockAccountProvider) GetAccount(ctx context.Context) (*models.Account, error) {
	return &models.Account{
		ID:             "test-account",
		Currency:       "USD",
		BuyingPower:    decimal.NewFromInt(100000),
		Cash:           decimal.NewFromInt(50000),
		PortfolioValue: decimal.NewFromInt(100000),
		Equity:         decimal.NewFromInt(100000),
	}, nil
}

func (m *integrationMockAccountProvider) GetPosition(ctx context.Context, symbol string) (*models.Position, error) {
	return nil, nil
}

func (m *integrationMockAccountProvider) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return &models.Quote{
		Symbol: symbol,
		Last:   decimal.NewFromInt(150),
		Bid:    decimal.NewFromFloat(149.50),
		Ask:    decimal.NewFromFloat(150.50),
	}, nil
}

func TestPortfolioManager_AnalyzeSymbol_Integration(t *testing.T) {
	ctx := context.Background()
	connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available for integration test")
	}
	defer repo.Close()

	manager := NewPortfolioManager(repo, config.NewTestConfig(), &integrationMockAccountProvider{})

	fundamental := &mockAgent{
		name:        "Mock Fundamental",
		agentType:   models.AgentTypeFundamental,
		isAvailable: true,
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
		name:        "Mock Technical",
		agentType:   models.AgentTypeTechnical,
		isAvailable: true,
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
		name:        "Mock News",
		agentType:   models.AgentTypeNews,
		isAvailable: true,
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

	manager := NewPortfolioManager(repo, config.NewTestConfig(), &integrationMockAccountProvider{})

	fundamental := &mockAgent{
		name:        "Mock Fundamental",
		agentType:   models.AgentTypeFundamental,
		isAvailable: true,
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
		name:        "Failing Agent",
		agentType:   models.AgentTypeTechnical,
		isAvailable: true,
		err:         context.DeadlineExceeded,
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

	manager := NewPortfolioManager(repo, config.NewTestConfig(), &integrationMockAccountProvider{})

	failingAgent := &mockAgent{
		name:        "Failing Agent",
		agentType:   models.AgentTypeFundamental,
		isAvailable: true,
		err:         context.DeadlineExceeded,
	}

	manager.RegisterAgent(failingAgent)

	_, err = manager.AnalyzeSymbol(ctx, "AAPL")
	if err == nil {
		t.Error("expected error when all agents fail")
	}
}

func TestPortfolioManager_AnalyzeSymbol_UnavailableAgents(t *testing.T) {
	ctx := context.Background()
	connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available for integration test")
	}
	defer repo.Close()

	manager := NewPortfolioManager(repo, config.NewTestConfig(), &integrationMockAccountProvider{})

	availableAgent := &mockAgent{
		name:        "Available Agent",
		agentType:   models.AgentTypeFundamental,
		isAvailable: true,
		analysis: &Analysis{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeFundamental,
			Score:      50,
			Confidence: 80,
			Reasoning:  "Strong fundamentals",
			Data:       map[string]interface{}{},
			Timestamp:  time.Now(),
		},
	}

	unavailableAgent := &mockAgent{
		name:        "Unavailable Agent",
		agentType:   models.AgentTypeTechnical,
		isAvailable: false,
		analysis: &Analysis{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeTechnical,
			Score:      100,
			Confidence: 100,
			Reasoning:  "Should not be used",
			Data:       map[string]interface{}{},
			Timestamp:  time.Now(),
		},
	}

	manager.RegisterAgent(availableAgent)
	manager.RegisterAgent(unavailableAgent)

	rec, err := manager.AnalyzeSymbol(ctx, "AAPL")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should only use the available agent
	if rec.FundamentalScore != 50 {
		t.Errorf("expected fundamental score 50, got %f", rec.FundamentalScore)
	}

	// Technical score should be 0 since that agent was unavailable
	if rec.TechnicalScore != 0 {
		t.Errorf("expected technical score 0 (unavailable agent), got %f", rec.TechnicalScore)
	}
}

func TestPortfolioManager_AnalyzeSymbol_AllUnavailable(t *testing.T) {
	ctx := context.Background()
	connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available for integration test")
	}
	defer repo.Close()

	manager := NewPortfolioManager(repo, config.NewTestConfig(), &integrationMockAccountProvider{})

	unavailableAgent := &mockAgent{
		name:        "Unavailable Agent",
		agentType:   models.AgentTypeFundamental,
		isAvailable: false,
	}

	manager.RegisterAgent(unavailableAgent)

	_, err = manager.AnalyzeSymbol(ctx, "AAPL")
	if err == nil {
		t.Error("expected error when all agents are unavailable")
	}
}
