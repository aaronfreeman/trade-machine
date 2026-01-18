package api

import (
	"net/http"
	"strconv"
	"time"

	"trade-machine/observability"

	"github.com/go-chi/chi/v5"
)

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // default status code
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += size
	return size, err
}

// MetricsMiddleware records HTTP metrics for each request
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code and size
		wrapped := newResponseWriter(w)

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Get the route pattern from chi
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		// Record metrics
		metrics := observability.GetMetrics()
		duration := time.Since(start)
		statusCode := strconv.Itoa(wrapped.statusCode)

		metrics.RecordHTTPRequest(r.Method, routePattern, statusCode, duration, wrapped.responseSize)
	})
}
