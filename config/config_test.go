package config

import (
	"os"
	"testing"
)

// saveEnv saves current environment variables for restoration
func saveEnv(t *testing.T, keys []string) map[string]string {
	t.Helper()
	saved := make(map[string]string)
	for _, key := range keys {
		saved[key] = os.Getenv(key)
	}
	return saved
}

// restoreEnv restores previously saved environment variables
func restoreEnv(t *testing.T, saved map[string]string) {
	t.Helper()
	for key, val := range saved {
		if val == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, val)
		}
	}
}

// clearEnv clears environment variables
func clearEnv(t *testing.T, keys []string) {
	t.Helper()
	for _, key := range keys {
		os.Unsetenv(key)
	}
}

var allEnvKeys = []string{
	"DATABASE_URL",
	"AWS_REGION",
	"BEDROCK_MODEL_ID",
	"BEDROCK_MAX_TOKENS",
	"BEDROCK_ANTHROPIC_VERSION",
	"ALPACA_API_KEY",
	"ALPACA_API_SECRET",
	"ALPACA_BASE_URL",
	"ALPHA_VANTAGE_API_KEY",
	"NEWS_API_KEY",
	"AGENT_TIMEOUT_SECONDS",
	"ANALYSIS_CONCURRENCY_LIMIT",
	"TECHNICAL_ANALYSIS_LOOKBACK_DAYS",
	"AGENT_WEIGHT_FUNDAMENTAL",
	"AGENT_WEIGHT_NEWS",
	"AGENT_WEIGHT_TECHNICAL",
	"CORS_ALLOWED_ORIGINS",
}

func TestLoad_Defaults(t *testing.T) {
	saved := saveEnv(t, allEnvKeys)
	defer restoreEnv(t, saved)
	clearEnv(t, allEnvKeys)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with defaults failed: %v", err)
	}

	// Check defaults
	if cfg.AWS.BedrockMaxTokens != 4096 {
		t.Errorf("expected BedrockMaxTokens=4096, got %d", cfg.AWS.BedrockMaxTokens)
	}
	if cfg.AWS.AnthropicVersion != "bedrock-2023-05-31" {
		t.Errorf("expected AnthropicVersion='bedrock-2023-05-31', got %s", cfg.AWS.AnthropicVersion)
	}
	if cfg.Alpaca.BaseURL != "https://paper-api.alpaca.markets" {
		t.Errorf("expected Alpaca.BaseURL='https://paper-api.alpaca.markets', got %s", cfg.Alpaca.BaseURL)
	}
	if cfg.Agent.TimeoutSeconds != 30 {
		t.Errorf("expected TimeoutSeconds=30, got %d", cfg.Agent.TimeoutSeconds)
	}
	if cfg.Agent.ConcurrencyLimit != 3 {
		t.Errorf("expected ConcurrencyLimit=3, got %d", cfg.Agent.ConcurrencyLimit)
	}
	if cfg.Agent.TechnicalLookbackDays != 100 {
		t.Errorf("expected TechnicalLookbackDays=100, got %d", cfg.Agent.TechnicalLookbackDays)
	}
	if cfg.Agent.WeightFundamental != 0.4 {
		t.Errorf("expected WeightFundamental=0.4, got %f", cfg.Agent.WeightFundamental)
	}
	if cfg.Agent.WeightNews != 0.3 {
		t.Errorf("expected WeightNews=0.3, got %f", cfg.Agent.WeightNews)
	}
	if cfg.Agent.WeightTechnical != 0.3 {
		t.Errorf("expected WeightTechnical=0.3, got %f", cfg.Agent.WeightTechnical)
	}
	if cfg.HTTP.CORSAllowedOrigins != "*" {
		t.Errorf("expected CORSAllowedOrigins='*', got %s", cfg.HTTP.CORSAllowedOrigins)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	saved := saveEnv(t, allEnvKeys)
	defer restoreEnv(t, saved)
	clearEnv(t, allEnvKeys)

	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("BEDROCK_MODEL_ID", "anthropic.claude-3-sonnet")
	os.Setenv("BEDROCK_MAX_TOKENS", "8192")
	os.Setenv("ALPACA_API_KEY", "test-key")
	os.Setenv("ALPACA_API_SECRET", "test-secret")
	os.Setenv("ALPACA_BASE_URL", "https://api.alpaca.markets")
	os.Setenv("ALPHA_VANTAGE_API_KEY", "av-key")
	os.Setenv("NEWS_API_KEY", "news-key")
	os.Setenv("AGENT_TIMEOUT_SECONDS", "60")
	os.Setenv("ANALYSIS_CONCURRENCY_LIMIT", "5")
	os.Setenv("TECHNICAL_ANALYSIS_LOOKBACK_DAYS", "200")
	os.Setenv("AGENT_WEIGHT_FUNDAMENTAL", "0.5")
	os.Setenv("AGENT_WEIGHT_NEWS", "0.25")
	os.Setenv("AGENT_WEIGHT_TECHNICAL", "0.25")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with custom values failed: %v", err)
	}

	if cfg.Database.URL != "postgres://localhost/test" {
		t.Errorf("expected Database.URL='postgres://localhost/test', got %s", cfg.Database.URL)
	}
	if cfg.AWS.Region != "us-west-2" {
		t.Errorf("expected AWS.Region='us-west-2', got %s", cfg.AWS.Region)
	}
	if cfg.AWS.BedrockModelID != "anthropic.claude-3-sonnet" {
		t.Errorf("expected AWS.BedrockModelID='anthropic.claude-3-sonnet', got %s", cfg.AWS.BedrockModelID)
	}
	if cfg.AWS.BedrockMaxTokens != 8192 {
		t.Errorf("expected BedrockMaxTokens=8192, got %d", cfg.AWS.BedrockMaxTokens)
	}
	if cfg.Alpaca.APIKey != "test-key" {
		t.Errorf("expected Alpaca.APIKey='test-key', got %s", cfg.Alpaca.APIKey)
	}
	if cfg.Alpaca.BaseURL != "https://api.alpaca.markets" {
		t.Errorf("expected Alpaca.BaseURL='https://api.alpaca.markets', got %s", cfg.Alpaca.BaseURL)
	}
	if cfg.Agent.TimeoutSeconds != 60 {
		t.Errorf("expected TimeoutSeconds=60, got %d", cfg.Agent.TimeoutSeconds)
	}
	if cfg.Agent.ConcurrencyLimit != 5 {
		t.Errorf("expected ConcurrencyLimit=5, got %d", cfg.Agent.ConcurrencyLimit)
	}
	if cfg.Agent.WeightFundamental != 0.5 {
		t.Errorf("expected WeightFundamental=0.5, got %f", cfg.Agent.WeightFundamental)
	}
	if cfg.HTTP.CORSAllowedOrigins != "http://localhost:3000" {
		t.Errorf("expected CORSAllowedOrigins='http://localhost:3000', got %s", cfg.HTTP.CORSAllowedOrigins)
	}
}

func TestValidate_WeightsSumTo1(t *testing.T) {
	saved := saveEnv(t, allEnvKeys)
	defer restoreEnv(t, saved)
	clearEnv(t, allEnvKeys)

	// Weights that don't sum to 1.0
	os.Setenv("AGENT_WEIGHT_FUNDAMENTAL", "0.5")
	os.Setenv("AGENT_WEIGHT_NEWS", "0.3")
	os.Setenv("AGENT_WEIGHT_TECHNICAL", "0.3") // Total = 1.1

	_, err := Load()
	if err == nil {
		t.Error("expected error for weights not summing to 1.0")
	}
}

func TestValidate_WeightRange(t *testing.T) {
	saved := saveEnv(t, allEnvKeys)
	defer restoreEnv(t, saved)
	clearEnv(t, allEnvKeys)

	// Weight out of range (negative) - but since getEnvFloat requires >= 0,
	// it will use the default instead. Let's test direct validation.
	cfg := &Config{
		Agent: AgentConfig{
			WeightFundamental:     -0.1,
			WeightNews:            0.5,
			WeightTechnical:       0.6,
			TimeoutSeconds:        30,
			ConcurrencyLimit:      3,
			TechnicalLookbackDays: 100,
		},
		AWS: AWSConfig{
			BedrockMaxTokens: 4096,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for negative weight")
	}
}

func TestValidate_PositiveIntegers(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		envVal  string
		wantErr bool
	}{
		{
			name:    "negative timeout uses default",
			envKey:  "AGENT_TIMEOUT_SECONDS",
			envVal:  "-5",
			wantErr: false, // uses default
		},
		{
			name:    "zero concurrency uses default",
			envKey:  "ANALYSIS_CONCURRENCY_LIMIT",
			envVal:  "0",
			wantErr: false, // uses default
		},
		{
			name:    "invalid number uses default",
			envKey:  "BEDROCK_MAX_TOKENS",
			envVal:  "not-a-number",
			wantErr: false, // uses default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := saveEnv(t, allEnvKeys)
			defer restoreEnv(t, saved)
			clearEnv(t, allEnvKeys)

			os.Setenv(tt.envKey, tt.envVal)

			_, err := Load()
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestHasDatabase(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{URL: ""},
	}
	if cfg.HasDatabase() {
		t.Error("expected HasDatabase() to return false for empty URL")
	}

	cfg.Database.URL = "postgres://localhost/test"
	if !cfg.HasDatabase() {
		t.Error("expected HasDatabase() to return true for non-empty URL")
	}
}

func TestHasBedrock(t *testing.T) {
	cfg := &Config{
		AWS: AWSConfig{Region: "", BedrockModelID: ""},
	}
	if cfg.HasBedrock() {
		t.Error("expected HasBedrock() to return false for empty config")
	}

	cfg.AWS.Region = "us-west-2"
	if cfg.HasBedrock() {
		t.Error("expected HasBedrock() to return false without model ID")
	}

	cfg.AWS.BedrockModelID = "anthropic.claude-3-sonnet"
	if !cfg.HasBedrock() {
		t.Error("expected HasBedrock() to return true for complete config")
	}
}

func TestHasAlpaca(t *testing.T) {
	cfg := &Config{
		Alpaca: AlpacaConfig{APIKey: "", APISecret: ""},
	}
	if cfg.HasAlpaca() {
		t.Error("expected HasAlpaca() to return false for empty config")
	}

	cfg.Alpaca.APIKey = "key"
	if cfg.HasAlpaca() {
		t.Error("expected HasAlpaca() to return false without secret")
	}

	cfg.Alpaca.APISecret = "secret"
	if !cfg.HasAlpaca() {
		t.Error("expected HasAlpaca() to return true for complete config")
	}
}

func TestHasAlphaVantage(t *testing.T) {
	cfg := &Config{
		AlphaVantage: AlphaVantageConfig{APIKey: ""},
	}
	if cfg.HasAlphaVantage() {
		t.Error("expected HasAlphaVantage() to return false for empty key")
	}

	cfg.AlphaVantage.APIKey = "key"
	if !cfg.HasAlphaVantage() {
		t.Error("expected HasAlphaVantage() to return true for non-empty key")
	}
}

func TestHasNewsAPI(t *testing.T) {
	cfg := &Config{
		NewsAPI: NewsAPIConfig{APIKey: ""},
	}
	if cfg.HasNewsAPI() {
		t.Error("expected HasNewsAPI() to return false for empty key")
	}

	cfg.NewsAPI.APIKey = "key"
	if !cfg.HasNewsAPI() {
		t.Error("expected HasNewsAPI() to return true for non-empty key")
	}
}

func TestGetEnvString(t *testing.T) {
	key := "TEST_GET_ENV_STRING"
	defer os.Unsetenv(key)

	// Empty returns default
	os.Unsetenv(key)
	if got := getEnvString(key, "default"); got != "default" {
		t.Errorf("expected 'default', got %s", got)
	}

	// Set value returns value
	os.Setenv(key, "custom")
	if got := getEnvString(key, "default"); got != "custom" {
		t.Errorf("expected 'custom', got %s", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	key := "TEST_GET_ENV_INT"
	defer os.Unsetenv(key)

	// Empty returns default
	os.Unsetenv(key)
	if got := getEnvInt(key, 42); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}

	// Valid integer
	os.Setenv(key, "100")
	if got := getEnvInt(key, 42); got != 100 {
		t.Errorf("expected 100, got %d", got)
	}

	// Invalid integer returns default
	os.Setenv(key, "invalid")
	if got := getEnvInt(key, 42); got != 42 {
		t.Errorf("expected 42 for invalid value, got %d", got)
	}

	// Negative returns default
	os.Setenv(key, "-5")
	if got := getEnvInt(key, 42); got != 42 {
		t.Errorf("expected 42 for negative value, got %d", got)
	}

	// Zero returns default
	os.Setenv(key, "0")
	if got := getEnvInt(key, 42); got != 42 {
		t.Errorf("expected 42 for zero value, got %d", got)
	}
}

func TestGetEnvFloat(t *testing.T) {
	key := "TEST_GET_ENV_FLOAT"
	defer os.Unsetenv(key)

	// Empty returns default
	os.Unsetenv(key)
	if got := getEnvFloat(key, 0.5); got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}

	// Valid float
	os.Setenv(key, "0.75")
	if got := getEnvFloat(key, 0.5); got != 0.75 {
		t.Errorf("expected 0.75, got %f", got)
	}

	// Invalid float returns default
	os.Setenv(key, "invalid")
	if got := getEnvFloat(key, 0.5); got != 0.5 {
		t.Errorf("expected 0.5 for invalid value, got %f", got)
	}

	// Out of range (> 1) returns default
	os.Setenv(key, "1.5")
	if got := getEnvFloat(key, 0.5); got != 0.5 {
		t.Errorf("expected 0.5 for value > 1, got %f", got)
	}

	// Negative returns default
	os.Setenv(key, "-0.1")
	if got := getEnvFloat(key, 0.5); got != 0.5 {
		t.Errorf("expected 0.5 for negative value, got %f", got)
	}
}
