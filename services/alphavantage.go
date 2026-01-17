package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"trade-machine/models"

	"github.com/shopspring/decimal"
)

// AlphaVantageService handles communication with Alpha Vantage API
type AlphaVantageService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewAlphaVantageService creates a new AlphaVantageService instance
func NewAlphaVantageService(apiKey string) *AlphaVantageService {
	return &AlphaVantageService{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://www.alphavantage.co/query",
	}
}

// OverviewResponse represents the company overview response from Alpha Vantage
type OverviewResponse struct {
	Symbol           string `json:"Symbol"`
	Name             string `json:"Name"`
	Description      string `json:"Description"`
	Exchange         string `json:"Exchange"`
	Currency         string `json:"Currency"`
	Country          string `json:"Country"`
	Sector           string `json:"Sector"`
	Industry         string `json:"Industry"`
	MarketCap        string `json:"MarketCapitalization"`
	PERatio          string `json:"PERatio"`
	PEGRatio         string `json:"PEGRatio"`
	BookValue        string `json:"BookValue"`
	DividendPerShare string `json:"DividendPerShare"`
	DividendYield    string `json:"DividendYield"`
	EPS              string `json:"EPS"`
	RevenuePerShare  string `json:"RevenuePerShareTTM"`
	ProfitMargin     string `json:"ProfitMargin"`
	Beta             string `json:"Beta"`
	Week52High       string `json:"52WeekHigh"`
	Week52Low        string `json:"52WeekLow"`
	AnalystTarget    string `json:"AnalystTargetPrice"`
}

// GetFundamentals returns fundamental data for a symbol
func (s *AlphaVantageService) GetFundamentals(ctx context.Context, symbol string) (*models.Fundamentals, error) {
	var fundamentals *models.Fundamentals

	err := WithRetry(ctx, DefaultRetryConfig, func() error {
		params := url.Values{}
		params.Set("function", "OVERVIEW")
		params.Set("symbol", symbol)
		params.Set("apikey", s.apiKey)

		resp, err := s.httpClient.Get(s.baseURL + "?" + params.Encode())
		if err != nil {
			return fmt.Errorf("failed to fetch overview: %w", err)
		}
		defer resp.Body.Close()

		var overview OverviewResponse
		if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
			return fmt.Errorf("failed to decode overview: %w", err)
		}

		marketCap, _ := decimal.NewFromString(overview.MarketCap)
		eps, _ := decimal.NewFromString(overview.EPS)
		week52High, _ := decimal.NewFromString(overview.Week52High)
		week52Low, _ := decimal.NewFromString(overview.Week52Low)

		var peRatio, dividendYield, beta float64
		if overview.PERatio != "" && overview.PERatio != "None" {
			peRatio, err = strconv.ParseFloat(overview.PERatio, 64)
			if err != nil {
				log.Printf("Warning: failed to parse P/E ratio '%s': %v", overview.PERatio, err)
			}
		}
		if overview.DividendYield != "" && overview.DividendYield != "None" {
			dividendYield, err = strconv.ParseFloat(overview.DividendYield, 64)
			if err != nil {
				log.Printf("Warning: failed to parse dividend yield '%s': %v", overview.DividendYield, err)
			}
		}
		if overview.Beta != "" && overview.Beta != "None" {
			beta, err = strconv.ParseFloat(overview.Beta, 64)
			if err != nil {
				log.Printf("Warning: failed to parse beta '%s': %v", overview.Beta, err)
			}
		}

		fundamentals = &models.Fundamentals{
			Symbol:        symbol,
			MarketCap:     marketCap,
			PERatio:       peRatio,
			EPS:           eps,
			DividendYield: dividendYield,
			Week52High:    week52High,
			Week52Low:     week52Low,
			Beta:          beta,
			UpdatedAt:     time.Now(),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fundamentals, nil
}

// NewsResponse represents the news response from Alpha Vantage
type NewsResponse struct {
	Items string `json:"items"`
	Feed  []struct {
		Title            string   `json:"title"`
		URL              string   `json:"url"`
		Summary          string   `json:"summary"`
		Source           string   `json:"source"`
		TimePublished    string   `json:"time_published"`
		Authors          []string `json:"authors"`
		OverallSentiment string   `json:"overall_sentiment_label"`
		SentimentScore   float64  `json:"overall_sentiment_score"`
	} `json:"feed"`
}

// GetNews returns recent news for a symbol
func (s *AlphaVantageService) GetNews(ctx context.Context, symbol string) ([]models.NewsArticle, error) {
	params := url.Values{}
	params.Set("function", "NEWS_SENTIMENT")
	params.Set("tickers", symbol)
	params.Set("limit", "10")
	params.Set("apikey", s.apiKey)

	resp, err := s.httpClient.Get(s.baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	var newsResp NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return nil, fmt.Errorf("failed to decode news: %w", err)
	}

	articles := make([]models.NewsArticle, 0, len(newsResp.Feed))
	for _, item := range newsResp.Feed {
		publishedAt, err := time.Parse("20060102T150405", item.TimePublished)
		if err != nil {
			log.Printf("Warning: failed to parse timestamp '%s': %v, using current time", item.TimePublished, err)
			publishedAt = time.Now()
		}

		author := ""
		if len(item.Authors) > 0 {
			author = item.Authors[0]
		}

		articles = append(articles, models.NewsArticle{
			Title:       item.Title,
			Description: item.Summary,
			URL:         item.URL,
			Source:      item.Source,
			Author:      author,
			PublishedAt: publishedAt,
		})
	}

	return articles, nil
}

// QuoteResponse represents a quote from Alpha Vantage
type QuoteResponse struct {
	GlobalQuote struct {
		Symbol        string `json:"01. symbol"`
		Open          string `json:"02. open"`
		High          string `json:"03. high"`
		Low           string `json:"04. low"`
		Price         string `json:"05. price"`
		Volume        string `json:"06. volume"`
		LatestDay     string `json:"07. latest trading day"`
		PrevClose     string `json:"08. previous close"`
		Change        string `json:"09. change"`
		ChangePercent string `json:"10. change percent"`
	} `json:"Global Quote"`
}

// GetQuote returns the latest quote for a symbol
func (s *AlphaVantageService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	params := url.Values{}
	params.Set("function", "GLOBAL_QUOTE")
	params.Set("symbol", symbol)
	params.Set("apikey", s.apiKey)

	resp, err := s.httpClient.Get(s.baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}
	defer resp.Body.Close()

	var quoteResp QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResp); err != nil {
		return nil, fmt.Errorf("failed to decode quote: %w", err)
	}

	price, _ := decimal.NewFromString(quoteResp.GlobalQuote.Price)
	var volume int64
	if quoteResp.GlobalQuote.Volume != "" {
		volume, err = strconv.ParseInt(quoteResp.GlobalQuote.Volume, 10, 64)
		if err != nil {
			log.Printf("Warning: failed to parse volume '%s': %v", quoteResp.GlobalQuote.Volume, err)
		}
	}

	return &models.Quote{
		Symbol:    symbol,
		Last:      price,
		Volume:    volume,
		Timestamp: time.Now(),
	}, nil
}
