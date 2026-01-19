package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"trade-machine/config"

	"github.com/openai/openai-go"
)

// mockOpenAIClient implements openaiClient for testing
type mockOpenAIClient struct {
	completionFunc func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error)
}

func (m *mockOpenAIClient) CreateChatCompletion(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	return m.completionFunc(ctx, params)
}

func newTestOpenAIService(client openaiClient) *OpenAIService {
	return &OpenAIService{
		client:    client,
		model:     "gpt-4o",
		maxTokens: 4096,
	}
}

func TestNewOpenAIService_MissingAPIKey(t *testing.T) {
	cfg := config.NewTestConfig()
	cfg.OpenAI.APIKey = ""

	_, err := NewOpenAIService(cfg)
	if err == nil {
		t.Error("expected error when API key is missing")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewOpenAIService_WithAPIKey(t *testing.T) {
	cfg := config.NewTestConfig()
	cfg.OpenAI.APIKey = "test-api-key"
	cfg.OpenAI.Model = "gpt-4o-mini"
	cfg.OpenAI.MaxTokens = 2048

	service, err := NewOpenAIService(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if service == nil {
		t.Fatal("service should not be nil")
	}
	if service.model != "gpt-4o-mini" {
		t.Errorf("model = %s, want gpt-4o-mini", service.model)
	}
	if service.maxTokens != 2048 {
		t.Errorf("maxTokens = %d, want 2048", service.maxTokens)
	}
}

func TestOpenAIService_ConfigValues(t *testing.T) {
	tests := []struct {
		name              string
		model             string
		maxTokens         int
		expectedModel     string
		expectedMaxTokens int
	}{
		{"Default GPT-4o", "gpt-4o", 4096, "gpt-4o", 4096},
		{"GPT-4 Turbo", "gpt-4-turbo", 8192, "gpt-4-turbo", 8192},
		{"GPT-3.5 Turbo", "gpt-3.5-turbo", 2048, "gpt-3.5-turbo", 2048},
		{"GPT-4o Mini", "gpt-4o-mini", 1024, "gpt-4o-mini", 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newOpenAIServiceWithClient(&mockOpenAIClient{}, tt.model, tt.maxTokens)
			if service.model != tt.expectedModel {
				t.Errorf("model = %s, want %s", service.model, tt.expectedModel)
			}
			if service.maxTokens != tt.expectedMaxTokens {
				t.Errorf("maxTokens = %d, want %d", service.maxTokens, tt.expectedMaxTokens)
			}
		})
	}
}

func TestOpenAIInvokeWithPrompt_Success(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Hello from GPT!",
						},
					},
				},
			}, nil
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	result, err := service.InvokeWithPrompt(ctx, "You are helpful", "Say hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Hello from GPT!" {
		t.Errorf("expected 'Hello from GPT!', got '%s'", result)
	}
}

func TestOpenAIInvokeWithPrompt_APIError(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return nil, errors.New("API error")
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	_, err := service.InvokeWithPrompt(ctx, "system", "user")
	if err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), "failed to invoke OpenAI") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOpenAIInvokeWithPrompt_EmptyChoices(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
			}, nil
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	_, err := service.InvokeWithPrompt(ctx, "system", "user")
	if err == nil {
		t.Error("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "empty response from OpenAI") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOpenAIInvokeStructured_Success(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `{"score": 85, "confidence": 90}`,
						},
					},
				},
			}, nil
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	type Result struct {
		Score      int `json:"score"`
		Confidence int `json:"confidence"`
	}

	var result Result
	err := service.InvokeStructured(ctx, "system", "user", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score != 85 || result.Confidence != 90 {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestOpenAIInvokeStructured_InvalidJSON(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "not valid json",
						},
					},
				},
			}, nil
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	var result map[string]interface{}
	err := service.InvokeStructured(ctx, "system", "user", &result)
	if err == nil {
		t.Error("expected error for invalid structured JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse response as JSON") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOpenAIInvokeStructured_APIError(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return nil, errors.New("API error")
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	var result map[string]interface{}
	err := service.InvokeStructured(ctx, "system", "user", &result)
	if err == nil {
		t.Error("expected error")
	}
}

func TestOpenAIChat_Success(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			// Verify that messages were converted correctly
			if len(params.Messages) != 4 { // 1 system + 3 conversation messages
				t.Errorf("expected 4 messages, got %d", len(params.Messages))
			}
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "I'm doing well, thank you!",
						},
					},
				},
			}, nil
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	result, err := service.Chat(ctx, "You are helpful", messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "I'm doing well, thank you!" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestOpenAIChat_APIError(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return nil, errors.New("API error")
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	_, err := service.Chat(ctx, "system", []ChatMessage{{Role: "user", Content: "test"}})
	if err == nil {
		t.Error("expected error")
	}
}

func TestOpenAIChat_EmptyChoices(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockClient := &mockOpenAIClient{
		completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
			}, nil
		},
	}

	service := newTestOpenAIService(mockClient)
	ctx := context.Background()

	_, err := service.Chat(ctx, "system", []ChatMessage{{Role: "user", Content: "test"}})
	if err == nil {
		t.Error("expected error for empty choices")
	}
}

func TestOpenAIService_ImplementsLLMService(t *testing.T) {
	// This test verifies the compile-time interface check
	// The var _ LLMService = (*OpenAIService)(nil) line
	// in interfaces.go will cause a compile error if the interface isn't implemented
	var _ LLMService = &OpenAIService{}
}

func TestOpenAIChat_MessageRoleConversion(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	tests := []struct {
		name     string
		messages []ChatMessage
	}{
		{
			name: "User only",
			messages: []ChatMessage{
				{Role: "user", Content: "Test message"},
			},
		},
		{
			name: "Mixed roles",
			messages: []ChatMessage{
				{Role: "user", Content: "First user"},
				{Role: "assistant", Content: "First assistant"},
				{Role: "user", Content: "Second user"},
			},
		},
		{
			name: "Unknown role ignored",
			messages: []ChatMessage{
				{Role: "user", Content: "User"},
				{Role: "unknown", Content: "Ignored"},
				{Role: "assistant", Content: "Assistant"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOpenAIClient{
				completionFunc: func(ctx context.Context, params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
					return &openai.ChatCompletion{
						Choices: []openai.ChatCompletionChoice{
							{
								Message: openai.ChatCompletionMessage{
									Content: "Response",
								},
							},
						},
					}, nil
				},
			}

			service := newTestOpenAIService(mockClient)
			ctx := context.Background()

			result, err := service.Chat(ctx, "system", tt.messages)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != "Response" {
				t.Errorf("unexpected result: %s", result)
			}
		})
	}
}
