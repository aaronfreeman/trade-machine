package main

import (
	"context"
	"embed"
	"log"
	"os"

	"trade-machine/agents"
	"trade-machine/repository"
	"trade-machine/services"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	// Initialize database
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
	}

	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		log.Println("Running without database connection. Some features will be unavailable.")
		repo = nil
	}

	// Initialize services (with nil checks for graceful degradation)
	var bedrockService *services.BedrockService
	var alpacaService *services.AlpacaService
	var alphaVantageService *services.AlphaVantageService
	var newsAPIService *services.NewsAPIService

	// AWS Bedrock Service
	awsRegion := os.Getenv("AWS_REGION")
	bedrockModelID := os.Getenv("BEDROCK_MODEL_ID")
	if awsRegion != "" && bedrockModelID != "" {
		bedrockService, err = services.NewBedrockService(ctx, awsRegion, bedrockModelID)
		if err != nil {
			log.Printf("Warning: Failed to initialize Bedrock service: %v", err)
		}
	} else {
		log.Println("Warning: AWS_REGION or BEDROCK_MODEL_ID not set, AI agents disabled")
	}

	// Alpaca Service
	alpacaKey := os.Getenv("ALPACA_API_KEY")
	alpacaSecret := os.Getenv("ALPACA_API_SECRET")
	alpacaBaseURL := os.Getenv("ALPACA_BASE_URL")
	if alpacaBaseURL == "" {
		alpacaBaseURL = "https://paper-api.alpaca.markets"
	}
	if alpacaKey != "" && alpacaSecret != "" {
		alpacaService = services.NewAlpacaService(alpacaKey, alpacaSecret, alpacaBaseURL)
	} else {
		log.Println("Warning: Alpaca API credentials not set, trading disabled")
	}

	// Alpha Vantage Service
	alphaVantageKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if alphaVantageKey != "" {
		alphaVantageService = services.NewAlphaVantageService(alphaVantageKey)
	} else {
		log.Println("Warning: Alpha Vantage API key not set, fundamental analysis disabled")
	}

	// NewsAPI Service
	newsAPIKey := os.Getenv("NEWS_API_KEY")
	if newsAPIKey != "" {
		newsAPIService = services.NewNewsAPIService(newsAPIKey)
	} else {
		log.Println("Warning: NewsAPI key not set, news sentiment analysis disabled")
	}

	// Initialize Portfolio Manager and register agents
	var portfolioManager *agents.PortfolioManager
	if repo != nil {
		portfolioManager = agents.NewPortfolioManager(repo)

		// Register agents if their dependencies are available
		if bedrockService != nil && alphaVantageService != nil {
			portfolioManager.RegisterAgent(agents.NewFundamentalAnalyst(bedrockService, alphaVantageService))
		}
		if bedrockService != nil && newsAPIService != nil {
			portfolioManager.RegisterAgent(agents.NewNewsAnalyst(bedrockService, newsAPIService))
		}
		if bedrockService != nil && alpacaService != nil {
			portfolioManager.RegisterAgent(agents.NewTechnicalAnalyst(bedrockService, alpacaService))
		}
	}

	// Initialize app
	app := NewApp(repo, portfolioManager, alpacaService)
	apiHandler := NewAPIHandler(app)

	// Run Wails application
	err = wails.Run(&options.App{
		Title:  "Trade Machine",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: apiHandler,
		},
		BackgroundColour: options.NewRGB(27, 38, 54),
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
