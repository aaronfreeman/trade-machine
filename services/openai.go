package services

import (
	"context"
	"encoding/json"
	"fmt"

	appconfig "trade-machine/config"
	"trade-machine/observability"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// openaiClient defines the interface for OpenAI API calls (for testing)
type openaiClient interface {
	CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

// openaiClientWrapper wraps the openai.Client to implement our interface
type openaiClientWrapper struct {
	client openai.Client
}

func (w *openaiClientWrapper) CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	return w.client.Chat.Completions.New(ctx, params)
}

// OpenAIService handles communication with OpenAI API
type OpenAIService struct {
	client    openaiClient
	model     string
	maxTokens int
}

// NewOpenAIService creates a new OpenAIService instance
func NewOpenAIService(cfg *appconfig.Config) (*OpenAIService, error) {
	if cfg.OpenAI.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	client := openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))

	return &OpenAIService{
		client:    &openaiClientWrapper{client: client},
		model:     cfg.OpenAI.Model,
		maxTokens: cfg.OpenAI.MaxTokens,
	}, nil
}

// newOpenAIServiceWithClient creates an OpenAIService with a custom client (for testing)
func newOpenAIServiceWithClient(client openaiClient, model string, maxTokens int) *OpenAIService {
	return &OpenAIService{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}
}

// InvokeWithPrompt sends a prompt to OpenAI and returns the response text
func (s *OpenAIService) InvokeWithPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	metrics := observability.GetMetrics()
	metrics.RecordExternalAPIRequest(BreakerOpenAI, "invoke")
	timer := metrics.NewTimer()

	result, err := WithCircuitBreaker(ctx, BreakerOpenAI, func() (string, error) {
		params := openai.ChatCompletionNewParams{
			Model:     shared.ChatModel(s.model),
			MaxTokens: openai.Int(int64(s.maxTokens)),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
		}

		completion, err := s.client.CreateChatCompletion(ctx, params)
		if err != nil {
			return "", fmt.Errorf("failed to invoke OpenAI: %w", err)
		}

		if len(completion.Choices) == 0 {
			return "", fmt.Errorf("empty response from OpenAI")
		}

		return completion.Choices[0].Message.Content, nil
	})

	timer.ObserveExternalAPI(BreakerOpenAI, "invoke")
	if err != nil {
		metrics.RecordExternalAPIError(BreakerOpenAI, "invoke", categorizeAPIError(err))
	}
	return result, err
}

// InvokeStructured sends a prompt and parses the JSON response into the provided struct
func (s *OpenAIService) InvokeStructured(ctx context.Context, systemPrompt, userPrompt string, result interface{}) error {
	text, err := s.InvokeWithPrompt(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(text), result); err != nil {
		return fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	return nil
}

// Chat enables multi-turn conversation with OpenAI
func (s *OpenAIService) Chat(ctx context.Context, systemPrompt string, messages []ChatMessage) (string, error) {
	metrics := observability.GetMetrics()
	metrics.RecordExternalAPIRequest(BreakerOpenAI, "chat")
	timer := metrics.NewTimer()

	result, err := WithCircuitBreaker(ctx, BreakerOpenAI, func() (string, error) {
		openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages)+1)
		openaiMessages = append(openaiMessages, openai.SystemMessage(systemPrompt))

		for _, msg := range messages {
			switch msg.Role {
			case "user":
				openaiMessages = append(openaiMessages, openai.UserMessage(msg.Content))
			case "assistant":
				openaiMessages = append(openaiMessages, openai.AssistantMessage(msg.Content))
			}
		}

		params := openai.ChatCompletionNewParams{
			Model:     shared.ChatModel(s.model),
			MaxTokens: openai.Int(int64(s.maxTokens)),
			Messages:  openaiMessages,
		}

		completion, err := s.client.CreateChatCompletion(ctx, params)
		if err != nil {
			return "", fmt.Errorf("failed to invoke OpenAI: %w", err)
		}

		if len(completion.Choices) == 0 {
			return "", fmt.Errorf("empty response from OpenAI")
		}

		return completion.Choices[0].Message.Content, nil
	})

	timer.ObserveExternalAPI(BreakerOpenAI, "chat")
	if err != nil {
		metrics.RecordExternalAPIError(BreakerOpenAI, "chat", categorizeAPIError(err))
	}
	return result, err
}

// categorizeAPIError categorizes an error for metrics purposes
func categorizeAPIError(err error) string {
	if err == nil {
		return "none"
	}
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout", "deadline"):
		return "timeout"
	case contains(errStr, "rate limit", "429"):
		return "rate_limit"
	case contains(errStr, "unauthorized", "401"):
		return "auth_error"
	case contains(errStr, "connection", "network"):
		return "connection_error"
	default:
		return "unknown"
	}
}

// contains checks if the string contains any of the substrings
func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
