package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Database configuration
	Database DatabaseConfig

	// OpenAI configuration
	OpenAI OpenAIConfig

	// External service configurations
	Alpaca       AlpacaConfig
	AlphaVantage AlphaVantageConfig
	NewsAPI      NewsAPIConfig
	FMP          FMPConfig

	// Agent configuration
	Agent AgentConfig

	// Position sizing configuration
	PositionSizing PositionSizingConfig

	// Screener configuration
	Screener ScreenerConfig

	// HTTP configuration
	HTTP HTTPConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL string
}

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// AlpacaConfig holds Alpaca API configuration
type AlpacaConfig struct {
	APIKey    string
	APISecret string
	BaseURL   string
}

// AlphaVantageConfig holds Alpha Vantage API configuration
type AlphaVantageConfig struct {
	APIKey string
}

// NewsAPIConfig holds NewsAPI configuration
type NewsAPIConfig struct {
	APIKey string
}

// FMPConfig holds Financial Modeling Prep API configuration
type FMPConfig struct {
	APIKey string
}

// AgentConfig holds agent-related configuration
type AgentConfig struct {
	TimeoutSeconds        int
	ConcurrencyLimit      int
	TechnicalLookbackDays int
	WeightFundamental     float64
	WeightNews            float64
	WeightTechnical       float64
	Strategy              string  // default, conservative, aggressive, or custom
	BuyThreshold          float64 // for custom strategy
	SellThreshold         float64 // for custom strategy
	MinConfidence         float64 // for custom/conservative strategy
	HealthCacheTTLSeconds int     // TTL for health check caching (default: 30)
}

// PositionSizingConfig holds position sizing configuration
type PositionSizingConfig struct {
	MaxPositionPercent   float64
	RiskPercent          float64
	MinShares            int64
	MaxShares            int64
	UseConfidenceScaling bool
}

// ScreenerConfig holds value screener configuration
type ScreenerConfig struct {
	MarketCapMin       int64   // Minimum market cap filter (default: 1B)
	PERatioMax         float64 // Maximum P/E ratio filter (default: 15)
	PBRatioMax         float64 // Maximum P/B ratio filter (default: 1.5)
	PreFilterLimit     int     // Number of candidates to pre-filter (default: 15)
	TopPicksCount      int     // Number of top picks to return (default: 3)
	AnalysisTimeoutSec int     // Timeout for full analysis in seconds (default: 120)
	MaxConcurrent      int     // Max concurrent analyses (default: 5)
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	CORSAllowedOrigins string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		OpenAI: OpenAIConfig{
			APIKey:    os.Getenv("OPENAI_API_KEY"),
			Model:     getEnvString("OPENAI_MODEL", "gpt-4o"),
			MaxTokens: getEnvInt("OPENAI_MAX_TOKENS", 4096),
		},
		Alpaca: AlpacaConfig{
			APIKey:    os.Getenv("ALPACA_API_KEY"),
			APISecret: os.Getenv("ALPACA_API_SECRET"),
			BaseURL:   getEnvString("ALPACA_BASE_URL", "https://paper-api.alpaca.markets"),
		},
		AlphaVantage: AlphaVantageConfig{
			APIKey: os.Getenv("ALPHA_VANTAGE_API_KEY"),
		},
		NewsAPI: NewsAPIConfig{
			APIKey: os.Getenv("NEWS_API_KEY"),
		},
		FMP: FMPConfig{
			APIKey: os.Getenv("FMP_API_KEY"),
		},
		Agent: AgentConfig{
			TimeoutSeconds:        getEnvInt("AGENT_TIMEOUT_SECONDS", 30),
			ConcurrencyLimit:      getEnvInt("ANALYSIS_CONCURRENCY_LIMIT", 3),
			TechnicalLookbackDays: getEnvInt("TECHNICAL_ANALYSIS_LOOKBACK_DAYS", 100),
			WeightFundamental:     getEnvFloat("AGENT_WEIGHT_FUNDAMENTAL", 0.4),
			WeightNews:            getEnvFloat("AGENT_WEIGHT_NEWS", 0.3),
			WeightTechnical:       getEnvFloat("AGENT_WEIGHT_TECHNICAL", 0.3),
			Strategy:              getEnvString("AGENT_STRATEGY", "default"),
			BuyThreshold:          getEnvFloatUnbounded("AGENT_BUY_THRESHOLD", 25),
			SellThreshold:         getEnvFloatUnbounded("AGENT_SELL_THRESHOLD", -25),
			MinConfidence:         getEnvFloatUnbounded("AGENT_MIN_CONFIDENCE", 0),
			HealthCacheTTLSeconds: getEnvInt("AGENT_HEALTH_CACHE_TTL_SECONDS", 30),
		},
		PositionSizing: PositionSizingConfig{
			MaxPositionPercent:   getEnvFloatRange("POSITION_MAX_PERCENT", 0.10, 0.01, 1.0),
			RiskPercent:          getEnvFloatRange("POSITION_RISK_PERCENT", 0.02, 0.001, 0.1),
			MinShares:            int64(getEnvInt("POSITION_MIN_SHARES", 1)),
			MaxShares:            int64(getEnvInt("POSITION_MAX_SHARES", 0)),
			UseConfidenceScaling: getEnvBool("POSITION_USE_CONFIDENCE_SCALING", true),
		},
		Screener: ScreenerConfig{
			MarketCapMin:       int64(getEnvInt("SCREENER_MARKET_CAP_MIN", 1_000_000_000)),
			PERatioMax:         getEnvFloatUnbounded("SCREENER_PE_RATIO_MAX", 15.0),
			PBRatioMax:         getEnvFloatUnbounded("SCREENER_PB_RATIO_MAX", 1.5),
			PreFilterLimit:     getEnvInt("SCREENER_PREFILTER_LIMIT", 15),
			TopPicksCount:      getEnvInt("SCREENER_TOP_PICKS_COUNT", 3),
			AnalysisTimeoutSec: getEnvInt("SCREENER_ANALYSIS_TIMEOUT_SEC", 120),
			MaxConcurrent:      getEnvInt("SCREENER_MAX_CONCURRENT", 5),
		},
		HTTP: HTTPConfig{
			CORSAllowedOrigins: getEnvString("CORS_ALLOWED_ORIGINS", "*"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate agent weights sum to 1.0
	weightSum := c.Agent.WeightFundamental + c.Agent.WeightNews + c.Agent.WeightTechnical
	if weightSum < 0.99 || weightSum > 1.01 {
		return fmt.Errorf("agent weights must sum to 1.0, got %.2f (fundamental=%.2f, news=%.2f, technical=%.2f)",
			weightSum, c.Agent.WeightFundamental, c.Agent.WeightNews, c.Agent.WeightTechnical)
	}

	// Validate weight ranges
	if c.Agent.WeightFundamental < 0 || c.Agent.WeightFundamental > 1 {
		return fmt.Errorf("AGENT_WEIGHT_FUNDAMENTAL must be between 0 and 1, got %.2f", c.Agent.WeightFundamental)
	}
	if c.Agent.WeightNews < 0 || c.Agent.WeightNews > 1 {
		return fmt.Errorf("AGENT_WEIGHT_NEWS must be between 0 and 1, got %.2f", c.Agent.WeightNews)
	}
	if c.Agent.WeightTechnical < 0 || c.Agent.WeightTechnical > 1 {
		return fmt.Errorf("AGENT_WEIGHT_TECHNICAL must be between 0 and 1, got %.2f", c.Agent.WeightTechnical)
	}

	// Validate positive integers
	if c.Agent.TimeoutSeconds <= 0 {
		return fmt.Errorf("AGENT_TIMEOUT_SECONDS must be positive, got %d", c.Agent.TimeoutSeconds)
	}
	if c.Agent.ConcurrencyLimit <= 0 {
		return fmt.Errorf("ANALYSIS_CONCURRENCY_LIMIT must be positive, got %d", c.Agent.ConcurrencyLimit)
	}
	if c.Agent.TechnicalLookbackDays <= 0 {
		return fmt.Errorf("TECHNICAL_ANALYSIS_LOOKBACK_DAYS must be positive, got %d", c.Agent.TechnicalLookbackDays)
	}

	return nil
}

// HasDatabase returns true if database configuration is available
func (c *Config) HasDatabase() bool {
	return c.Database.URL != ""
}

// HasOpenAI returns true if OpenAI configuration is available
func (c *Config) HasOpenAI() bool {
	return c.OpenAI.APIKey != ""
}

// HasAlpaca returns true if Alpaca configuration is available
func (c *Config) HasAlpaca() bool {
	return c.Alpaca.APIKey != "" && c.Alpaca.APISecret != ""
}

// HasAlphaVantage returns true if Alpha Vantage configuration is available
func (c *Config) HasAlphaVantage() bool {
	return c.AlphaVantage.APIKey != ""
}

// HasNewsAPI returns true if NewsAPI configuration is available
func (c *Config) HasNewsAPI() bool {
	return c.NewsAPI.APIKey != ""
}

// HasFMP returns true if Financial Modeling Prep configuration is available
func (c *Config) HasFMP() bool {
	return c.FMP.APIKey != ""
}

func getEnvString(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= 0 && parsed <= 1 {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloatRange(key string, defaultValue, minVal, maxVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= minVal && parsed <= maxVal {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloatUnbounded(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// NewTestConfig creates a Config with default values for testing
func NewTestConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			URL: "",
		},
		OpenAI: OpenAIConfig{
			APIKey:    "",
			Model:     "gpt-4o",
			MaxTokens: 4096,
		},
		Alpaca: AlpacaConfig{
			APIKey:    "",
			APISecret: "",
			BaseURL:   "https://paper-api.alpaca.markets",
		},
		AlphaVantage: AlphaVantageConfig{
			APIKey: "",
		},
		NewsAPI: NewsAPIConfig{
			APIKey: "",
		},
		FMP: FMPConfig{
			APIKey: "",
		},
		Agent: AgentConfig{
			TimeoutSeconds:        30,
			ConcurrencyLimit:      3,
			TechnicalLookbackDays: 100,
			WeightFundamental:     0.4,
			WeightNews:            0.3,
			WeightTechnical:       0.3,
			Strategy:              "default",
			BuyThreshold:          25,
			SellThreshold:         -25,
			MinConfidence:         0,
			HealthCacheTTLSeconds: 30,
		},
		PositionSizing: PositionSizingConfig{
			MaxPositionPercent:   0.10,
			RiskPercent:          0.02,
			MinShares:            1,
			MaxShares:            0,
			UseConfidenceScaling: true,
		},
		Screener: ScreenerConfig{
			MarketCapMin:       1_000_000_000,
			PERatioMax:         15.0,
			PBRatioMax:         1.5,
			PreFilterLimit:     15,
			TopPicksCount:      3,
			AnalysisTimeoutSec: 120,
			MaxConcurrent:      5,
		},
		HTTP: HTTPConfig{
			CORSAllowedOrigins: "*",
		},
	}
}
