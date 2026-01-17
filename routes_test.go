package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	app := testApp(nil)
	router := testRouter(app)

	if router == nil {
		t.Fatal("expected router to be created")
	}
}

func TestCorsMiddleware(t *testing.T) {
	allowedOrigins := "http://localhost:3000"
	middleware := corsMiddleware(allowedOrigins)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("sets CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigins {
			t.Errorf("expected Access-Control-Allow-Origin %q, got %q", allowedOrigins, got)
		}

		if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Error("expected Access-Control-Allow-Methods header")
		}

		if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
			t.Error("expected Access-Control-Allow-Headers header")
		}
	})

	t.Run("handles OPTIONS preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
		}
	})

	t.Run("passes through other methods", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for GET, got %d", w.Code)
		}
	})
}

func TestRouterRoutes(t *testing.T) {
	app := testApp(nil)
	router := testRouter(app)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"GET /", http.MethodGet, "/", http.StatusOK},
		{"GET /index.html", http.MethodGet, "/index.html", http.StatusOK},
		{"GET /api/health", http.MethodGet, "/api/health", http.StatusOK},
		{"GET /api/nonexistent", http.MethodGet, "/api/nonexistent", http.StatusNotFound},
		{"POST / not allowed", http.MethodPost, "/", http.StatusMethodNotAllowed},
		{"DELETE /api/health not allowed", http.MethodDelete, "/api/health", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
