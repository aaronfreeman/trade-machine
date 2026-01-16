package services

import (
	"encoding/json"
	"testing"
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
