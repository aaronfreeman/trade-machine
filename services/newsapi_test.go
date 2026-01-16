package services

import (
	"encoding/json"
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
