package main

import (
	"context"
	"os"

	"trade-machine/agents"
	"trade-machine/config"
	"trade-machine/internal/api"
	"trade-machine/internal/app"
	"trade-machine/internal/settings"
	"trade-machine/observability"
	"trade-machine/repository"
	"trade-machine/screener"
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

	// Initialize Prometheus metrics
	observability.InitMetrics()

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
	var llmService services.LLMService
	var alpacaService *services.AlpacaService
	var alphaVantageService *services.AlphaVantageService
	var newsAPIService *services.NewsAPIService
	var fmpService *services.FMPService

	// Initialize OpenAI Service
	if cfg.HasOpenAI() {
		openaiService, err := services.NewOpenAIService(cfg)
		if err != nil {
			observability.Warn("failed to initialize OpenAI service", "error", err)
		} else {
			llmService = openaiService
			observability.Info("initialized OpenAI service", "model", cfg.OpenAI.Model)
		}
	}

	if llmService == nil {
		observability.Warn("no LLM service configured, AI agents disabled - set OPENAI_API_KEY")
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

	// FMP Service (Financial Modeling Prep for stock screening)
	if cfg.HasFMP() {
		fmpService = services.NewFMPService(cfg.FMP.APIKey)
	} else {
		observability.Warn("FMP_API_KEY not set, stock screener disabled")
	}

	// Initialize Portfolio Manager and register agents
	var portfolioManager *agents.PortfolioManager
	if repo != nil && alpacaService != nil {
		portfolioManager = agents.NewPortfolioManager(repo, cfg, alpacaService)

		// Register agents if their dependencies are available
		if llmService != nil && alphaVantageService != nil {
			portfolioManager.RegisterAgent(agents.NewFundamentalAnalyst(llmService, alphaVantageService))
		}
		if llmService != nil && newsAPIService != nil {
			portfolioManager.RegisterAgent(agents.NewNewsAnalyst(llmService, newsAPIService))
		}
		if llmService != nil {
			portfolioManager.RegisterAgent(agents.NewTechnicalAnalyst(llmService, alpacaService, cfg))
		}
	} else if repo != nil {
		observability.Warn("Alpaca service required for position sizing, portfolio manager disabled")
	}

	// Initialize app
	application := app.New(cfg, repo, portfolioManager, alpacaService)

	// Initialize Settings Store
	settingsPassphrase := os.Getenv("SETTINGS_PASSPHRASE")
	settingsDir := os.Getenv("SETTINGS_DIR")
	settingsStore, err := settings.NewStore(settingsDir, settingsPassphrase)
	if err != nil {
		observability.Warn("failed to initialize settings store", "error", err)
	} else {
		application.SetSettings(settingsStore)
		observability.Info("settings store initialized")
	}

	// Initialize Value Screener if dependencies are available
	if fmpService != nil && portfolioManager != nil && repo != nil {
		valueScreener := screener.NewValueScreener(fmpService, portfolioManager, repo, &cfg.Screener)
		application.SetScreener(valueScreener)
		observability.Info("value screener initialized")
	} else {
		if fmpService == nil {
			observability.Warn("screener disabled: FMP service not available")
		}
		if portfolioManager == nil {
			observability.Warn("screener disabled: portfolio manager not available")
		}
	}

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
