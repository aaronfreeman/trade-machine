package models

import (
	"errors"
	"testing"
	"time"
)

func TestNewAgentRun(t *testing.T) {
	run := NewAgentRun(AgentTypeFundamental, "AAPL")

	if run.AgentType != AgentTypeFundamental {
		t.Errorf("AgentType = %v, want AgentTypeFundamental", run.AgentType)
	}
	if run.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", run.Symbol)
	}
	if run.Status != AgentRunStatusRunning {
		t.Errorf("Status = %v, want AgentRunStatusRunning", run.Status)
	}
	if run.ID == [16]byte{} {
		t.Error("ID should not be zero UUID")
	}
	if run.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}
	if run.CompletedAt != nil {
		t.Error("CompletedAt should be nil for new run")
	}
}

func TestAgentRun_Complete(t *testing.T) {
	run := NewAgentRun(AgentTypeNews, "MSFT")

	// Small delay to ensure duration > 0
	time.Sleep(5 * time.Millisecond)

	output := map[string]interface{}{
		"score":      75.5,
		"confidence": 80.0,
		"reasoning":  "Positive sentiment detected",
	}

	run.Complete(output)

	if run.Status != AgentRunStatusCompleted {
		t.Errorf("Status = %v, want AgentRunStatusCompleted", run.Status)
	}
	if run.CompletedAt == nil {
		t.Error("CompletedAt should not be nil after completion")
	}
	if run.OutputData == nil {
		t.Error("OutputData should not be nil after completion")
	}
	if run.OutputData["score"] != 75.5 {
		t.Errorf("OutputData[score] = %v, want 75.5", run.OutputData["score"])
	}
	if run.DurationMs <= 0 {
		t.Errorf("DurationMs = %v, should be > 0", run.DurationMs)
	}
	if run.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %v, want empty string", run.ErrorMessage)
	}
}

func TestAgentRun_Fail(t *testing.T) {
	run := NewAgentRun(AgentTypeTechnical, "GOOGL")

	// Small delay
	time.Sleep(5 * time.Millisecond)

	err := errors.New("failed to fetch price data")
	run.Fail(err)

	if run.Status != AgentRunStatusFailed {
		t.Errorf("Status = %v, want AgentRunStatusFailed", run.Status)
	}
	if run.CompletedAt == nil {
		t.Error("CompletedAt should not be nil after failure")
	}
	if run.ErrorMessage != "failed to fetch price data" {
		t.Errorf("ErrorMessage = %v, want 'failed to fetch price data'", run.ErrorMessage)
	}
	if run.DurationMs <= 0 {
		t.Errorf("DurationMs = %v, should be > 0", run.DurationMs)
	}
}

func TestAgentType_Constants(t *testing.T) {
	types := map[AgentType]string{
		AgentTypeFundamental: "fundamental",
		AgentTypeNews:        "news",
		AgentTypeTechnical:   "technical",
		AgentTypeManager:     "manager",
	}

	for agentType, expected := range types {
		if string(agentType) != expected {
			t.Errorf("AgentType %v = %v, want '%v'", agentType, string(agentType), expected)
		}
	}
}

func TestAgentRunStatus_Constants(t *testing.T) {
	statuses := map[AgentRunStatus]string{
		AgentRunStatusRunning:   "running",
		AgentRunStatusCompleted: "completed",
		AgentRunStatusFailed:    "failed",
	}

	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("AgentRunStatus %v = %v, want '%v'", status, string(status), expected)
		}
	}
}

func TestAgentRun_InputData(t *testing.T) {
	run := NewAgentRun(AgentTypeManager, "NVDA")
	run.InputData = map[string]interface{}{
		"watchlist":  []string{"NVDA", "AMD", "INTC"},
		"max_trades": 5,
		"risk_level": "moderate",
	}

	if run.InputData == nil {
		t.Error("InputData should not be nil")
	}
	if run.InputData["risk_level"] != "moderate" {
		t.Errorf("InputData[risk_level] = %v, want 'moderate'", run.InputData["risk_level"])
	}
}

func TestAgentRun_DurationCalculation(t *testing.T) {
	run := NewAgentRun(AgentTypeFundamental, "AAPL")

	// Sleep for a known duration
	time.Sleep(50 * time.Millisecond)

	run.Complete(map[string]interface{}{"result": "success"})

	// Duration should be at least 50ms
	if run.DurationMs < 50 {
		t.Errorf("DurationMs = %v, should be >= 50", run.DurationMs)
	}
	// But not too much more (allow some buffer)
	if run.DurationMs > 200 {
		t.Errorf("DurationMs = %v, should be < 200 (unexpectedly high)", run.DurationMs)
	}
}

func TestAgentRun_AllAgentTypes(t *testing.T) {
	agentTypes := []AgentType{
		AgentTypeFundamental,
		AgentTypeNews,
		AgentTypeTechnical,
		AgentTypeManager,
	}

	for _, agentType := range agentTypes {
		t.Run(string(agentType), func(t *testing.T) {
			run := NewAgentRun(agentType, "TEST")
			if run.AgentType != agentType {
				t.Errorf("AgentType = %v, want %v", run.AgentType, agentType)
			}
		})
	}
}
