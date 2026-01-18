package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)

	// Test default status code
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code to be 200, got %d", rw.statusCode)
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code to be 404, got %d", rw.statusCode)
	}

	// Test Write
	data := []byte("Hello, World!")
	n, err := rw.Write(data)
	if err != nil {
		t.Errorf("Write returned error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if rw.responseSize != len(data) {
		t.Errorf("Expected response size to be %d, got %d", len(data), rw.responseSize)
	}

	// Test multiple writes
	n2, _ := rw.Write(data)
	if rw.responseSize != len(data)+n2 {
		t.Errorf("Expected cumulative response size to be %d, got %d", len(data)+n2, rw.responseSize)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with metrics middleware
	wrapped := MetricsMiddleware(handler)

	// Create a chi router context (required for route pattern extraction)
	r := chi.NewRouter()
	r.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsMiddleware_Error(t *testing.T) {
	// Create a handler that returns an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})

	wrapped := MetricsMiddleware(handler)

	req := httptest.NewRequest("GET", "/api/error", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestMetricsMiddleware_POST(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "123"}`))
	})

	wrapped := MetricsMiddleware(handler)

	req := httptest.NewRequest("POST", "/api/analyze", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}
