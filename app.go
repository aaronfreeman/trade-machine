package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"trade-machine/agents"
	"trade-machine/models"
	"trade-machine/repository"
	"trade-machine/services"

	"github.com/google/uuid"
)

// App struct
type App struct {
	ctx              context.Context
	repo             *repository.Repository
	portfolioManager *agents.PortfolioManager
	alpacaService    *services.AlpacaService
	analysisSem      chan struct{}
}

// NewApp creates a new App application struct
func NewApp(repo *repository.Repository, manager *agents.PortfolioManager, alpaca *services.AlpacaService) *App {
	concurrencyLimit := 3
	if val := os.Getenv("ANALYSIS_CONCURRENCY_LIMIT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			concurrencyLimit = parsed
		}
	}

	return &App{
		repo:             repo,
		portfolioManager: manager,
		alpacaService:    alpaca,
		analysisSem:      make(chan struct{}, concurrencyLimit),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.repo != nil {
		a.repo.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s! Welcome to Trade Machine.", name)
}

// AnalyzeStock runs all agents to analyze a stock and generate a recommendation
func (a *App) AnalyzeStock(symbol string) (*models.Recommendation, error) {
	if a.portfolioManager == nil {
		return nil, fmt.Errorf("portfolio manager not initialized")
	}

	select {
	case a.analysisSem <- struct{}{}:
		defer func() { <-a.analysisSem }()
	default:
		return nil, fmt.Errorf("analysis queue full, too many concurrent requests - try again later")
	}

	return a.portfolioManager.AnalyzeSymbol(a.ctx, symbol)
}

// GetRecommendations returns recent recommendations
func (a *App) GetRecommendations(limit int) ([]models.Recommendation, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetRecommendations(a.ctx, "", limit)
}

// GetPendingRecommendations returns pending recommendations awaiting approval
func (a *App) GetPendingRecommendations() ([]models.Recommendation, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetPendingRecommendations(a.ctx)
}

// ApproveRecommendation approves a recommendation for execution
func (a *App) ApproveRecommendation(id string) error {
	if a.repo == nil {
		return fmt.Errorf("database not initialized")
	}

	uuid, err := parseUUID(id)
	if err != nil {
		return err
	}

	return a.repo.ApproveRecommendation(a.ctx, uuid)
}

// RejectRecommendation rejects a recommendation
func (a *App) RejectRecommendation(id string) error {
	if a.repo == nil {
		return fmt.Errorf("database not initialized")
	}

	uuid, err := parseUUID(id)
	if err != nil {
		return err
	}

	return a.repo.RejectRecommendation(a.ctx, uuid)
}

// GetPositions returns all current positions
func (a *App) GetPositions() ([]models.Position, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetPositions(a.ctx)
}

// GetTrades returns recent trades
func (a *App) GetTrades(limit int) ([]models.Trade, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetTrades(a.ctx, limit)
}

// GetAgentRuns returns recent agent runs
func (a *App) GetAgentRuns(limit int) ([]models.AgentRun, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetAgentRuns(a.ctx, "", limit)
}

func parseUUID(id string) ([16]byte, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return [16]byte{}, fmt.Errorf("invalid UUID: %w", err)
	}
	return parsed, nil
}
