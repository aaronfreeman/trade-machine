package agents

import (
	"context"
	"testing"

	"trade-machine/config"
	"trade-machine/models"

	"github.com/shopspring/decimal"
)

// testConfig returns a test configuration
func testConfig() *config.Config {
	return config.NewTestConfig()
}

// mockAccountProvider implements AccountProvider for testing
type mockAccountProvider struct {
	account  *models.Account
	position *models.Position
	quote    *models.Quote
}

func newMockAccountProvider() *mockAccountProvider {
	return &mockAccountProvider{
		account: &models.Account{
			ID:             "test-account",
			Currency:       "USD",
			BuyingPower:    decimal.NewFromInt(100000),
			Cash:           decimal.NewFromInt(50000),
			PortfolioValue: decimal.NewFromInt(100000),
			Equity:         decimal.NewFromInt(100000),
		},
		quote: &models.Quote{
			Symbol: "TEST",
			Last:   decimal.NewFromInt(100),
			Bid:    decimal.NewFromFloat(99.50),
			Ask:    decimal.NewFromFloat(100.50),
		},
	}
}

func (m *mockAccountProvider) GetAccount(ctx context.Context) (*models.Account, error) {
	return m.account, nil
}

func (m *mockAccountProvider) GetPosition(ctx context.Context, symbol string) (*models.Position, error) {
	return m.position, nil
}

func (m *mockAccountProvider) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	m.quote.Symbol = symbol
	return m.quote, nil
}

func TestPortfolioManager_RegisterAgent(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())

	if len(manager.GetAgents()) != 0 {
		t.Errorf("Initial agents count = %v, want 0", len(manager.GetAgents()))
	}

	// Register a mock agent
	mockAgent := &testMockAgent{name: "Test Agent", agentType: models.AgentTypeFundamental, isAvailable: true}
	manager.RegisterAgent(mockAgent)

	if len(manager.GetAgents()) != 1 {
		t.Errorf("After registration, agents count = %v, want 1", len(manager.GetAgents()))
	}

	// Register another
	mockAgent2 := &testMockAgent{name: "Test Agent 2", agentType: models.AgentTypeNews, isAvailable: true}
	manager.RegisterAgent(mockAgent2)

	if len(manager.GetAgents()) != 2 {
		t.Errorf("After second registration, agents count = %v, want 2", len(manager.GetAgents()))
	}
}

func TestPortfolioManager_Name(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())
	if manager.Name() != "Portfolio Manager" {
		t.Errorf("Name() = %v, want 'Portfolio Manager'", manager.Name())
	}
}

func TestPortfolioManager_Type(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())
	if manager.Type() != models.AgentTypeManager {
		t.Errorf("Type() = %v, want AgentTypeManager", manager.Type())
	}
}

func TestPortfolioManager_SynthesizeRecommendation(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())

	// Test with multiple analyses
	analyses := []*Analysis{
		{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeFundamental,
			Score:      60.0,
			Confidence: 80.0,
			Reasoning:  "Strong fundamentals",
		},
		{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeNews,
			Score:      50.0,
			Confidence: 70.0,
			Reasoning:  "Positive sentiment",
		},
		{
			Symbol:     "AAPL",
			AgentType:  models.AgentTypeTechnical,
			Score:      40.0,
			Confidence: 75.0,
			Reasoning:  "Bullish signals",
		},
	}

	rec := manager.synthesizeRecommendation(context.Background(), "AAPL", analyses)

	if rec.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", rec.Symbol)
	}

	// With all positive scores, action should be buy
	if rec.Action != models.RecommendationActionBuy {
		t.Errorf("Action = %v, want Buy (all scores positive)", rec.Action)
	}

	// Check individual scores are captured
	if rec.FundamentalScore != 60.0 {
		t.Errorf("FundamentalScore = %v, want 60.0", rec.FundamentalScore)
	}
	if rec.SentimentScore != 50.0 {
		t.Errorf("SentimentScore = %v, want 50.0", rec.SentimentScore)
	}
	if rec.TechnicalScore != 40.0 {
		t.Errorf("TechnicalScore = %v, want 40.0", rec.TechnicalScore)
	}

	// Reasoning should mention all agents
	if rec.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}
}

func TestPortfolioManager_SynthesizeRecommendation_Hold(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())

	// Test with mixed scores that should result in hold
	analyses := []*Analysis{
		{
			Symbol:     "MSFT",
			AgentType:  models.AgentTypeFundamental,
			Score:      10.0,
			Confidence: 80.0,
			Reasoning:  "Neutral fundamentals",
		},
		{
			Symbol:     "MSFT",
			AgentType:  models.AgentTypeNews,
			Score:      -5.0,
			Confidence: 70.0,
			Reasoning:  "Slightly negative sentiment",
		},
	}

	rec := manager.synthesizeRecommendation(context.Background(), "MSFT", analyses)

	// With mixed low scores, should be hold
	if rec.Action != models.RecommendationActionHold {
		t.Errorf("Action = %v, want Hold (mixed low scores)", rec.Action)
	}
}

func TestPortfolioManager_SynthesizeRecommendation_Sell(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())

	// Test with negative scores that should result in sell
	analyses := []*Analysis{
		{
			Symbol:     "TSLA",
			AgentType:  models.AgentTypeFundamental,
			Score:      -60.0,
			Confidence: 80.0,
			Reasoning:  "Weak fundamentals",
		},
		{
			Symbol:     "TSLA",
			AgentType:  models.AgentTypeNews,
			Score:      -50.0,
			Confidence: 75.0,
			Reasoning:  "Negative sentiment",
		},
		{
			Symbol:     "TSLA",
			AgentType:  models.AgentTypeTechnical,
			Score:      -40.0,
			Confidence: 70.0,
			Reasoning:  "Bearish signals",
		},
	}

	rec := manager.synthesizeRecommendation(context.Background(), "TSLA", analyses)

	// With all negative scores, action should be sell
	if rec.Action != models.RecommendationActionSell {
		t.Errorf("Action = %v, want Sell (all scores negative)", rec.Action)
	}
}

func TestPortfolioManager_GetAgents(t *testing.T) {
	manager := NewPortfolioManager(nil, testConfig(), newMockAccountProvider())

	agents := manager.GetAgents()
	if agents == nil {
		t.Error("GetAgents should not return nil")
	}
	if len(agents) != 0 {
		t.Errorf("Initial GetAgents length = %v, want 0", len(agents))
	}
}

// Mock agent for testing
type testMockAgent struct {
	name        string
	agentType   models.AgentType
	isAvailable bool
}

func (m *testMockAgent) Analyze(ctx context.Context, symbol string) (*Analysis, error) {
	return &Analysis{
		Symbol:     symbol,
		AgentType:  m.agentType,
		Score:      50.0,
		Confidence: 75.0,
		Reasoning:  "Mock analysis",
	}, nil
}

func (m *testMockAgent) Name() string {
	return m.name
}

func (m *testMockAgent) Type() models.AgentType {
	return m.agentType
}

func (m *testMockAgent) IsAvailable(ctx context.Context) bool {
	return m.isAvailable
}

func (m *testMockAgent) GetMetadata() AgentMetadata {
	return AgentMetadata{
		Description:      "Mock agent for testing",
		Version:          "1.0.0",
		RequiredServices: []string{"mock"},
	}
}
