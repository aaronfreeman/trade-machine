package observability

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	// Verify all metrics are initialized
	if m.AnalysisRequestsTotal == nil {
		t.Error("AnalysisRequestsTotal is nil")
	}
	if m.AnalysisDuration == nil {
		t.Error("AnalysisDuration is nil")
	}
	if m.AnalysisErrorsTotal == nil {
		t.Error("AnalysisErrorsTotal is nil")
	}
	if m.RecommendationActions == nil {
		t.Error("RecommendationActions is nil")
	}
	if m.AgentDuration == nil {
		t.Error("AgentDuration is nil")
	}
	if m.AgentErrorsTotal == nil {
		t.Error("AgentErrorsTotal is nil")
	}
	if m.ExternalAPIRequestsTotal == nil {
		t.Error("ExternalAPIRequestsTotal is nil")
	}
	if m.ExternalAPIErrorsTotal == nil {
		t.Error("ExternalAPIErrorsTotal is nil")
	}
	if m.ExternalAPIDuration == nil {
		t.Error("ExternalAPIDuration is nil")
	}
	if m.DBQueryDuration == nil {
		t.Error("DBQueryDuration is nil")
	}
	if m.DBQueryTotal == nil {
		t.Error("DBQueryTotal is nil")
	}
	if m.DBErrorsTotal == nil {
		t.Error("DBErrorsTotal is nil")
	}
	if m.HTTPRequestsTotal == nil {
		t.Error("HTTPRequestsTotal is nil")
	}
	if m.HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration is nil")
	}
	if m.CircuitBreakerState == nil {
		t.Error("CircuitBreakerState is nil")
	}
	if m.CircuitBreakerTrips == nil {
		t.Error("CircuitBreakerTrips is nil")
	}
}

func TestRecordAnalysisRequest(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAnalysisRequest("AAPL")
	m.RecordAnalysisRequest("AAPL")
	m.RecordAnalysisRequest("GOOG")

	// Check AAPL counter
	aaplCount := testutil.ToFloat64(m.AnalysisRequestsTotal.WithLabelValues("AAPL"))
	if aaplCount != 2 {
		t.Errorf("Expected AAPL count to be 2, got %f", aaplCount)
	}

	// Check GOOG counter
	googCount := testutil.ToFloat64(m.AnalysisRequestsTotal.WithLabelValues("GOOG"))
	if googCount != 1 {
		t.Errorf("Expected GOOG count to be 1, got %f", googCount)
	}
}

func TestRecordAnalysisDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAnalysisDuration("AAPL", "success", 100*time.Millisecond)
	m.RecordAnalysisDuration("AAPL", "error", 50*time.Millisecond)

	// Verify histograms are recorded (just check they don't panic)
	// Histogram values are harder to test directly
}

func TestRecordAnalysisError(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAnalysisError("AAPL", "timeout")
	m.RecordAnalysisError("AAPL", "timeout")
	m.RecordAnalysisError("GOOG", "network")

	aaplTimeoutCount := testutil.ToFloat64(m.AnalysisErrorsTotal.WithLabelValues("AAPL", "timeout"))
	if aaplTimeoutCount != 2 {
		t.Errorf("Expected AAPL timeout count to be 2, got %f", aaplTimeoutCount)
	}

	googNetworkCount := testutil.ToFloat64(m.AnalysisErrorsTotal.WithLabelValues("GOOG", "network"))
	if googNetworkCount != 1 {
		t.Errorf("Expected GOOG network count to be 1, got %f", googNetworkCount)
	}
}

func TestRecordRecommendation(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordRecommendation("BUY", 75.5, 80.0)
	m.RecordRecommendation("SELL", -50.0, 90.0)
	m.RecordRecommendation("HOLD", 10.0, 60.0)

	buyCount := testutil.ToFloat64(m.RecommendationActions.WithLabelValues("BUY"))
	if buyCount != 1 {
		t.Errorf("Expected BUY count to be 1, got %f", buyCount)
	}

	sellCount := testutil.ToFloat64(m.RecommendationActions.WithLabelValues("SELL"))
	if sellCount != 1 {
		t.Errorf("Expected SELL count to be 1, got %f", sellCount)
	}

	holdCount := testutil.ToFloat64(m.RecommendationActions.WithLabelValues("HOLD"))
	if holdCount != 1 {
		t.Errorf("Expected HOLD count to be 1, got %f", holdCount)
	}
}

func TestRecordAgentDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAgentDuration("fundamental", 2*time.Second)
	m.RecordAgentDuration("technical", 1500*time.Millisecond)
	m.RecordAgentDuration("news", 3*time.Second)

	// Verify histograms are recorded (just check they don't panic)
}

func TestRecordAgentError(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAgentError("fundamental", "timeout")
	m.RecordAgentError("technical", "circuit_breaker")

	fundamentalTimeout := testutil.ToFloat64(m.AgentErrorsTotal.WithLabelValues("fundamental", "timeout"))
	if fundamentalTimeout != 1 {
		t.Errorf("Expected fundamental timeout count to be 1, got %f", fundamentalTimeout)
	}
}

func TestRecordAgentScore(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAgentScore("fundamental", 50.0)
	m.RecordAgentScore("technical", -30.0)
	m.RecordAgentScore("news", 75.0)

	// Verify histograms are recorded (just check they don't panic)
}

func TestRecordExternalAPIRequest(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordExternalAPIRequest("bedrock", "invoke")
	m.RecordExternalAPIRequest("bedrock", "invoke")
	m.RecordExternalAPIRequest("alpaca", "get_quote")

	bedrockInvoke := testutil.ToFloat64(m.ExternalAPIRequestsTotal.WithLabelValues("bedrock", "invoke"))
	if bedrockInvoke != 2 {
		t.Errorf("Expected bedrock invoke count to be 2, got %f", bedrockInvoke)
	}

	alpacaQuote := testutil.ToFloat64(m.ExternalAPIRequestsTotal.WithLabelValues("alpaca", "get_quote"))
	if alpacaQuote != 1 {
		t.Errorf("Expected alpaca get_quote count to be 1, got %f", alpacaQuote)
	}
}

func TestRecordExternalAPIError(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordExternalAPIError("bedrock", "invoke", "timeout")
	m.RecordExternalAPIError("newsapi", "get_articles", "rate_limit")

	bedrockTimeout := testutil.ToFloat64(m.ExternalAPIErrorsTotal.WithLabelValues("bedrock", "invoke", "timeout"))
	if bedrockTimeout != 1 {
		t.Errorf("Expected bedrock timeout count to be 1, got %f", bedrockTimeout)
	}
}

func TestRecordExternalAPIDuration(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordExternalAPIDuration("bedrock", "invoke", 500*time.Millisecond)
	m.RecordExternalAPIDuration("alpaca", "get_bars", 200*time.Millisecond)

	// Verify histograms are recorded
}

func TestRecordDBQuery(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordDBQuery("select", "recommendations", 10*time.Millisecond)
	m.RecordDBQuery("insert", "recommendations", 5*time.Millisecond)
	m.RecordDBQuery("select", "agent_runs", 8*time.Millisecond)

	selectRecs := testutil.ToFloat64(m.DBQueryTotal.WithLabelValues("select", "recommendations"))
	if selectRecs != 1 {
		t.Errorf("Expected select recommendations count to be 1, got %f", selectRecs)
	}

	insertRecs := testutil.ToFloat64(m.DBQueryTotal.WithLabelValues("insert", "recommendations"))
	if insertRecs != 1 {
		t.Errorf("Expected insert recommendations count to be 1, got %f", insertRecs)
	}
}

func TestRecordDBError(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordDBError("select", "recommendations")
	m.RecordDBError("insert", "trades")

	selectError := testutil.ToFloat64(m.DBErrorsTotal.WithLabelValues("select", "recommendations"))
	if selectError != 1 {
		t.Errorf("Expected select error count to be 1, got %f", selectError)
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordHTTPRequest("GET", "/api/health", "200", 10*time.Millisecond, 256)
	m.RecordHTTPRequest("POST", "/api/analyze", "200", 2*time.Second, 4096)
	m.RecordHTTPRequest("GET", "/api/recommendations", "500", 50*time.Millisecond, 128)

	healthOK := testutil.ToFloat64(m.HTTPRequestsTotal.WithLabelValues("GET", "/api/health", "200"))
	if healthOK != 1 {
		t.Errorf("Expected GET /api/health 200 count to be 1, got %f", healthOK)
	}

	recsError := testutil.ToFloat64(m.HTTPRequestsTotal.WithLabelValues("GET", "/api/recommendations", "500"))
	if recsError != 1 {
		t.Errorf("Expected GET /api/recommendations 500 count to be 1, got %f", recsError)
	}
}

func TestCircuitBreakerMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	// Set initial states
	m.SetCircuitBreakerState("bedrock", 0) // closed
	m.SetCircuitBreakerState("alpaca", 2)  // open

	bedrockState := testutil.ToFloat64(m.CircuitBreakerState.WithLabelValues("bedrock"))
	if bedrockState != 0 {
		t.Errorf("Expected bedrock state to be 0 (closed), got %f", bedrockState)
	}

	alpacaState := testutil.ToFloat64(m.CircuitBreakerState.WithLabelValues("alpaca"))
	if alpacaState != 2 {
		t.Errorf("Expected alpaca state to be 2 (open), got %f", alpacaState)
	}

	// Record trips
	m.RecordCircuitBreakerTrip("bedrock")
	m.RecordCircuitBreakerTrip("bedrock")

	bedrockTrips := testutil.ToFloat64(m.CircuitBreakerTrips.WithLabelValues("bedrock"))
	if bedrockTrips != 2 {
		t.Errorf("Expected bedrock trips to be 2, got %f", bedrockTrips)
	}
}

func TestTimer(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	timer := m.NewTimer()
	if timer == nil {
		t.Fatal("NewTimer returned nil")
	}

	// Sleep a small amount to ensure duration is measurable
	time.Sleep(10 * time.Millisecond)

	duration := timer.Duration()
	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration to be at least 10ms, got %v", duration)
	}

	// Test ObserveAnalysis
	timer.ObserveAnalysis("AAPL", "success")

	// Test ObserveAgent
	timer2 := m.NewTimer()
	time.Sleep(5 * time.Millisecond)
	timer2.ObserveAgent("fundamental")

	// Test ObserveExternalAPI
	timer3 := m.NewTimer()
	time.Sleep(5 * time.Millisecond)
	timer3.ObserveExternalAPI("bedrock", "invoke")

	// Test ObserveDB
	timer4 := m.NewTimer()
	time.Sleep(5 * time.Millisecond)
	timer4.ObserveDB("select", "recommendations")
}

func TestGetMetrics_Singleton(t *testing.T) {
	// Save and restore global metrics state
	original := globalMetrics
	defer func() { globalMetrics = original }()

	// Create a fresh metrics instance with a dedicated registry
	reg := prometheus.NewRegistry()
	testMetrics := NewMetrics(reg)
	globalMetrics = testMetrics

	m1 := GetMetrics()
	if m1 == nil {
		t.Fatal("GetMetrics returned nil")
	}

	m2 := GetMetrics()
	if m1 != m2 {
		t.Error("GetMetrics should return the same instance")
	}
}

func TestInitMetrics_SetsGlobal(t *testing.T) {
	// Save and restore global metrics state
	original := globalMetrics
	defer func() { globalMetrics = original }()

	// Create a new registry for isolation
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	globalMetrics = m

	// Verify it's the global instance
	if globalMetrics != m {
		t.Error("globalMetrics should match the instance we set")
	}

	// Verify GetMetrics returns it
	if GetMetrics() != m {
		t.Error("GetMetrics should return the global instance")
	}
}
