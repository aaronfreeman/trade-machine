package settings

import (
	"context"
	"testing"
)

func TestValidatorValidateAPIKey(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	// Test with nil config
	_, err := validator.ValidateAPIKey(ctx, nil)
	if err == nil {
		t.Error("ValidateAPIKey(nil) should return error")
	}
}

func TestValidatorUnknownService(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	config := &APIKeyConfig{
		ServiceName: ServiceName("unknown"),
		APIKey:      "test",
	}

	result, err := validator.ValidateAPIKey(ctx, config)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}

	if result.Valid {
		t.Error("ValidateAPIKey() unknown service should not be valid")
	}
}

func TestValidatorMissingAPIKey(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		service ServiceName
	}{
		{"OpenAI", ServiceOpenAI},
		{"AlphaVantage", ServiceAlphaVantage},
		{"NewsAPI", ServiceNewsAPI},
		{"FMP", ServiceFMP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &APIKeyConfig{
				ServiceName: tt.service,
				APIKey:      "", // Empty key
			}

			result, err := validator.ValidateAPIKey(ctx, config)
			if err != nil {
				t.Fatalf("ValidateAPIKey() error = %v", err)
			}

			if result.Valid {
				t.Error("ValidateAPIKey() with empty key should not be valid")
			}
			if result.Message == "" {
				t.Error("ValidateAPIKey() should have error message")
			}
		})
	}
}

func TestValidatorAlpacaMissingSecret(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	config := &APIKeyConfig{
		ServiceName: ServiceAlpaca,
		APIKey:      "AKTEST123",
		APISecret:   "", // Empty secret
	}

	result, err := validator.ValidateAPIKey(ctx, config)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}

	if result.Valid {
		t.Error("ValidateAPIKey() Alpaca with empty secret should not be valid")
	}
}

func TestValidatorAWSBedrock(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		config  *APIKeyConfig
		isValid bool
	}{
		{
			name: "missing region",
			config: &APIKeyConfig{
				ServiceName: ServiceAWSBedrock,
				Region:      "",
				ModelID:     "anthropic.claude-3-sonnet",
			},
			isValid: false,
		},
		{
			name: "missing model ID",
			config: &APIKeyConfig{
				ServiceName: ServiceAWSBedrock,
				Region:      "us-east-1",
				ModelID:     "",
			},
			isValid: false,
		},
		{
			name: "valid config",
			config: &APIKeyConfig{
				ServiceName: ServiceAWSBedrock,
				Region:      "us-east-1",
				ModelID:     "anthropic.claude-3-sonnet",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAPIKey(ctx, tt.config)
			if err != nil {
				t.Fatalf("ValidateAPIKey() error = %v", err)
			}

			if result.Valid != tt.isValid {
				t.Errorf("ValidateAPIKey() Valid = %v, want %v", result.Valid, tt.isValid)
			}
		})
	}
}

func TestValidatorResultFields(t *testing.T) {
	validator := NewValidator()
	ctx := context.Background()

	config := &APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "", // Will fail validation
	}

	result, err := validator.ValidateAPIKey(ctx, config)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}

	// Check result fields are populated
	if result.Service != ServiceOpenAI {
		t.Errorf("ValidateAPIKey() Service = %v, want %v", result.Service, ServiceOpenAI)
	}
	if result.Duration == 0 {
		t.Error("ValidateAPIKey() Duration should be > 0")
	}
	if result.Message == "" {
		t.Error("ValidateAPIKey() Message should not be empty")
	}
}

// Note: Actual API connectivity tests are skipped as they require valid API keys
// Those would be integration tests
