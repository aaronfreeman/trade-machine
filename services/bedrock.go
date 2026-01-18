package services

import (
	"context"
	"encoding/json"
	"fmt"

	appconfig "trade-machine/config"
	"trade-machine/observability"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// bedrockClient defines the interface for Bedrock API calls (for testing)
type bedrockClient interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

// BedrockService handles communication with AWS Bedrock for Claude models
type BedrockService struct {
	client           bedrockClient
	model            string
	maxTokens        int
	anthropicVersion string
}

// ClaudeRequest represents the request format for Claude models via Bedrock
type ClaudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	System           string          `json:"system,omitempty"`
	Messages         []ClaudeMessage `json:"messages"`
}

// ClaudeMessage represents a message in the Claude conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude models
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewBedrockService creates a new BedrockService instance
func NewBedrockService(ctx context.Context, cfg *appconfig.Config) (*BedrockService, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &BedrockService{
		client:           bedrockruntime.NewFromConfig(awsCfg),
		model:            cfg.AWS.BedrockModelID,
		maxTokens:        cfg.AWS.BedrockMaxTokens,
		anthropicVersion: cfg.AWS.AnthropicVersion,
	}, nil
}

// InvokeWithPrompt sends a prompt to Claude and returns the response text
func (s *BedrockService) InvokeWithPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	metrics := observability.GetMetrics()
	metrics.RecordExternalAPIRequest(BreakerBedrock, "invoke")
	timer := metrics.NewTimer()

	result, err := WithCircuitBreaker(ctx, BreakerBedrock, func() (string, error) {
		request := ClaudeRequest{
			AnthropicVersion: s.anthropicVersion,
			MaxTokens:        s.maxTokens,
			System:           systemPrompt,
			Messages: []ClaudeMessage{
				{Role: "user", Content: userPrompt},
			},
		}

		reqBody, err := json.Marshal(request)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request: %w", err)
		}

		output, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(s.model),
			Body:        reqBody,
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			return "", fmt.Errorf("failed to invoke model: %w", err)
		}

		var response ClaudeResponse
		if err := json.Unmarshal(output.Body, &response); err != nil {
			return "", fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(response.Content) == 0 {
			return "", fmt.Errorf("empty response from model")
		}

		return response.Content[0].Text, nil
	})

	timer.ObserveExternalAPI(BreakerBedrock, "invoke")
	if err != nil {
		metrics.RecordExternalAPIError(BreakerBedrock, "invoke", categorizeAPIError(err))
	}
	return result, err
}

// InvokeStructured sends a prompt and parses the JSON response into the provided struct
func (s *BedrockService) InvokeStructured(ctx context.Context, systemPrompt, userPrompt string, result interface{}) error {
	text, err := s.InvokeWithPrompt(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(text), result); err != nil {
		return fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	return nil
}

// Chat enables multi-turn conversation with Claude
func (s *BedrockService) Chat(ctx context.Context, systemPrompt string, messages []ClaudeMessage) (string, error) {
	metrics := observability.GetMetrics()
	metrics.RecordExternalAPIRequest(BreakerBedrock, "chat")
	timer := metrics.NewTimer()

	result, err := WithCircuitBreaker(ctx, BreakerBedrock, func() (string, error) {
		request := ClaudeRequest{
			AnthropicVersion: s.anthropicVersion,
			MaxTokens:        s.maxTokens,
			System:           systemPrompt,
			Messages:         messages,
		}

		reqBody, err := json.Marshal(request)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request: %w", err)
		}

		output, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(s.model),
			Body:        reqBody,
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			return "", fmt.Errorf("failed to invoke model: %w", err)
		}

		var response ClaudeResponse
		if err := json.Unmarshal(output.Body, &response); err != nil {
			return "", fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(response.Content) == 0 {
			return "", fmt.Errorf("empty response from model")
		}

		return response.Content[0].Text, nil
	})

	timer.ObserveExternalAPI(BreakerBedrock, "chat")
	if err != nil {
		metrics.RecordExternalAPIError(BreakerBedrock, "chat", categorizeAPIError(err))
	}
	return result, err
}

// categorizeAPIError categorizes an API error for metrics labeling
func categorizeAPIError(err error) string {
	if err == nil {
		return "none"
	}
	errStr := err.Error()
	switch {
	case containsStr(errStr, "timeout"), containsStr(errStr, "context deadline"):
		return "timeout"
	case containsStr(errStr, "circuit breaker"):
		return "circuit_breaker"
	case containsStr(errStr, "rate limit"), containsStr(errStr, "throttl"):
		return "rate_limit"
	case containsStr(errStr, "connection"), containsStr(errStr, "network"):
		return "network"
	case containsStr(errStr, "unauthorized"), containsStr(errStr, "forbidden"):
		return "auth"
	default:
		return "other"
	}
}

// containsStr is a simple case-insensitive contains check
func containsStr(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
