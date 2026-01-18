package models

import (
	"time"

	"github.com/google/uuid"
)

// ScreenerRunStatus represents the status of a screener run
type ScreenerRunStatus string

const (
	ScreenerRunStatusRunning   ScreenerRunStatus = "running"
	ScreenerRunStatusCompleted ScreenerRunStatus = "completed"
	ScreenerRunStatusFailed    ScreenerRunStatus = "failed"
)

// ScreenerRun represents a single execution of the value screener
type ScreenerRun struct {
	ID         uuid.UUID           `json:"id"`
	RunAt      time.Time           `json:"run_at"`
	Criteria   ScreenerCriteria    `json:"criteria"`
	Candidates []ScreenerCandidate `json:"candidates"`
	TopPicks   []uuid.UUID         `json:"top_picks"` // Recommendation IDs
	DurationMs int64               `json:"duration_ms"`
	Status     ScreenerRunStatus   `json:"status"`
	Error      string              `json:"error,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
}

// ScreenerCriteria defines the filtering criteria used for a screener run
type ScreenerCriteria struct {
	MarketCapMin     int64   `json:"market_cap_min"`
	MarketCapMax     int64   `json:"market_cap_max,omitempty"`
	PERatioMax       float64 `json:"pe_ratio_max"`
	PBRatioMax       float64 `json:"pb_ratio_max"`
	EPSMin           float64 `json:"eps_min"`
	DividendYieldMin float64 `json:"dividend_yield_min,omitempty"`
	Sector           string  `json:"sector,omitempty"`
	Limit            int     `json:"limit"`
}

// ScreenerCandidate represents a stock candidate from the screener
type ScreenerCandidate struct {
	Symbol        string   `json:"symbol"`
	CompanyName   string   `json:"company_name"`
	MarketCap     int64    `json:"market_cap"`
	PERatio       float64  `json:"pe_ratio"`
	PBRatio       float64  `json:"pb_ratio"`
	EPS           float64  `json:"eps"`
	DividendYield float64  `json:"dividend_yield"`
	Sector        string   `json:"sector"`
	Industry      string   `json:"industry"`
	Price         float64  `json:"price"`
	Beta          float64  `json:"beta"`
	ValueScore    float64  `json:"value_score"`    // Pre-filter score
	Score         *float64 `json:"score,omitempty"` // After full analysis
	Confidence    *float64 `json:"confidence,omitempty"`
	Analyzed      bool     `json:"analyzed"`
}

// NewScreenerRun creates a new ScreenerRun with default values
func NewScreenerRun(criteria ScreenerCriteria) *ScreenerRun {
	now := time.Now()
	return &ScreenerRun{
		ID:         uuid.New(),
		RunAt:      now,
		Criteria:   criteria,
		Candidates: []ScreenerCandidate{},
		TopPicks:   []uuid.UUID{},
		Status:     ScreenerRunStatusRunning,
		CreatedAt:  now,
	}
}

// Complete marks the screener run as completed
func (s *ScreenerRun) Complete(durationMs int64, topPicks []uuid.UUID) {
	s.Status = ScreenerRunStatusCompleted
	s.DurationMs = durationMs
	s.TopPicks = topPicks
}

// Fail marks the screener run as failed with an error message
func (s *ScreenerRun) Fail(err string, durationMs int64) {
	s.Status = ScreenerRunStatusFailed
	s.Error = err
	s.DurationMs = durationMs
}

// AddCandidate adds a candidate to the screener run
func (s *ScreenerRun) AddCandidate(candidate ScreenerCandidate) {
	s.Candidates = append(s.Candidates, candidate)
}

// SetCandidates sets all candidates at once
func (s *ScreenerRun) SetCandidates(candidates []ScreenerCandidate) {
	s.Candidates = candidates
}

// IsRunning returns true if the screener run is still in progress
func (s *ScreenerRun) IsRunning() bool {
	return s.Status == ScreenerRunStatusRunning
}

// IsCompleted returns true if the screener run completed successfully
func (s *ScreenerRun) IsCompleted() bool {
	return s.Status == ScreenerRunStatusCompleted
}

// IsFailed returns true if the screener run failed
func (s *ScreenerRun) IsFailed() bool {
	return s.Status == ScreenerRunStatusFailed
}
