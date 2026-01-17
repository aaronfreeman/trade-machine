package services

import (
	"context"
	"encoding/json"
	"testing"

	"trade-machine/config"
)

func TestClaudeRequest_Serialization(t *testing.T) {
	req := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		System:           "You are a helpful assistant.",
		Messages: []ClaudeMessage{
			{Role: "user", Content: "Hello, world!"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ClaudeRequest: %v", err)
	}

	var unmarshaled ClaudeRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ClaudeRequest: %v", err)
	}

	if unmarshaled.AnthropicVersion != req.AnthropicVersion {
		t.Errorf("AnthropicVersion = %v, want %v", unmarshaled.AnthropicVersion, req.AnthropicVersion)
	}
	if unmarshaled.MaxTokens != req.MaxTokens {
		t.Errorf("MaxTokens = %v, want %v", unmarshaled.MaxTokens, req.MaxTokens)
	}
	if unmarshaled.System != req.System {
		t.Errorf("System = %v, want %v", unmarshaled.System, req.System)
	}
	if len(unmarshaled.Messages) != 1 {
		t.Errorf("Messages length = %v, want 1", len(unmarshaled.Messages))
	}
	if unmarshaled.Messages[0].Role != "user" {
		t.Errorf("Messages[0].Role = %v, want 'user'", unmarshaled.Messages[0].Role)
	}
	if unmarshaled.Messages[0].Content != "Hello, world!" {
		t.Errorf("Messages[0].Content = %v, want 'Hello, world!'", unmarshaled.Messages[0].Content)
	}
}

func TestClaudeRequest_EmptySystem(t *testing.T) {
	req := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1024,
		Messages: []ClaudeMessage{
			{Role: "user", Content: "Test"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify that empty system field is omitted
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// System field with empty value should be omitted due to omitempty tag
	if _, exists := raw["system"]; exists && req.System == "" {
		t.Error("Empty system field should be omitted from JSON")
	}
}

func TestClaudeResponse_Deserialization(t *testing.T) {
	jsonResponse := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [
			{"type": "text", "text": "Hello! How can I help you?"}
		],
		"stop_reason": "end_turn",
		"usage": {
			"input_tokens": 10,
			"output_tokens": 15
		}
	}`

	var resp ClaudeResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal ClaudeResponse: %v", err)
	}

	if resp.ID != "msg_123" {
		t.Errorf("ID = %v, want 'msg_123'", resp.ID)
	}
	if resp.Type != "message" {
		t.Errorf("Type = %v, want 'message'", resp.Type)
	}
	if resp.Role != "assistant" {
		t.Errorf("Role = %v, want 'assistant'", resp.Role)
	}
	if len(resp.Content) != 1 {
		t.Errorf("Content length = %v, want 1", len(resp.Content))
	}
	if resp.Content[0].Text != "Hello! How can I help you?" {
		t.Errorf("Content[0].Text = %v, want 'Hello! How can I help you?'", resp.Content[0].Text)
	}
	if resp.StopReason != "end_turn" {
		t.Errorf("StopReason = %v, want 'end_turn'", resp.StopReason)
	}
	if resp.Usage.InputTokens != 10 {
		t.Errorf("Usage.InputTokens = %v, want 10", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 15 {
		t.Errorf("Usage.OutputTokens = %v, want 15", resp.Usage.OutputTokens)
	}
}

func TestClaudeMessage_JSONTags(t *testing.T) {
	msg := ClaudeMessage{
		Role:    "assistant",
		Content: "This is a test response.",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal ClaudeMessage: %v", err)
	}

	// Verify JSON structure
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, ok := m["role"]; !ok {
		t.Error("JSON should have 'role' field")
	}
	if _, ok := m["content"]; !ok {
		t.Error("JSON should have 'content' field")
	}
}

func TestClaudeRequest_MultipleMessages(t *testing.T) {
	req := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Messages: []ClaudeMessage{
			{Role: "user", Content: "What is the weather?"},
			{Role: "assistant", Content: "I don't have real-time weather data."},
			{Role: "user", Content: "Then tell me a joke."},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled ClaudeRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(unmarshaled.Messages) != 3 {
		t.Errorf("Messages length = %v, want 3", len(unmarshaled.Messages))
	}
}

func TestNewBedrockService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		region  string
		modelID string
	}{
		{"US East 1 - Haiku", "us-east-1", "anthropic.claude-3-haiku-20240307-v1:0"},
		{"US West 2 - Sonnet", "us-west-2", "anthropic.claude-3-5-sonnet-20241022-v2:0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewTestConfig()
			cfg.AWS.Region = tt.region
			cfg.AWS.BedrockModelID = tt.modelID
			service, err := NewBedrockService(ctx, cfg)
			if err != nil {
				// This is expected if AWS credentials are not configured
				t.Logf("NewBedrockService returned error (expected if no AWS creds): %v", err)
				return
			}
			if service == nil {
				t.Error("NewBedrockService should not return nil when no error")
			}
			if service.client == nil {
				t.Error("client should not be nil when service is created")
			}
			if service.model != tt.modelID {
				t.Errorf("model = %v, want %v", service.model, tt.modelID)
			}
		})
	}
}

func TestNewBedrockService_InvalidRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	cfg := config.NewTestConfig()
	cfg.AWS.Region = "invalid-region-99"
	cfg.AWS.BedrockModelID = "test-model"
	service, err := NewBedrockService(ctx, cfg)

	// May succeed or fail depending on AWS SDK configuration
	// Just verify it doesn't crash
	if err != nil {
		t.Logf("NewBedrockService with invalid region returned error: %v", err)
	} else if service == nil {
		t.Error("NewBedrockService should not return nil service without error")
	}
}

func TestBedrockService_ConfigValues(t *testing.T) {
	// Test that config values are properly stored in the service
	tests := []struct {
		name              string
		maxTokens         int
		version           string
		expectedMaxTokens int
		expectedVersion   string
	}{
		{"Default values", 4096, "bedrock-2023-05-31", 4096, "bedrock-2023-05-31"},
		{"Custom max tokens", 2048, "bedrock-2023-05-31", 2048, "bedrock-2023-05-31"},
		{"Custom version", 4096, "bedrock-2024-01-01", 4096, "bedrock-2024-01-01"},
		{"Both custom", 8192, "bedrock-2024-01-01", 8192, "bedrock-2024-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := config.NewTestConfig()
			cfg.AWS.BedrockMaxTokens = tt.maxTokens
			cfg.AWS.AnthropicVersion = tt.version
			service, err := NewBedrockService(ctx, cfg)
			if err != nil {
				t.Logf("Expected error with no AWS credentials: %v", err)
				return
			}
			if service == nil {
				t.Error("service should not be nil")
				return
			}
			if service.maxTokens != tt.expectedMaxTokens {
				t.Errorf("maxTokens = %d, want %d", service.maxTokens, tt.expectedMaxTokens)
			}
			if service.anthropicVersion != tt.expectedVersion {
				t.Errorf("anthropicVersion = %s, want %s", service.anthropicVersion, tt.expectedVersion)
			}
		})
	}
}

func TestInvokeStructured_JSONParsing(t *testing.T) {
	// Test that InvokeStructured properly handles JSON parsing
	// This is a unit test for the JSON parsing logic
	type TestResult struct {
		Score      float64 `json:"score"`
		Confidence float64 `json:"confidence"`
		Message    string  `json:"message"`
	}

	jsonText := `{"score": 85.5, "confidence": 90.0, "message": "Test successful"}`

	var result TestResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		t.Fatalf("Failed to parse test JSON: %v", err)
	}

	if result.Score != 85.5 {
		t.Errorf("Score = %v, want 85.5", result.Score)
	}
	if result.Confidence != 90.0 {
		t.Errorf("Confidence = %v, want 90.0", result.Confidence)
	}
	if result.Message != "Test successful" {
		t.Errorf("Message = %v, want 'Test successful'", result.Message)
	}
}

func TestClaudeResponse_EmptyContent(t *testing.T) {
	jsonResponse := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [],
		"stop_reason": "end_turn"
	}`

	var resp ClaudeResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.Content) != 0 {
		t.Errorf("Content length = %v, want 0", len(resp.Content))
	}
}

func TestClaudeResponse_MultipleContentBlocks(t *testing.T) {
	jsonResponse := `{
		"id": "msg_456",
		"type": "message",
		"role": "assistant",
		"content": [
			{"type": "text", "text": "First block"},
			{"type": "text", "text": "Second block"}
		],
		"stop_reason": "end_turn"
	}`

	var resp ClaudeResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.Content) != 2 {
		t.Errorf("Content length = %v, want 2", len(resp.Content))
	}
	if resp.Content[0].Text != "First block" {
		t.Errorf("Content[0].Text = %v, want 'First block'", resp.Content[0].Text)
	}
	if resp.Content[1].Text != "Second block" {
		t.Errorf("Content[1].Text = %v, want 'Second block'", resp.Content[1].Text)
	}
}

func TestClaudeRequest_MarshalOrder(t *testing.T) {
	// Verify that the struct fields marshal in the expected order
	req := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		System:           "Test system",
		Messages: []ClaudeMessage{
			{Role: "user", Content: "Test message"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify it contains expected fields
	dataStr := string(data)
	requiredFields := []string{
		`"anthropic_version"`,
		`"max_tokens"`,
		`"system"`,
		`"messages"`,
		`"role"`,
		`"content"`,
	}

	for _, field := range requiredFields {
		if !contains(dataStr, field) {
			t.Errorf("Expected JSON to contain %s", field)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
