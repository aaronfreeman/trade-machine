package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"

	"trade-machine/observability"
)

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	MaxRequests uint32        // max requests allowed in half-open state
	Interval    time.Duration // cyclic period of the closed state to clear counts
	Timeout     time.Duration // period of the open state before transitioning to half-open
}

// DefaultCircuitBreakerConfig returns sensible defaults per the issue spec
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	MaxRequests: 5,
	Interval:    1 * time.Minute,
	Timeout:     30 * time.Second,
}

// CircuitBreakerRegistry manages circuit breakers for different services
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*gobreaker.CircuitBreaker[any]
	config   CircuitBreakerConfig
}

// NewCircuitBreakerRegistry creates a new registry with the given config
func NewCircuitBreakerRegistry(config CircuitBreakerConfig) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*gobreaker.CircuitBreaker[any]),
		config:   config,
	}
}

// GetBreaker returns (or creates) a circuit breaker for the given service name
func (r *CircuitBreakerRegistry) GetBreaker(name string) *gobreaker.CircuitBreaker[any] {
	r.mu.RLock()
	cb, exists := r.breakers[name]
	r.mu.RUnlock()

	if exists {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = r.breakers[name]; exists {
		return cb
	}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: r.config.MaxRequests,
		Interval:    r.config.Interval,
		Timeout:     r.config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the breaker if failure ratio exceeds 50% with at least 5 requests
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			observability.Warn("circuit breaker state change",
				"breaker", name,
				"from", from.String(),
				"to", to.String())

			// Record metrics for circuit breaker state changes
			metrics := observability.GetMetrics()
			metrics.SetCircuitBreakerState(name, stateToInt(to))
			if to == gobreaker.StateOpen {
				metrics.RecordCircuitBreakerTrip(name)
			}
		},
	}

	cb = gobreaker.NewCircuitBreaker[any](settings)
	r.breakers[name] = cb

	return cb
}

// Execute runs the given function through the named circuit breaker
func (r *CircuitBreakerRegistry) Execute(ctx context.Context, name string, fn func() (any, error)) (any, error) {
	cb := r.GetBreaker(name)

	result, err := cb.Execute(func() (any, error) {
		// Check context before executing
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return fn()
	})

	if err != nil {
		// Check if it's a circuit breaker open error
		if err == gobreaker.ErrOpenState {
			observability.Warn("circuit breaker open, rejecting request",
				"breaker", name)
			return nil, fmt.Errorf("service %s unavailable: circuit breaker open", name)
		}
		if err == gobreaker.ErrTooManyRequests {
			observability.Warn("circuit breaker half-open, too many requests",
				"breaker", name)
			return nil, fmt.Errorf("service %s unavailable: too many requests in half-open state", name)
		}
	}

	return result, err
}

// Status returns the current state of all circuit breakers
func (r *CircuitBreakerRegistry) Status() map[string]CircuitBreakerStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]CircuitBreakerStatus)
	for name, cb := range r.breakers {
		counts := cb.Counts()
		status[name] = CircuitBreakerStatus{
			Name:             name,
			State:            cb.State().String(),
			Requests:         counts.Requests,
			TotalSuccesses:   counts.TotalSuccesses,
			TotalFailures:    counts.TotalFailures,
			ConsecutiveSucc:  counts.ConsecutiveSuccesses,
			ConsecutiveFails: counts.ConsecutiveFailures,
		}
	}
	return status
}

// CircuitBreakerStatus represents the current state of a circuit breaker
type CircuitBreakerStatus struct {
	Name             string `json:"name"`
	State            string `json:"state"`
	Requests         uint32 `json:"requests"`
	TotalSuccesses   uint32 `json:"total_successes"`
	TotalFailures    uint32 `json:"total_failures"`
	ConsecutiveSucc  uint32 `json:"consecutive_successes"`
	ConsecutiveFails uint32 `json:"consecutive_failures"`
}

// Global registry instance (can be overridden for testing)
var globalRegistry *CircuitBreakerRegistry
var registryOnce sync.Once

// GetGlobalRegistry returns the global circuit breaker registry
func GetGlobalRegistry() *CircuitBreakerRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig)
	})
	return globalRegistry
}

// SetGlobalRegistry allows overriding the global registry (useful for testing)
func SetGlobalRegistry(r *CircuitBreakerRegistry) {
	globalRegistry = r
}

// WithCircuitBreaker wraps a function call with circuit breaker protection
func WithCircuitBreaker[T any](ctx context.Context, name string, fn func() (T, error)) (T, error) {
	registry := GetGlobalRegistry()

	result, err := registry.Execute(ctx, name, func() (any, error) {
		return fn()
	})

	if err != nil {
		var zero T
		return zero, err
	}

	return result.(T), nil
}

// Circuit breaker names for external services
const (
	BreakerAlphaVantage = "alphavantage"
	BreakerNewsAPI      = "newsapi"
	BreakerAlpaca       = "alpaca"
	BreakerBedrock      = "bedrock"
	BreakerFMP          = "fmp"
)

// stateToInt converts a circuit breaker state to an integer for metrics
// 0=closed, 1=half-open, 2=open
func stateToInt(state gobreaker.State) int {
	switch state {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateHalfOpen:
		return 1
	case gobreaker.StateOpen:
		return 2
	default:
		return -1
	}
}
