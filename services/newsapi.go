package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"trade-machine/models"
)

// NewsAPIService handles communication with NewsAPI.org
type NewsAPIService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewNewsAPIService creates a new NewsAPIService instance
func NewNewsAPIService(apiKey string) *NewsAPIService {
	return &NewsAPIService{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://newsapi.org/v2",
	}
}

// NewsAPIResponse represents the response from NewsAPI
type NewsAPIResponse struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Articles     []struct {
		Source struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"source"`
		Author      string `json:"author"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		URLToImage  string `json:"urlToImage"`
		PublishedAt string `json:"publishedAt"`
		Content     string `json:"content"`
	} `json:"articles"`
}

// GetNews returns news articles for a query (typically a stock symbol or company name)
func (s *NewsAPIService) GetNews(ctx context.Context, query string, limit int) ([]models.NewsArticle, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var articles []models.NewsArticle
	err := WithRetry(ctx, DefaultRetryConfig, func() error {
		params := url.Values{}
		params.Set("q", query)
		params.Set("language", "en")
		params.Set("sortBy", "publishedAt")
		params.Set("pageSize", fmt.Sprintf("%d", limit))

		req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/everything?"+params.Encode(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("X-Api-Key", s.apiKey)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch news: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("NewsAPI returned status %d", resp.StatusCode)
		}

		var newsResp NewsAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		articles = make([]models.NewsArticle, 0, len(newsResp.Articles))
		for _, item := range newsResp.Articles {
			publishedAt, err := time.Parse(time.RFC3339, item.PublishedAt)
			if err != nil {
				log.Printf("Warning: failed to parse timestamp '%s': %v, using current time", item.PublishedAt, err)
				publishedAt = time.Now()
			}

			articles = append(articles, models.NewsArticle{
				Title:       item.Title,
				Description: item.Description,
				URL:         item.URL,
				Source:      item.Source.Name,
				Author:      item.Author,
				ImageURL:    item.URLToImage,
				PublishedAt: publishedAt,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return articles, nil
}

// GetHeadlines returns top headlines mentioning a company or stock
func (s *NewsAPIService) GetHeadlines(ctx context.Context, query string, limit int) ([]models.NewsArticle, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("country", "us")
	params.Set("category", "business")
	params.Set("pageSize", fmt.Sprintf("%d", limit))

	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/top-headlines?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Api-Key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch headlines: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NewsAPI returned status %d", resp.StatusCode)
	}

	var newsResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	articles := make([]models.NewsArticle, 0, len(newsResp.Articles))
	for _, item := range newsResp.Articles {
		publishedAt, err := time.Parse(time.RFC3339, item.PublishedAt)
		if err != nil {
			log.Printf("Warning: failed to parse timestamp '%s': %v, using current time", item.PublishedAt, err)
			publishedAt = time.Now()
		}

		articles = append(articles, models.NewsArticle{
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			Source:      item.Source.Name,
			Author:      item.Author,
			ImageURL:    item.URLToImage,
			PublishedAt: publishedAt,
		})
	}

	return articles, nil
}
