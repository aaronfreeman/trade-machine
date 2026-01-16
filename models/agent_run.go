package models

import (
	"time"

	"github.com/google/uuid"
)

type AgentRun struct {
	ID           uuid.UUID              `json:"id"`
	AgentType    AgentType              `json:"agent_type"`
	Symbol       string                 `json:"symbol,omitempty"`
	Status       AgentRunStatus         `json:"status"`
	InputData    map[string]interface{} `json:"input_data,omitempty"`
	OutputData   map[string]interface{} `json:"output_data,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	DurationMs   int                    `json:"duration_ms"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

type AgentType string

const (
	AgentTypeFundamental AgentType = "fundamental"
	AgentTypeNews        AgentType = "news"
	AgentTypeTechnical   AgentType = "technical"
	AgentTypeManager     AgentType = "manager"
)

type AgentRunStatus string

const (
	AgentRunStatusRunning   AgentRunStatus = "running"
	AgentRunStatusCompleted AgentRunStatus = "completed"
	AgentRunStatusFailed    AgentRunStatus = "failed"
)

func NewAgentRun(agentType AgentType, symbol string) *AgentRun {
	return &AgentRun{
		ID:        uuid.New(),
		AgentType: agentType,
		Symbol:    symbol,
		Status:    AgentRunStatusRunning,
		StartedAt: time.Now(),
	}
}

func (r *AgentRun) Complete(output map[string]interface{}) {
	now := time.Now()
	r.CompletedAt = &now
	r.Status = AgentRunStatusCompleted
	r.OutputData = output
	r.DurationMs = int(now.Sub(r.StartedAt).Milliseconds())
}

func (r *AgentRun) Fail(err error) {
	now := time.Now()
	r.CompletedAt = &now
	r.Status = AgentRunStatusFailed
	r.ErrorMessage = err.Error()
	r.DurationMs = int(now.Sub(r.StartedAt).Milliseconds())
}
