package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ValidationResult represents the result of validating an API key
type ValidationResult struct {
	Service  ServiceName `json:"service"`
	Valid    bool        `json:"valid"`
	Message  string      `json:"message"`
	Duration time.Duration `json:"duration_ms"`
}

// Validator validates API key configurations
type Validator struct {
	client *http.Client
}

// NewValidator creates a new API key validator
func NewValidator() *Validator {
	return &Validator{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateAPIKey tests if an API key is valid for the given service
func (v *Validator) ValidateAPIKey(ctx context.Context, config *APIKeyConfig) (*ValidationResult, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	start := time.Now()
	result := &ValidationResult{
		Service: config.ServiceName,
	}

	var err error
	switch config.ServiceName {
	case ServiceOpenAI:
		err = v.validateOpenAI(ctx, config)
	case ServiceAlpaca:
		err = v.validateAlpaca(ctx, config)
	case ServiceAlphaVantage:
		err = v.validateAlphaVantage(ctx, config)
	case ServiceNewsAPI:
		err = v.validateNewsAPI(ctx, config)
	case ServiceFMP:
		err = v.validateFMP(ctx, config)
	default:
		err = fmt.Errorf("unknown service: %s", config.ServiceName)
	}

	result.Duration = time.Since(start)

	if err != nil {
		result.Valid = false
		result.Message = err.Error()
	} else {
		result.Valid = true
		result.Message = "Connection successful"
	}

	return result, nil
}

// validateOpenAI tests OpenAI API connectivity
func (v *Validator) validateOpenAI(ctx context.Context, config *APIKeyConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return errors.New("invalid API key")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// validateAlpaca tests Alpaca API connectivity
func (v *Validator) validateAlpaca(ctx context.Context, config *APIKeyConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required")
	}
	if config.APISecret == "" {
		return errors.New("API secret is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v2/account", nil)
	if err != nil {
		return err
	}
	req.Header.Set("APCA-API-KEY-ID", config.APIKey)
	req.Header.Set("APCA-API-SECRET-KEY", config.APISecret)

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return errors.New("invalid API credentials")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// validateAlphaVantage tests Alpha Vantage API connectivity
func (v *Validator) validateAlphaVantage(ctx context.Context, config *APIKeyConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required")
	}

	// Use a simple function call to test the API
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=IBM&interval=5min&apikey=%s", config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Check response for error message
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	// Alpha Vantage returns error messages in "Error Message" or "Note" fields
	if errMsg, ok := result["Error Message"].(string); ok {
		return fmt.Errorf("API error: %s", errMsg)
	}
	if note, ok := result["Note"].(string); ok {
		// Rate limit message - key is valid but rate limited
		if len(note) > 50 {
			return nil // Key is valid, just rate limited
		}
	}

	return nil
}

// validateNewsAPI tests NewsAPI connectivity
func (v *Validator) validateNewsAPI(ctx context.Context, config *APIKeyConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required")
	}

	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=test&pageSize=1&apiKey=%s", config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return errors.New("invalid API key")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// validateFMP tests Financial Modeling Prep API connectivity
func (v *Validator) validateFMP(ctx context.Context, config *APIKeyConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required")
	}

	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/profile/AAPL?apikey=%s", config.APIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return errors.New("invalid API key")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

