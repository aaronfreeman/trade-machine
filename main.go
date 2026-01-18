package main

import (
	"context"
	"os"

	"trade-machine/agents"
	"trade-machine/config"
	"trade-machine/internal/api"
	"trade-machine/internal/app"
	"trade-machine/observability"
	"trade-machine/repository"
	"trade-machine/services"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	// Initialize logger (production mode based on environment)
	production := os.Getenv("ENVIRONMENT") == "production"
	observability.InitLogger(production)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		observability.Info("no .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		observability.Fatal("failed to load configuration", "error", err)
	}

	ctx := context.Background()

	// Initialize database
	var repo *repository.Repository
	if cfg.HasDatabase() {
		repo, err = repository.NewRepository(ctx, cfg.Database.URL)
		if err != nil {
			observability.Warn("failed to initialize database",
				"error", err,
				"note", "running without database connection, some features will be unavailable")
			repo = nil
		}
	} else {
		observability.Fatal("DATABASE_URL environment variable is required")
	}

	// Initialize services (with nil checks for graceful degradation)
	var bedrockService *services.BedrockService
	var alpacaService *services.AlpacaService
	var alphaVantageService *services.AlphaVantageService
	var newsAPIService *services.NewsAPIService

	// AWS Bedrock Service
	if cfg.HasBedrock() {
		bedrockService, err = services.NewBedrockService(ctx, cfg)
		if err != nil {
			observability.Warn("failed to initialize Bedrock service", "error", err)
		}
	} else {
		observability.Warn("AWS_REGION or BEDROCK_MODEL_ID not set, AI agents disabled")
	}

	// Alpaca Service
	if cfg.HasAlpaca() {
		alpacaService = services.NewAlpacaService(cfg.Alpaca.APIKey, cfg.Alpaca.APISecret, cfg.Alpaca.BaseURL)
	} else {
		observability.Warn("Alpaca API credentials not set, trading disabled")
	}

	// Alpha Vantage Service
	if cfg.HasAlphaVantage() {
		alphaVantageService = services.NewAlphaVantageService(cfg.AlphaVantage.APIKey)
	} else {
		observability.Warn("Alpha Vantage API key not set, fundamental analysis disabled")
	}

	// NewsAPI Service
	if cfg.HasNewsAPI() {
		newsAPIService = services.NewNewsAPIService(cfg.NewsAPI.APIKey)
	} else {
		observability.Warn("NewsAPI key not set, news sentiment analysis disabled")
	}

	// Initialize Portfolio Manager and register agents
	var portfolioManager *agents.PortfolioManager
	if repo != nil {
		portfolioManager = agents.NewPortfolioManager(repo, cfg)

		// Register agents if their dependencies are available
		if bedrockService != nil && alphaVantageService != nil {
			portfolioManager.RegisterAgent(agents.NewFundamentalAnalyst(bedrockService, alphaVantageService))
		}
		if bedrockService != nil && newsAPIService != nil {
			portfolioManager.RegisterAgent(agents.NewNewsAnalyst(bedrockService, newsAPIService))
		}
		if bedrockService != nil && alpacaService != nil {
			portfolioManager.RegisterAgent(agents.NewTechnicalAnalyst(bedrockService, alpacaService, cfg))
		}
	}

	// Initialize app
	application := app.New(cfg, repo, portfolioManager, alpacaService)
	handler := api.NewHandler(application, cfg)
	router := api.NewRouter(handler, cfg)

	// Run Wails application
	err = wails.Run(&options.App{
		Title:  "Trade Machine",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Handler: router,
		},
		BackgroundColour: options.NewRGB(27, 38, 54),
		OnStartup:        application.Startup,
		OnShutdown:       application.Shutdown,
	})

	if err != nil {
		observability.Fatal("wails application error", "error", err)
	}
}
