package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FMPService handles communication with Financial Modeling Prep API
type FMPService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewFMPService creates a new FMPService instance
func NewFMPService(apiKey string) *FMPService {
	return &FMPService{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://financialmodelingprep.com/api/v3",
	}
}

// fmpScreenerResponse represents a single result from the FMP stock screener API
type fmpScreenerResponse struct {
	Symbol            string  `json:"symbol"`
	CompanyName       string  `json:"companyName"`
	MarketCap         int64   `json:"marketCap"`
	Sector            string  `json:"sector"`
	Industry          string  `json:"industry"`
	Beta              float64 `json:"beta"`
	Price             float64 `json:"price"`
	LastAnnualDividend float64 `json:"lastAnnualDividend"`
	Volume            int64   `json:"volume"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Country           string  `json:"country"`
	IsEtf             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
}

// fmpProfileResponse represents a company profile from the FMP API
type fmpProfileResponse struct {
	Symbol            string  `json:"symbol"`
	CompanyName       string  `json:"companyName"`
	Price             float64 `json:"price"`
	Beta              float64 `json:"beta"`
	VolAvg            int64   `json:"volAvg"`
	MktCap            int64   `json:"mktCap"`
	LastDiv           float64 `json:"lastDiv"`
	Range             string  `json:"range"`
	Changes           float64 `json:"changes"`
	Currency          string  `json:"currency"`
	CIK               string  `json:"cik"`
	ISIN              string  `json:"isin"`
	CUSIP             string  `json:"cusip"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Industry          string  `json:"industry"`
	Website           string  `json:"website"`
	Description       string  `json:"description"`
	CEO               string  `json:"ceo"`
	Sector            string  `json:"sector"`
	Country           string  `json:"country"`
	FullTimeEmployees string  `json:"fullTimeEmployees"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	Zip               string  `json:"zip"`
	DCF               float64 `json:"dcf"`
	DCFDiff           float64 `json:"dcfDiff"`
	Image             string  `json:"image"`
	IPODate           string  `json:"ipoDate"`
	DefaultImage      bool    `json:"defaultImage"`
	IsEtf             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
	IsFund            bool    `json:"isFund"`
	IsAdr             bool    `json:"isAdr"`
}

// fmpRatiosResponse represents key ratios from the FMP API
type fmpRatiosResponse struct {
	Symbol                   string  `json:"symbol"`
	PERatio                  float64 `json:"peRatioTTM"`
	PriceToBookRatio         float64 `json:"priceToBookRatioTTM"`
	DividendYield            float64 `json:"dividendYieldTTM"`
	DividendYieldPercentage  float64 `json:"dividendYieldPercentageTTM"`
	EPS                      float64 `json:"netIncomePerShareTTM"`
}

// Screen searches for stocks matching the given criteria
func (s *FMPService) Screen(ctx context.Context, criteria ScreenCriteria) ([]ScreenerResult, error) {
	return WithCircuitBreaker(ctx, BreakerFMP, func() ([]ScreenerResult, error) {
		var results []ScreenerResult

		err := WithRetry(ctx, DefaultRetryConfig, func() error {
			params := url.Values{}
			params.Set("apikey", s.apiKey)

			if criteria.MarketCapMin > 0 {
				params.Set("marketCapMoreThan", strconv.FormatInt(criteria.MarketCapMin, 10))
			}
			if criteria.MarketCapMax > 0 {
				params.Set("marketCapLowerThan", strconv.FormatInt(criteria.MarketCapMax, 10))
			}
			if criteria.Sector != "" {
				params.Set("sector", criteria.Sector)
			}
			if criteria.Limit > 0 {
				params.Set("limit", strconv.Itoa(criteria.Limit))
			}

			// FMP screener endpoint
			reqURL := s.baseURL + "/stock-screener?" + params.Encode()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			resp, err := s.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to fetch screener results: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("screener API returned status %d", resp.StatusCode)
			}

			var screenerResp []fmpScreenerResponse
			if err := json.NewDecoder(resp.Body).Decode(&screenerResp); err != nil {
				return fmt.Errorf("failed to decode screener response: %w", err)
			}

			// Now we need to fetch ratios for each stock to get P/E, P/B, and dividend yield
			// For efficiency, we'll fetch them in batches or filter post-fetch
			// The FMP screener doesn't directly support P/E and P/B filters,
			// so we need to fetch ratios and filter client-side
			results = make([]ScreenerResult, 0, len(screenerResp))

			for _, stock := range screenerResp {
				// Skip ETFs and inactive stocks
				if stock.IsEtf || !stock.IsActivelyTrading {
					continue
				}

				result := ScreenerResult{
					Symbol:      stock.Symbol,
					CompanyName: stock.CompanyName,
					MarketCap:   stock.MarketCap,
					Sector:      stock.Sector,
					Industry:    stock.Industry,
					Price:       stock.Price,
					Beta:        stock.Beta,
					Volume:      stock.Volume,
					Exchange:    stock.ExchangeShortName,
					Country:     stock.Country,
				}

				results = append(results, result)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		// If we need to filter by P/E, P/B, or dividend yield, we need to fetch ratios
		// This is done as a second pass to avoid hitting rate limits on every screener call
		if criteria.PERatioMax > 0 || criteria.PBRatioMax > 0 || criteria.DividendYieldMin > 0 || criteria.EPSMin > 0 {
			results, err = s.enrichAndFilterResults(ctx, results, criteria)
			if err != nil {
				return nil, err
			}
		}

		return results, nil
	})
}

// enrichAndFilterResults fetches ratios for screener results and filters by P/E, P/B, etc.
func (s *FMPService) enrichAndFilterResults(ctx context.Context, results []ScreenerResult, criteria ScreenCriteria) ([]ScreenerResult, error) {
	filtered := make([]ScreenerResult, 0, len(results))

	for _, result := range results {
		ratios, err := s.getRatios(ctx, result.Symbol)
		if err != nil {
			// Skip stocks where we can't fetch ratios, but don't fail the whole operation
			continue
		}

		// Apply filters
		if criteria.PERatioMax > 0 && (ratios.PERatio <= 0 || ratios.PERatio > criteria.PERatioMax) {
			continue
		}
		if criteria.PBRatioMax > 0 && (ratios.PriceToBookRatio <= 0 || ratios.PriceToBookRatio > criteria.PBRatioMax) {
			continue
		}
		if criteria.DividendYieldMin > 0 && ratios.DividendYieldPercentage < criteria.DividendYieldMin {
			continue
		}
		if criteria.EPSMin > 0 && ratios.EPS < criteria.EPSMin {
			continue
		}

		// Enrich the result with ratio data
		result.PERatio = ratios.PERatio
		result.PBRatio = ratios.PriceToBookRatio
		result.DividendYield = ratios.DividendYieldPercentage
		result.EPS = ratios.EPS

		filtered = append(filtered, result)
	}

	return filtered, nil
}

// getRatios fetches key ratios for a symbol
func (s *FMPService) getRatios(ctx context.Context, symbol string) (*fmpRatiosResponse, error) {
	reqURL := fmt.Sprintf("%s/ratios-ttm/%s?apikey=%s", s.baseURL, url.PathEscape(symbol), s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ratios request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ratios: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ratios API returned status %d", resp.StatusCode)
	}

	var ratiosResp []fmpRatiosResponse
	if err := json.NewDecoder(resp.Body).Decode(&ratiosResp); err != nil {
		return nil, fmt.Errorf("failed to decode ratios response: %w", err)
	}

	if len(ratiosResp) == 0 {
		return nil, fmt.Errorf("no ratios data for symbol %s", symbol)
	}

	return &ratiosResp[0], nil
}

// GetCompanyProfile returns enriched company profile data for a symbol
func (s *FMPService) GetCompanyProfile(ctx context.Context, symbol string) (*CompanyProfile, error) {
	return WithCircuitBreaker(ctx, BreakerFMP, func() (*CompanyProfile, error) {
		var profile *CompanyProfile

		err := WithRetry(ctx, DefaultRetryConfig, func() error {
			reqURL := fmt.Sprintf("%s/profile/%s?apikey=%s", s.baseURL, url.PathEscape(symbol), s.apiKey)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create profile request: %w", err)
			}

			resp, err := s.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to fetch company profile: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("profile API returned status %d", resp.StatusCode)
			}

			var profileResp []fmpProfileResponse
			if err := json.NewDecoder(resp.Body).Decode(&profileResp); err != nil {
				return fmt.Errorf("failed to decode profile response: %w", err)
			}

			if len(profileResp) == 0 {
				return fmt.Errorf("no profile data for symbol %s", symbol)
			}

			p := profileResp[0]
			profile = &CompanyProfile{
				Symbol:            p.Symbol,
				CompanyName:       p.CompanyName,
				Price:             p.Price,
				MarketCap:         p.MktCap,
				Sector:            p.Sector,
				Industry:          p.Industry,
				Description:       p.Description,
				CEO:               p.CEO,
				Website:           p.Website,
				Exchange:          p.ExchangeShortName,
				Country:           p.Country,
				Beta:              p.Beta,
				VolAvg:            p.VolAvg,
				LastDividend:      p.LastDiv,
				Range52Week:       p.Range,
				Changes:           p.Changes,
				DCF:               p.DCF,
				IPODate:           p.IPODate,
				IsActivelyTrading: p.IsActivelyTrading,
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		return profile, nil
	})
}

// Compile-time interface verification
var _ FMPServiceInterface = (*FMPService)(nil)
