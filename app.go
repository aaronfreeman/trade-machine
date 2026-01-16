package main

import (
	"context"
	"fmt"

	"trade-machine/agents"
	"trade-machine/models"
	"trade-machine/repository"
	"trade-machine/services"
)

// App struct
type App struct {
	ctx              context.Context
	repo             *repository.Repository
	portfolioManager *agents.PortfolioManager
	alpacaService    *services.AlpacaService
}

// NewApp creates a new App application struct
func NewApp(repo *repository.Repository, manager *agents.PortfolioManager, alpaca *services.AlpacaService) *App {
	return &App{
		repo:             repo,
		portfolioManager: manager,
		alpacaService:    alpaca,
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

// Helper function to parse UUID
func parseUUID(id string) ([16]byte, error) {
	var uuid [16]byte
	if len(id) != 36 {
		return uuid, fmt.Errorf("invalid UUID format")
	}
	// Simple UUID parsing (format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
	hex := id[0:8] + id[9:13] + id[14:18] + id[19:23] + id[24:36]
	for i := 0; i < 16; i++ {
		var b byte
		_, err := fmt.Sscanf(hex[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return uuid, fmt.Errorf("invalid UUID: %w", err)
		}
		uuid[i] = b
	}
	return uuid, nil
}
