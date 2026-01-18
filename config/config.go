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

	// AWS/Bedrock configuration
	AWS AWSConfig

	// External service configurations
	Alpaca       AlpacaConfig
	AlphaVantage AlphaVantageConfig
	NewsAPI      NewsAPIConfig

	// Agent configuration
	Agent AgentConfig

	// Position sizing configuration
	PositionSizing PositionSizingConfig

	// HTTP configuration
	HTTP HTTPConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL string
}

// AWSConfig holds AWS/Bedrock configuration
type AWSConfig struct {
	Region           string
	BedrockModelID   string
	BedrockMaxTokens int
	AnthropicVersion string
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

// AgentConfig holds agent-related configuration
type AgentConfig struct {
	TimeoutSeconds           int
	ConcurrencyLimit         int
	TechnicalLookbackDays    int
	WeightFundamental        float64
	WeightNews               float64
	WeightTechnical          float64
	Strategy                 string  // default, conservative, aggressive, or custom
	BuyThreshold             float64 // for custom strategy
	SellThreshold            float64 // for custom strategy
	MinConfidence            float64 // for custom/conservative strategy
}

// PositionSizingConfig holds position sizing configuration
type PositionSizingConfig struct {
	MaxPositionPercent   float64
	RiskPercent          float64
	MinShares            int64
	MaxShares            int64
	UseConfidenceScaling bool
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
		AWS: AWSConfig{
			Region:           os.Getenv("AWS_REGION"),
			BedrockModelID:   os.Getenv("BEDROCK_MODEL_ID"),
			BedrockMaxTokens: getEnvInt("BEDROCK_MAX_TOKENS", 4096),
			AnthropicVersion: getEnvString("BEDROCK_ANTHROPIC_VERSION", "bedrock-2023-05-31"),
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
		Agent: AgentConfig{
			TimeoutSeconds:           getEnvInt("AGENT_TIMEOUT_SECONDS", 30),
			ConcurrencyLimit:         getEnvInt("ANALYSIS_CONCURRENCY_LIMIT", 3),
			TechnicalLookbackDays:    getEnvInt("TECHNICAL_ANALYSIS_LOOKBACK_DAYS", 100),
			WeightFundamental:        getEnvFloat("AGENT_WEIGHT_FUNDAMENTAL", 0.4),
			WeightNews:               getEnvFloat("AGENT_WEIGHT_NEWS", 0.3),
			WeightTechnical:          getEnvFloat("AGENT_WEIGHT_TECHNICAL", 0.3),
			Strategy:                 getEnvString("AGENT_STRATEGY", "default"),
			BuyThreshold:             getEnvFloatUnbounded("AGENT_BUY_THRESHOLD", 25),
			SellThreshold:            getEnvFloatUnbounded("AGENT_SELL_THRESHOLD", -25),
			MinConfidence:            getEnvFloatUnbounded("AGENT_MIN_CONFIDENCE", 0),
		},
		PositionSizing: PositionSizingConfig{
			MaxPositionPercent:   getEnvFloatRange("POSITION_MAX_PERCENT", 0.10, 0.01, 1.0),
			RiskPercent:          getEnvFloatRange("POSITION_RISK_PERCENT", 0.02, 0.001, 0.1),
			MinShares:            int64(getEnvInt("POSITION_MIN_SHARES", 1)),
			MaxShares:            int64(getEnvInt("POSITION_MAX_SHARES", 0)),
			UseConfidenceScaling: getEnvBool("POSITION_USE_CONFIDENCE_SCALING", true),
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
	if c.AWS.BedrockMaxTokens <= 0 {
		return fmt.Errorf("BEDROCK_MAX_TOKENS must be positive, got %d", c.AWS.BedrockMaxTokens)
	}

	return nil
}

// HasDatabase returns true if database configuration is available
func (c *Config) HasDatabase() bool {
	return c.Database.URL != ""
}

// HasBedrock returns true if Bedrock configuration is available
func (c *Config) HasBedrock() bool {
	return c.AWS.Region != "" && c.AWS.BedrockModelID != ""
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

// getEnvString gets an environment variable with a default value
func getEnvString(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

// getEnvFloat gets an environment variable as a float with a default value (bounded 0-1)
func getEnvFloat(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= 0 && parsed <= 1 {
			return parsed
		}
	}
	return defaultValue
}

// getEnvFloatRange gets an environment variable as a float with min/max bounds
func getEnvFloatRange(key string, defaultValue, minVal, maxVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= minVal && parsed <= maxVal {
			return parsed
		}
	}
	return defaultValue
}

// getEnvFloatUnbounded gets an environment variable as a float with a default value (unbounded)
func getEnvFloatUnbounded(key string, defaultValue float64) float64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as a bool with a default value
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
		AWS: AWSConfig{
			Region:           "us-east-1",
			BedrockModelID:   "anthropic.claude-3-sonnet",
			BedrockMaxTokens: 4096,
			AnthropicVersion: "bedrock-2023-05-31",
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
		},
		PositionSizing: PositionSizingConfig{
			MaxPositionPercent:   0.10,
			RiskPercent:          0.02,
			MinShares:            1,
			MaxShares:            0,
			UseConfidenceScaling: true,
		},
		HTTP: HTTPConfig{
			CORSAllowedOrigins: "*",
		},
	}
}
