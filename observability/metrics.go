package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// Analysis metrics
	AnalysisRequestsTotal    *prometheus.CounterVec
	AnalysisDuration         *prometheus.HistogramVec
	AnalysisErrorsTotal      *prometheus.CounterVec
	RecommendationActions    *prometheus.CounterVec
	RecommendationScores     *prometheus.HistogramVec
	RecommendationConfidence *prometheus.HistogramVec

	// Agent metrics
	AgentDuration    *prometheus.HistogramVec
	AgentErrorsTotal *prometheus.CounterVec
	AgentScores      *prometheus.HistogramVec

	// External API metrics
	ExternalAPIRequestsTotal *prometheus.CounterVec
	ExternalAPIErrorsTotal   *prometheus.CounterVec
	ExternalAPIDuration      *prometheus.HistogramVec

	// Database metrics
	DBQueryDuration *prometheus.HistogramVec
	DBQueryTotal    *prometheus.CounterVec
	DBErrorsTotal   *prometheus.CounterVec

	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Circuit breaker metrics
	CircuitBreakerState *prometheus.GaugeVec
	CircuitBreakerTrips *prometheus.CounterVec
}

// defaultBuckets are the default histogram buckets for duration metrics (in seconds)
var defaultBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30}

// scoreBuckets are histogram buckets for score metrics (-100 to 100)
var scoreBuckets = []float64{-100, -75, -50, -25, 0, 25, 50, 75, 100}

// confidenceBuckets are histogram buckets for confidence metrics (0 to 100)
var confidenceBuckets = []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

// globalMetrics is the global metrics instance
var globalMetrics *Metrics

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	factory := promauto.With(reg)

	m := &Metrics{
		// Analysis metrics
		AnalysisRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "analysis",
				Name:      "requests_total",
				Help:      "Total number of stock analysis requests",
			},
			[]string{"symbol"},
		),
		AnalysisDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "analysis",
				Name:      "duration_seconds",
				Help:      "Duration of stock analysis in seconds",
				Buckets:   defaultBuckets,
			},
			[]string{"symbol", "status"},
		),
		AnalysisErrorsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "analysis",
				Name:      "errors_total",
				Help:      "Total number of analysis errors",
			},
			[]string{"symbol", "error_type"},
		),
		RecommendationActions: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "recommendation",
				Name:      "actions_total",
				Help:      "Total number of recommendations by action type",
			},
			[]string{"action"},
		),
		RecommendationScores: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "recommendation",
				Name:      "score",
				Help:      "Distribution of recommendation scores",
				Buckets:   scoreBuckets,
			},
			[]string{"action"},
		),
		RecommendationConfidence: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "recommendation",
				Name:      "confidence",
				Help:      "Distribution of recommendation confidence levels",
				Buckets:   confidenceBuckets,
			},
			[]string{"action"},
		),

		// Agent metrics
		AgentDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "agent",
				Name:      "duration_seconds",
				Help:      "Duration of agent analysis in seconds",
				Buckets:   defaultBuckets,
			},
			[]string{"agent_type"},
		),
		AgentErrorsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "agent",
				Name:      "errors_total",
				Help:      "Total number of agent errors",
			},
			[]string{"agent_type", "error_type"},
		),
		AgentScores: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "agent",
				Name:      "score",
				Help:      "Distribution of agent analysis scores",
				Buckets:   scoreBuckets,
			},
			[]string{"agent_type"},
		),

		// External API metrics
		ExternalAPIRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "external_api",
				Name:      "requests_total",
				Help:      "Total number of external API requests",
			},
			[]string{"service", "operation"},
		),
		ExternalAPIErrorsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "external_api",
				Name:      "errors_total",
				Help:      "Total number of external API errors",
			},
			[]string{"service", "operation", "error_type"},
		),
		ExternalAPIDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "external_api",
				Name:      "duration_seconds",
				Help:      "Duration of external API calls in seconds",
				Buckets:   defaultBuckets,
			},
			[]string{"service", "operation"},
		),

		// Database metrics
		DBQueryDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "database",
				Name:      "query_duration_seconds",
				Help:      "Duration of database queries in seconds",
				Buckets:   defaultBuckets,
			},
			[]string{"operation", "table"},
		),
		DBQueryTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "database",
				Name:      "queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table"},
		),
		DBErrorsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "database",
				Name:      "errors_total",
				Help:      "Total number of database errors",
			},
			[]string{"operation", "table"},
		),

		// HTTP metrics
		HTTPRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),
		HTTPRequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds",
				Buckets:   defaultBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "trade_machine",
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "Size of HTTP responses in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
			},
			[]string{"method", "path"},
		),

		// Circuit breaker metrics
		CircuitBreakerState: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "trade_machine",
				Subsystem: "circuit_breaker",
				Name:      "state",
				Help:      "Current state of circuit breakers (0=closed, 1=half-open, 2=open)",
			},
			[]string{"service"},
		),
		CircuitBreakerTrips: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "trade_machine",
				Subsystem: "circuit_breaker",
				Name:      "trips_total",
				Help:      "Total number of circuit breaker trips",
			},
			[]string{"service"},
		),
	}

	return m
}

// InitMetrics initializes the global metrics instance
func InitMetrics() *Metrics {
	globalMetrics = NewMetrics(nil)
	return globalMetrics
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		return InitMetrics()
	}
	return globalMetrics
}

// RecordAnalysisRequest records a stock analysis request
func (m *Metrics) RecordAnalysisRequest(symbol string) {
	m.AnalysisRequestsTotal.WithLabelValues(symbol).Inc()
}

// RecordAnalysisDuration records the duration of a stock analysis
func (m *Metrics) RecordAnalysisDuration(symbol, status string, duration time.Duration) {
	m.AnalysisDuration.WithLabelValues(symbol, status).Observe(duration.Seconds())
}

// RecordAnalysisError records an analysis error
func (m *Metrics) RecordAnalysisError(symbol, errorType string) {
	m.AnalysisErrorsTotal.WithLabelValues(symbol, errorType).Inc()
}

// RecordRecommendation records a recommendation
func (m *Metrics) RecordRecommendation(action string, score, confidence float64) {
	m.RecommendationActions.WithLabelValues(action).Inc()
	m.RecommendationScores.WithLabelValues(action).Observe(score)
	m.RecommendationConfidence.WithLabelValues(action).Observe(confidence)
}

// RecordAgentDuration records the duration of an agent analysis
func (m *Metrics) RecordAgentDuration(agentType string, duration time.Duration) {
	m.AgentDuration.WithLabelValues(agentType).Observe(duration.Seconds())
}

// RecordAgentError records an agent error
func (m *Metrics) RecordAgentError(agentType, errorType string) {
	m.AgentErrorsTotal.WithLabelValues(agentType, errorType).Inc()
}

// RecordAgentScore records an agent analysis score
func (m *Metrics) RecordAgentScore(agentType string, score float64) {
	m.AgentScores.WithLabelValues(agentType).Observe(score)
}

// RecordExternalAPIRequest records an external API request
func (m *Metrics) RecordExternalAPIRequest(service, operation string) {
	m.ExternalAPIRequestsTotal.WithLabelValues(service, operation).Inc()
}

// RecordExternalAPIError records an external API error
func (m *Metrics) RecordExternalAPIError(service, operation, errorType string) {
	m.ExternalAPIErrorsTotal.WithLabelValues(service, operation, errorType).Inc()
}

// RecordExternalAPIDuration records the duration of an external API call
func (m *Metrics) RecordExternalAPIDuration(service, operation string, duration time.Duration) {
	m.ExternalAPIDuration.WithLabelValues(service, operation).Observe(duration.Seconds())
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(operation, table string, duration time.Duration) {
	m.DBQueryTotal.WithLabelValues(operation, table).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordDBError records a database error
func (m *Metrics) RecordDBError(operation, table string) {
	m.DBErrorsTotal.WithLabelValues(operation, table).Inc()
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, path, statusCode string, duration time.Duration, responseSize int) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	m.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// SetCircuitBreakerState sets the current state of a circuit breaker
func (m *Metrics) SetCircuitBreakerState(service string, state int) {
	m.CircuitBreakerState.WithLabelValues(service).Set(float64(state))
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (m *Metrics) RecordCircuitBreakerTrip(service string) {
	m.CircuitBreakerTrips.WithLabelValues(service).Inc()
}

// Timer is a helper for timing operations
type Timer struct {
	start   time.Time
	metrics *Metrics
}

// NewTimer creates a new timer
func (m *Metrics) NewTimer() *Timer {
	return &Timer{
		start:   time.Now(),
		metrics: m,
	}
}

// ObserveAnalysis records the analysis duration and status
func (t *Timer) ObserveAnalysis(symbol, status string) {
	t.metrics.RecordAnalysisDuration(symbol, status, time.Since(t.start))
}

// ObserveAgent records the agent duration
func (t *Timer) ObserveAgent(agentType string) {
	t.metrics.RecordAgentDuration(agentType, time.Since(t.start))
}

// ObserveExternalAPI records the external API duration
func (t *Timer) ObserveExternalAPI(service, operation string) {
	t.metrics.RecordExternalAPIDuration(service, operation, time.Since(t.start))
}

// ObserveDB records the database query duration
func (t *Timer) ObserveDB(operation, table string) {
	t.metrics.RecordDBQuery(operation, table, time.Since(t.start))
}

// Duration returns the elapsed time
func (t *Timer) Duration() time.Duration {
	return time.Since(t.start)
}
