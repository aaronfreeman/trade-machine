package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewNewsAPIService(t *testing.T) {
	service := NewNewsAPIService("test-api-key")
	if service == nil {
		t.Error("NewNewsAPIService should not return nil")
	}
	if service.apiKey != "test-api-key" {
		t.Errorf("apiKey = %v, want 'test-api-key'", service.apiKey)
	}
	if service.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if service.baseURL != "https://newsapi.org/v2" {
		t.Errorf("baseURL = %v, want 'https://newsapi.org/v2'", service.baseURL)
	}
}

func TestNewsAPIResponse_Deserialization(t *testing.T) {
	jsonResponse := `{
		"status": "ok",
		"totalResults": 100,
		"articles": [
			{
				"source": {
					"id": "techcrunch",
					"name": "TechCrunch"
				},
				"author": "Sarah Perez",
				"title": "Apple Stock Rises on Strong Earnings",
				"description": "Apple Inc reported better-than-expected earnings...",
				"url": "https://techcrunch.com/apple-earnings",
				"urlToImage": "https://example.com/image.jpg",
				"publishedAt": "2024-01-15T14:30:00Z",
				"content": "Full article content here..."
			},
			{
				"source": {
					"id": null,
					"name": "Reuters"
				},
				"author": "John Smith",
				"title": "Tech Stocks Rally",
				"description": "Technology stocks rallied on Wednesday...",
				"url": "https://reuters.com/tech-rally",
				"urlToImage": "https://example.com/image2.jpg",
				"publishedAt": "2024-01-15T10:00:00Z",
				"content": "Another article content..."
			}
		]
	}`

	var resp NewsAPIResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal NewsAPIResponse: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %v, want 'ok'", resp.Status)
	}
	if resp.TotalResults != 100 {
		t.Errorf("TotalResults = %v, want 100", resp.TotalResults)
	}
	if len(resp.Articles) != 2 {
		t.Errorf("Articles length = %v, want 2", len(resp.Articles))
	}

	// Check first article
	article := resp.Articles[0]
	if article.Title != "Apple Stock Rises on Strong Earnings" {
		t.Errorf("Article[0].Title = %v, want 'Apple Stock Rises on Strong Earnings'", article.Title)
	}
	if article.Source.Name != "TechCrunch" {
		t.Errorf("Article[0].Source.Name = %v, want 'TechCrunch'", article.Source.Name)
	}
	if article.Author != "Sarah Perez" {
		t.Errorf("Article[0].Author = %v, want 'Sarah Perez'", article.Author)
	}
	if article.URL != "https://techcrunch.com/apple-earnings" {
		t.Errorf("Article[0].URL = %v, want 'https://techcrunch.com/apple-earnings'", article.URL)
	}
	if article.URLToImage != "https://example.com/image.jpg" {
		t.Errorf("Article[0].URLToImage = %v, want 'https://example.com/image.jpg'", article.URLToImage)
	}
}

func TestNewsAPIResponse_EmptyArticles(t *testing.T) {
	jsonResponse := `{
		"status": "ok",
		"totalResults": 0,
		"articles": []
	}`

	var resp NewsAPIResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal NewsAPIResponse: %v", err)
	}

	if resp.TotalResults != 0 {
		t.Errorf("TotalResults = %v, want 0", resp.TotalResults)
	}
	if len(resp.Articles) != 0 {
		t.Errorf("Articles length = %v, want 0", len(resp.Articles))
	}
}

func TestNewsAPIResponse_NullSource(t *testing.T) {
	// Test handling of null source ID (common in API responses)
	jsonResponse := `{
		"status": "ok",
		"totalResults": 1,
		"articles": [
			{
				"source": {
					"id": null,
					"name": "Unknown Source"
				},
				"author": null,
				"title": "Test Article",
				"description": "Test description",
				"url": "https://example.com",
				"urlToImage": null,
				"publishedAt": "2024-01-15T00:00:00Z",
				"content": null
			}
		]
	}`

	var resp NewsAPIResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.Articles[0].Source.Name != "Unknown Source" {
		t.Errorf("Source.Name = %v, want 'Unknown Source'", resp.Articles[0].Source.Name)
	}
	// Null author should be empty string
	if resp.Articles[0].Author != "" {
		t.Errorf("Author = %v, want empty string for null", resp.Articles[0].Author)
	}
}

func TestNewsAPIResponse_ErrorStatus(t *testing.T) {
	jsonResponse := `{
		"status": "error",
		"code": "apiKeyInvalid",
		"message": "Your API key is invalid or incorrect."
	}`

	var resp NewsAPIResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if resp.Status != "error" {
		t.Errorf("Status = %v, want 'error'", resp.Status)
	}
}

func TestNewsAPIResponse_MultipleArticles(t *testing.T) {
	jsonResponse := `{
		"status": "ok",
		"totalResults": 5,
		"articles": [
			{
				"source": {"id": "source1", "name": "Source 1"},
				"author": "Author 1",
				"title": "Title 1",
				"description": "Description 1",
				"url": "https://example.com/1",
				"urlToImage": "https://example.com/img1.jpg",
				"publishedAt": "2024-01-15T10:00:00Z",
				"content": "Content 1"
			},
			{
				"source": {"id": "source2", "name": "Source 2"},
				"author": "Author 2",
				"title": "Title 2",
				"description": "Description 2",
				"url": "https://example.com/2",
				"urlToImage": "https://example.com/img2.jpg",
				"publishedAt": "2024-01-15T11:00:00Z",
				"content": "Content 2"
			},
			{
				"source": {"id": "source3", "name": "Source 3"},
				"author": "Author 3",
				"title": "Title 3",
				"description": "Description 3",
				"url": "https://example.com/3",
				"urlToImage": "https://example.com/img3.jpg",
				"publishedAt": "2024-01-15T12:00:00Z",
				"content": "Content 3"
			}
		]
	}`

	var resp NewsAPIResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if resp.TotalResults != 5 {
		t.Errorf("TotalResults = %v, want 5", resp.TotalResults)
	}
	if len(resp.Articles) != 3 {
		t.Errorf("Articles length = %v, want 3", len(resp.Articles))
	}

	for i, article := range resp.Articles {
		if article.Source.ID != fmt.Sprintf("source%d", i+1) {
			t.Errorf("Article[%d].Source.ID = %v, want 'source%d'", i, article.Source.ID, i+1)
		}
	}
}

func TestGetNews_LimitValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewNewsAPIService("invalid-api-key")
	ctx := context.Background()

	tests := []struct {
		name  string
		limit int
	}{
		{"Zero limit (defaults to 10)", 0},
		{"Negative limit (defaults to 10)", -5},
		{"Valid limit", 25},
		{"Over max limit (caps at 100)", 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetNews(ctx, "AAPL", tt.limit)
			// We expect an error due to invalid API key, but the function should not panic
			if err == nil {
				t.Error("GetNews should return error with invalid API key")
			}
		})
	}
}

func TestGetHeadlines_LimitValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewNewsAPIService("invalid-api-key")
	ctx := context.Background()

	tests := []struct {
		name  string
		limit int
	}{
		{"Zero limit (defaults to 10)", 0},
		{"Negative limit (defaults to 10)", -5},
		{"Valid limit", 15},
		{"Over max limit (caps at 100)", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetHeadlines(ctx, "Tesla", tt.limit)
			// We expect an error due to invalid API key
			if err == nil {
				t.Error("GetHeadlines should return error with invalid API key")
			}
		})
	}
}

func TestGetNews_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewNewsAPIService("test-api-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetNews(ctx, "AAPL", 10)
	if err == nil {
		t.Error("GetNews should return error when context is cancelled")
	}
}

func TestGetHeadlines_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewNewsAPIService("test-api-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetHeadlines(ctx, "AAPL", 10)
	if err == nil {
		t.Error("GetHeadlines should return error when context is cancelled")
	}
}

func TestNewNewsAPIService_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{"Valid API key", "test-api-key-123"},
		{"Empty API key", ""},
		{"Long API key", "very-long-api-key-" + string(make([]byte, 100))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewNewsAPIService(tt.apiKey)
			if service == nil {
				t.Error("NewNewsAPIService should not return nil")
			}
			if service.httpClient == nil {
				t.Error("httpClient should not be nil")
			}
			if service.baseURL != "https://newsapi.org/v2" {
				t.Errorf("baseURL = %v, want 'https://newsapi.org/v2'", service.baseURL)
			}
		})
	}
}

func TestGetNews_WithMockServer(t *testing.T) {
	// Reset circuit breaker for test isolation
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Path != "/everything" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing or wrong API key header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "ok",
			"totalResults": 2,
			"articles": [
				{
					"source": {"id": "src1", "name": "Source One"},
					"author": "Author One",
					"title": "Test Article 1",
					"description": "Description 1",
					"url": "https://example.com/1",
					"urlToImage": "https://example.com/img1.jpg",
					"publishedAt": "2024-01-15T10:00:00Z",
					"content": "Content 1"
				},
				{
					"source": {"id": "src2", "name": "Source Two"},
					"author": "Author Two",
					"title": "Test Article 2",
					"description": "Description 2",
					"url": "https://example.com/2",
					"urlToImage": "https://example.com/img2.jpg",
					"publishedAt": "2024-01-15T11:00:00Z",
					"content": "Content 2"
				}
			]
		}`))
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	articles, err := service.GetNews(ctx, "AAPL", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(articles))
	}
	if articles[0].Title != "Test Article 1" {
		t.Errorf("unexpected title: %s", articles[0].Title)
	}
	if articles[0].Source != "Source One" {
		t.Errorf("unexpected source: %s", articles[0].Source)
	}
}

func TestGetNews_InvalidTimestamp(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "ok",
			"totalResults": 1,
			"articles": [
				{
					"source": {"id": "src1", "name": "Source"},
					"author": "Author",
					"title": "Test",
					"description": "Desc",
					"url": "https://example.com",
					"urlToImage": "https://example.com/img.jpg",
					"publishedAt": "invalid-timestamp",
					"content": "Content"
				}
			]
		}`))
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	articles, err := service.GetNews(ctx, "AAPL", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(articles))
	}
	// Invalid timestamp should default to current time (non-zero)
	if articles[0].PublishedAt.IsZero() {
		t.Error("expected non-zero timestamp for invalid input")
	}
}

func TestGetNews_NonOKStatus(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetNews(ctx, "AAPL", 10)

	if err == nil {
		t.Error("expected error for non-OK status")
	}
}

func TestGetNews_InvalidJSON(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetNews(ctx, "AAPL", 10)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetHeadlines_WithMockServer(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/top-headlines" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "ok",
			"totalResults": 1,
			"articles": [
				{
					"source": {"id": "bbc", "name": "BBC News"},
					"author": "BBC Reporter",
					"title": "Breaking News",
					"description": "Important news",
					"url": "https://bbc.com/news",
					"urlToImage": "https://bbc.com/img.jpg",
					"publishedAt": "2024-01-15T12:00:00Z",
					"content": "Full story"
				}
			]
		}`))
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	articles, err := service.GetHeadlines(ctx, "Tesla", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(articles))
	}
	if articles[0].Source != "BBC News" {
		t.Errorf("unexpected source: %s", articles[0].Source)
	}
}

func TestGetHeadlines_InvalidTimestamp(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "ok",
			"totalResults": 1,
			"articles": [
				{
					"source": {"id": "src", "name": "Source"},
					"author": "Author",
					"title": "Title",
					"description": "Desc",
					"url": "https://example.com",
					"urlToImage": "",
					"publishedAt": "bad-date",
					"content": ""
				}
			]
		}`))
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	articles, err := service.GetHeadlines(ctx, "AAPL", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if articles[0].PublishedAt.IsZero() {
		t.Error("expected fallback timestamp for invalid date")
	}
}

func TestGetHeadlines_NonOKStatus(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetHeadlines(ctx, "AAPL", 10)

	if err == nil {
		t.Error("expected error for non-OK status")
	}
}

func TestGetHeadlines_InvalidJSON(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	service := NewNewsAPIService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetHeadlines(ctx, "AAPL", 10)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
