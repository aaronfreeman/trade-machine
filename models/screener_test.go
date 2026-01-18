package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewScreenerRun(t *testing.T) {
	criteria := ScreenerCriteria{
		MarketCapMin: 1_000_000_000,
		PERatioMax:   15.0,
		PBRatioMax:   1.5,
		EPSMin:       0.01,
		Limit:        20,
	}

	run := NewScreenerRun(criteria)

	if run.ID == uuid.Nil {
		t.Error("ID should not be nil UUID")
	}
	if run.RunAt.IsZero() {
		t.Error("RunAt should not be zero")
	}
	if run.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if run.Status != ScreenerRunStatusRunning {
		t.Errorf("Status = %v, want ScreenerRunStatusRunning", run.Status)
	}
	if run.Criteria.MarketCapMin != 1_000_000_000 {
		t.Errorf("Criteria.MarketCapMin = %v, want 1000000000", run.Criteria.MarketCapMin)
	}
	if len(run.Candidates) != 0 {
		t.Errorf("Candidates length = %v, want 0", len(run.Candidates))
	}
	if len(run.TopPicks) != 0 {
		t.Errorf("TopPicks length = %v, want 0", len(run.TopPicks))
	}
}

func TestScreenerRun_Complete(t *testing.T) {
	run := NewScreenerRun(ScreenerCriteria{})
	topPicks := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	run.Complete(5000, topPicks)

	if run.Status != ScreenerRunStatusCompleted {
		t.Errorf("Status = %v, want ScreenerRunStatusCompleted", run.Status)
	}
	if run.DurationMs != 5000 {
		t.Errorf("DurationMs = %v, want 5000", run.DurationMs)
	}
	if len(run.TopPicks) != 3 {
		t.Errorf("TopPicks length = %v, want 3", len(run.TopPicks))
	}
}

func TestScreenerRun_Fail(t *testing.T) {
	run := NewScreenerRun(ScreenerCriteria{})

	run.Fail("API rate limit exceeded", 1500)

	if run.Status != ScreenerRunStatusFailed {
		t.Errorf("Status = %v, want ScreenerRunStatusFailed", run.Status)
	}
	if run.Error != "API rate limit exceeded" {
		t.Errorf("Error = %v, want 'API rate limit exceeded'", run.Error)
	}
	if run.DurationMs != 1500 {
		t.Errorf("DurationMs = %v, want 1500", run.DurationMs)
	}
}

func TestScreenerRun_AddCandidate(t *testing.T) {
	run := NewScreenerRun(ScreenerCriteria{})

	candidate := ScreenerCandidate{
		Symbol:      "AAPL",
		CompanyName: "Apple Inc.",
		MarketCap:   2_500_000_000_000,
		PERatio:     28.5,
		PBRatio:     45.0,
	}

	run.AddCandidate(candidate)

	if len(run.Candidates) != 1 {
		t.Errorf("Candidates length = %v, want 1", len(run.Candidates))
	}
	if run.Candidates[0].Symbol != "AAPL" {
		t.Errorf("Candidate symbol = %v, want 'AAPL'", run.Candidates[0].Symbol)
	}
}

func TestScreenerRun_SetCandidates(t *testing.T) {
	run := NewScreenerRun(ScreenerCriteria{})

	candidates := []ScreenerCandidate{
		{Symbol: "AAPL", CompanyName: "Apple Inc."},
		{Symbol: "MSFT", CompanyName: "Microsoft Corporation"},
		{Symbol: "GOOGL", CompanyName: "Alphabet Inc."},
	}

	run.SetCandidates(candidates)

	if len(run.Candidates) != 3 {
		t.Errorf("Candidates length = %v, want 3", len(run.Candidates))
	}
}

func TestScreenerRun_StatusChecks(t *testing.T) {
	run := NewScreenerRun(ScreenerCriteria{})

	// Initial state: running
	if !run.IsRunning() {
		t.Error("IsRunning should return true for new run")
	}
	if run.IsCompleted() {
		t.Error("IsCompleted should return false for new run")
	}
	if run.IsFailed() {
		t.Error("IsFailed should return false for new run")
	}

	// Completed state
	run.Complete(1000, nil)
	if run.IsRunning() {
		t.Error("IsRunning should return false after completion")
	}
	if !run.IsCompleted() {
		t.Error("IsCompleted should return true after completion")
	}
	if run.IsFailed() {
		t.Error("IsFailed should return false after completion")
	}

	// Failed state
	run2 := NewScreenerRun(ScreenerCriteria{})
	run2.Fail("error", 100)
	if run2.IsRunning() {
		t.Error("IsRunning should return false after failure")
	}
	if run2.IsCompleted() {
		t.Error("IsCompleted should return false after failure")
	}
	if !run2.IsFailed() {
		t.Error("IsFailed should return true after failure")
	}
}

func TestScreenerRunStatus_Constants(t *testing.T) {
	statuses := map[ScreenerRunStatus]string{
		ScreenerRunStatusRunning:   "running",
		ScreenerRunStatusCompleted: "completed",
		ScreenerRunStatusFailed:    "failed",
	}

	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("ScreenerRunStatus %v = %v, want '%v'", status, string(status), expected)
		}
	}
}

func TestScreenerCriteria(t *testing.T) {
	criteria := ScreenerCriteria{
		MarketCapMin:     1_000_000_000,
		MarketCapMax:     100_000_000_000,
		PERatioMax:       15.0,
		PBRatioMax:       1.5,
		EPSMin:           0.5,
		DividendYieldMin: 2.0,
		Sector:           "Technology",
		Limit:            25,
	}

	if criteria.MarketCapMin != 1_000_000_000 {
		t.Errorf("MarketCapMin = %v, want 1000000000", criteria.MarketCapMin)
	}
	if criteria.MarketCapMax != 100_000_000_000 {
		t.Errorf("MarketCapMax = %v, want 100000000000", criteria.MarketCapMax)
	}
	if criteria.PERatioMax != 15.0 {
		t.Errorf("PERatioMax = %v, want 15.0", criteria.PERatioMax)
	}
	if criteria.Sector != "Technology" {
		t.Errorf("Sector = %v, want 'Technology'", criteria.Sector)
	}
}

func TestScreenerCandidate(t *testing.T) {
	score := 75.5
	confidence := 85.0

	candidate := ScreenerCandidate{
		Symbol:        "JNJ",
		CompanyName:   "Johnson & Johnson",
		MarketCap:     400_000_000_000,
		PERatio:       15.0,
		PBRatio:       5.5,
		EPS:           10.0,
		DividendYield: 3.0,
		Sector:        "Healthcare",
		Industry:      "Pharmaceuticals",
		Price:         155.00,
		Beta:          0.65,
		ValueScore:    80.0,
		Score:         &score,
		Confidence:    &confidence,
		Analyzed:      true,
	}

	if candidate.Symbol != "JNJ" {
		t.Errorf("Symbol = %v, want 'JNJ'", candidate.Symbol)
	}
	if candidate.Score == nil || *candidate.Score != 75.5 {
		t.Errorf("Score = %v, want 75.5", candidate.Score)
	}
	if candidate.Confidence == nil || *candidate.Confidence != 85.0 {
		t.Errorf("Confidence = %v, want 85.0", candidate.Confidence)
	}
	if !candidate.Analyzed {
		t.Error("Analyzed should be true")
	}
}

func TestScreenerCandidate_BeforeAnalysis(t *testing.T) {
	candidate := ScreenerCandidate{
		Symbol:      "AAPL",
		CompanyName: "Apple Inc.",
		MarketCap:   2_500_000_000_000,
		PERatio:     28.5,
		ValueScore:  50.0,
		Analyzed:    false,
	}

	if candidate.Score != nil {
		t.Error("Score should be nil before analysis")
	}
	if candidate.Confidence != nil {
		t.Error("Confidence should be nil before analysis")
	}
	if candidate.Analyzed {
		t.Error("Analyzed should be false before analysis")
	}
}

func TestScreenerRun_FullWorkflow(t *testing.T) {
	// Create screener run with criteria
	criteria := ScreenerCriteria{
		MarketCapMin: 1_000_000_000,
		PERatioMax:   15.0,
		Limit:        10,
	}
	run := NewScreenerRun(criteria)

	if !run.IsRunning() {
		t.Error("Run should be in running state initially")
	}

	// Add candidates from screener
	candidates := []ScreenerCandidate{
		{Symbol: "JNJ", CompanyName: "Johnson & Johnson", PERatio: 12.0, ValueScore: 85.0},
		{Symbol: "PG", CompanyName: "Procter & Gamble", PERatio: 14.0, ValueScore: 75.0},
		{Symbol: "KO", CompanyName: "Coca-Cola", PERatio: 13.5, ValueScore: 70.0},
	}
	run.SetCandidates(candidates)

	if len(run.Candidates) != 3 {
		t.Errorf("Candidates = %v, want 3", len(run.Candidates))
	}

	// Complete with top picks
	topPicks := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	run.Complete(120000, topPicks) // 2 minutes

	if !run.IsCompleted() {
		t.Error("Run should be completed")
	}
	if run.DurationMs != 120000 {
		t.Errorf("DurationMs = %v, want 120000", run.DurationMs)
	}
	if len(run.TopPicks) != 3 {
		t.Errorf("TopPicks = %v, want 3", len(run.TopPicks))
	}
}
