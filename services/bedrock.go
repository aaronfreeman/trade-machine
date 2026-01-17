package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockService handles communication with AWS Bedrock for Claude models
type BedrockService struct {
	client *bedrockruntime.Client
	model  string
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
func NewBedrockService(ctx context.Context, region, modelID string) (*BedrockService, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &BedrockService{
		client: bedrockruntime.NewFromConfig(cfg),
		model:  modelID,
	}, nil
}

// InvokeWithPrompt sends a prompt to Claude and returns the response text
func (s *BedrockService) InvokeWithPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	maxTokens := 4096
	if val := os.Getenv("BEDROCK_MAX_TOKENS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			maxTokens = parsed
		}
	}

	anthropicVersion := "bedrock-2023-05-31"
	if val := os.Getenv("BEDROCK_ANTHROPIC_VERSION"); val != "" {
		anthropicVersion = val
	}

	request := ClaudeRequest{
		AnthropicVersion: anthropicVersion,
		MaxTokens:        maxTokens,
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
	request := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
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
}
