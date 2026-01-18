package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewFMPService(t *testing.T) {
	service := NewFMPService("test-api-key")
	if service == nil {
		t.Error("NewFMPService should not return nil")
	}
	if service.apiKey != "test-api-key" {
		t.Errorf("apiKey = %v, want 'test-api-key'", service.apiKey)
	}
	if service.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if service.baseURL != "https://financialmodelingprep.com/api/v3" {
		t.Errorf("baseURL = %v, want 'https://financialmodelingprep.com/api/v3'", service.baseURL)
	}
}

func TestNewFMPService_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{"Valid API key", "test-api-key-123"},
		{"Empty API key", ""},
		{"Long API key", "very-long-api-key-with-many-characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewFMPService(tt.apiKey)
			if service == nil {
				t.Error("NewFMPService should not return nil")
			}
			if service.httpClient == nil {
				t.Error("httpClient should not be nil")
			}
		})
	}
}

func TestFMPScreenerResponse_Deserialization(t *testing.T) {
	jsonResponse := `[
		{
			"symbol": "AAPL",
			"companyName": "Apple Inc.",
			"marketCap": 2500000000000,
			"sector": "Technology",
			"industry": "Consumer Electronics",
			"beta": 1.25,
			"price": 175.50,
			"lastAnnualDividend": 0.96,
			"volume": 50000000,
			"exchange": "NASDAQ",
			"exchangeShortName": "NASDAQ",
			"country": "US",
			"isEtf": false,
			"isActivelyTrading": true
		},
		{
			"symbol": "MSFT",
			"companyName": "Microsoft Corporation",
			"marketCap": 2800000000000,
			"sector": "Technology",
			"industry": "Software - Infrastructure",
			"beta": 0.95,
			"price": 380.25,
			"lastAnnualDividend": 2.72,
			"volume": 25000000,
			"exchange": "NASDAQ",
			"exchangeShortName": "NASDAQ",
			"country": "US",
			"isEtf": false,
			"isActivelyTrading": true
		}
	]`

	var resp []fmpScreenerResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal fmpScreenerResponse: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Response length = %v, want 2", len(resp))
	}

	// Check first stock
	stock := resp[0]
	if stock.Symbol != "AAPL" {
		t.Errorf("Stock[0].Symbol = %v, want 'AAPL'", stock.Symbol)
	}
	if stock.CompanyName != "Apple Inc." {
		t.Errorf("Stock[0].CompanyName = %v, want 'Apple Inc.'", stock.CompanyName)
	}
	if stock.MarketCap != 2500000000000 {
		t.Errorf("Stock[0].MarketCap = %v, want 2500000000000", stock.MarketCap)
	}
	if stock.Sector != "Technology" {
		t.Errorf("Stock[0].Sector = %v, want 'Technology'", stock.Sector)
	}
	if stock.IsEtf {
		t.Error("Stock[0].IsEtf should be false")
	}
	if !stock.IsActivelyTrading {
		t.Error("Stock[0].IsActivelyTrading should be true")
	}
}

func TestFMPProfileResponse_Deserialization(t *testing.T) {
	jsonResponse := `[
		{
			"symbol": "AAPL",
			"companyName": "Apple Inc.",
			"price": 175.50,
			"beta": 1.25,
			"volAvg": 50000000,
			"mktCap": 2500000000000,
			"lastDiv": 0.24,
			"range": "140.82-199.62",
			"changes": 2.35,
			"currency": "USD",
			"cik": "0000320193",
			"isin": "US0378331005",
			"cusip": "037833100",
			"exchange": "NASDAQ Global Select",
			"exchangeShortName": "NASDAQ",
			"industry": "Consumer Electronics",
			"website": "https://www.apple.com",
			"description": "Apple Inc. designs, manufactures, and markets smartphones...",
			"ceo": "Tim Cook",
			"sector": "Technology",
			"country": "US",
			"fullTimeEmployees": "164000",
			"phone": "14089961010",
			"address": "One Apple Park Way",
			"city": "Cupertino",
			"state": "CA",
			"zip": "95014",
			"dcf": 150.25,
			"dcfDiff": 25.25,
			"image": "https://example.com/aapl.png",
			"ipoDate": "1980-12-12",
			"defaultImage": false,
			"isEtf": false,
			"isActivelyTrading": true,
			"isFund": false,
			"isAdr": false
		}
	]`

	var resp []fmpProfileResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal fmpProfileResponse: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("Response length = %v, want 1", len(resp))
	}

	profile := resp[0]
	if profile.Symbol != "AAPL" {
		t.Errorf("Profile.Symbol = %v, want 'AAPL'", profile.Symbol)
	}
	if profile.CompanyName != "Apple Inc." {
		t.Errorf("Profile.CompanyName = %v, want 'Apple Inc.'", profile.CompanyName)
	}
	if profile.CEO != "Tim Cook" {
		t.Errorf("Profile.CEO = %v, want 'Tim Cook'", profile.CEO)
	}
	if profile.Sector != "Technology" {
		t.Errorf("Profile.Sector = %v, want 'Technology'", profile.Sector)
	}
	if profile.IPODate != "1980-12-12" {
		t.Errorf("Profile.IPODate = %v, want '1980-12-12'", profile.IPODate)
	}
	if !profile.IsActivelyTrading {
		t.Error("Profile.IsActivelyTrading should be true")
	}
}

func TestFMPRatiosResponse_Deserialization(t *testing.T) {
	jsonResponse := `[
		{
			"symbol": "AAPL",
			"peRatioTTM": 28.5,
			"priceToBookRatioTTM": 45.2,
			"dividendYieldTTM": 0.0055,
			"dividendYieldPercentageTTM": 0.55,
			"netIncomePerShareTTM": 6.15
		}
	]`

	var resp []fmpRatiosResponse
	if err := json.Unmarshal([]byte(jsonResponse), &resp); err != nil {
		t.Fatalf("Failed to unmarshal fmpRatiosResponse: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("Response length = %v, want 1", len(resp))
	}

	ratios := resp[0]
	if ratios.Symbol != "AAPL" {
		t.Errorf("Ratios.Symbol = %v, want 'AAPL'", ratios.Symbol)
	}
	if ratios.PERatio != 28.5 {
		t.Errorf("Ratios.PERatio = %v, want 28.5", ratios.PERatio)
	}
	if ratios.PriceToBookRatio != 45.2 {
		t.Errorf("Ratios.PriceToBookRatio = %v, want 45.2", ratios.PriceToBookRatio)
	}
	if ratios.DividendYieldPercentage != 0.55 {
		t.Errorf("Ratios.DividendYieldPercentage = %v, want 0.55", ratios.DividendYieldPercentage)
	}
	if ratios.EPS != 6.15 {
		t.Errorf("Ratios.EPS = %v, want 6.15", ratios.EPS)
	}
}

func TestScreen_WithMockServer(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stock-screener" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("apikey") != "test-key" {
			t.Error("missing or wrong API key")
		}
		if query.Get("marketCapMoreThan") != "1000000000" {
			t.Errorf("marketCapMoreThan = %s, want 1000000000", query.Get("marketCapMoreThan"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"symbol": "AAPL",
				"companyName": "Apple Inc.",
				"marketCap": 2500000000000,
				"sector": "Technology",
				"industry": "Consumer Electronics",
				"beta": 1.25,
				"price": 175.50,
				"lastAnnualDividend": 0.96,
				"volume": 50000000,
				"exchange": "NASDAQ",
				"exchangeShortName": "NASDAQ",
				"country": "US",
				"isEtf": false,
				"isActivelyTrading": true
			}
		]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		MarketCapMin: 1_000_000_000,
		Limit:        10,
	}

	results, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Symbol != "AAPL" {
		t.Errorf("unexpected symbol: %s", results[0].Symbol)
	}
	if results[0].MarketCap != 2500000000000 {
		t.Errorf("unexpected market cap: %d", results[0].MarketCap)
	}
}

func TestScreen_WithSectorFilter(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("sector") != "Technology" {
			t.Errorf("sector = %s, want Technology", query.Get("sector"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		Sector: "Technology",
	}

	_, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScreen_FiltersEtfsAndInactiveStocks(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"symbol": "AAPL",
				"companyName": "Apple Inc.",
				"marketCap": 2500000000000,
				"sector": "Technology",
				"industry": "Consumer Electronics",
				"beta": 1.25,
				"price": 175.50,
				"volume": 50000000,
				"exchangeShortName": "NASDAQ",
				"country": "US",
				"isEtf": false,
				"isActivelyTrading": true
			},
			{
				"symbol": "SPY",
				"companyName": "SPDR S&P 500 ETF",
				"marketCap": 400000000000,
				"sector": "",
				"industry": "",
				"beta": 1.0,
				"price": 450.00,
				"volume": 100000000,
				"exchangeShortName": "NYSE",
				"country": "US",
				"isEtf": true,
				"isActivelyTrading": true
			},
			{
				"symbol": "DEAD",
				"companyName": "Dead Company",
				"marketCap": 0,
				"sector": "Technology",
				"industry": "Software",
				"beta": 0,
				"price": 0,
				"volume": 0,
				"exchangeShortName": "OTC",
				"country": "US",
				"isEtf": false,
				"isActivelyTrading": false
			}
		]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{}

	results, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (AAPL only), got %d", len(results))
	}
	if len(results) > 0 && results[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", results[0].Symbol)
	}
}

func TestScreen_NonOKStatus(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.Screen(ctx, ScreenCriteria{})

	if err == nil {
		t.Error("expected error for non-OK status")
	}
}

func TestScreen_InvalidJSON(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.Screen(ctx, ScreenCriteria{})

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestScreen_ContextCancellation(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	service := NewFMPService("test-api-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.Screen(ctx, ScreenCriteria{})
	if err == nil {
		t.Error("Screen should return error when context is cancelled")
	}
}

func TestScreen_WithPERatioFilter(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/stock-screener" {
			w.Write([]byte(`[
				{
					"symbol": "AAPL",
					"companyName": "Apple Inc.",
					"marketCap": 2500000000000,
					"sector": "Technology",
					"industry": "Consumer Electronics",
					"beta": 1.25,
					"price": 175.50,
					"volume": 50000000,
					"exchangeShortName": "NASDAQ",
					"country": "US",
					"isEtf": false,
					"isActivelyTrading": true
				},
				{
					"symbol": "MSFT",
					"companyName": "Microsoft Corporation",
					"marketCap": 2800000000000,
					"sector": "Technology",
					"industry": "Software",
					"beta": 0.95,
					"price": 380.25,
					"volume": 25000000,
					"exchangeShortName": "NASDAQ",
					"country": "US",
					"isEtf": false,
					"isActivelyTrading": true
				}
			]`))
		} else if r.URL.Path == "/ratios-ttm/AAPL" {
			w.Write([]byte(`[{
				"symbol": "AAPL",
				"peRatioTTM": 12.5,
				"priceToBookRatioTTM": 1.2,
				"dividendYieldPercentageTTM": 0.55,
				"netIncomePerShareTTM": 6.15
			}]`))
		} else if r.URL.Path == "/ratios-ttm/MSFT" {
			w.Write([]byte(`[{
				"symbol": "MSFT",
				"peRatioTTM": 35.0,
				"priceToBookRatioTTM": 12.5,
				"dividendYieldPercentageTTM": 0.8,
				"netIncomePerShareTTM": 10.5
			}]`))
		}
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		PERatioMax: 15.0, // Should filter out MSFT (P/E = 35)
	}

	results, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL (P/E=12.5), got %s", results[0].Symbol)
	}
}

func TestGetCompanyProfile_WithMockServer(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profile/AAPL" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"symbol": "AAPL",
				"companyName": "Apple Inc.",
				"price": 175.50,
				"beta": 1.25,
				"volAvg": 50000000,
				"mktCap": 2500000000000,
				"lastDiv": 0.24,
				"range": "140.82-199.62",
				"changes": 2.35,
				"exchangeShortName": "NASDAQ",
				"industry": "Consumer Electronics",
				"website": "https://www.apple.com",
				"description": "Apple Inc. designs, manufactures, and markets smartphones...",
				"ceo": "Tim Cook",
				"sector": "Technology",
				"country": "US",
				"dcf": 150.25,
				"ipoDate": "1980-12-12",
				"isActivelyTrading": true
			}
		]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	profile, err := service.GetCompanyProfile(ctx, "AAPL")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Symbol != "AAPL" {
		t.Errorf("unexpected symbol: %s", profile.Symbol)
	}
	if profile.CompanyName != "Apple Inc." {
		t.Errorf("unexpected company name: %s", profile.CompanyName)
	}
	if profile.CEO != "Tim Cook" {
		t.Errorf("unexpected CEO: %s", profile.CEO)
	}
	if profile.Sector != "Technology" {
		t.Errorf("unexpected sector: %s", profile.Sector)
	}
	if profile.MarketCap != 2500000000000 {
		t.Errorf("unexpected market cap: %d", profile.MarketCap)
	}
	if profile.Range52Week != "140.82-199.62" {
		t.Errorf("unexpected 52 week range: %s", profile.Range52Week)
	}
	if !profile.IsActivelyTrading {
		t.Error("expected IsActivelyTrading to be true")
	}
}

func TestGetCompanyProfile_NonOKStatus(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetCompanyProfile(ctx, "INVALID")

	if err == nil {
		t.Error("expected error for non-OK status")
	}
}

func TestGetCompanyProfile_EmptyResponse(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetCompanyProfile(ctx, "UNKNOWN")

	if err == nil {
		t.Error("expected error for empty response")
	}
}

func TestGetCompanyProfile_InvalidJSON(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.GetCompanyProfile(ctx, "AAPL")

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetCompanyProfile_ContextCancellation(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	service := NewFMPService("test-api-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := service.GetCompanyProfile(ctx, "AAPL")
	if err == nil {
		t.Error("GetCompanyProfile should return error when context is cancelled")
	}
}

func TestFMPServiceInterface_Implementation(t *testing.T) {
	// Verify FMPService implements FMPServiceInterface
	var _ FMPServiceInterface = (*FMPService)(nil)
}

func TestScreen_WithMarketCapRange(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("marketCapMoreThan") != "1000000000" {
			t.Errorf("marketCapMoreThan = %s, want 1000000000", query.Get("marketCapMoreThan"))
		}
		if query.Get("marketCapLowerThan") != "10000000000" {
			t.Errorf("marketCapLowerThan = %s, want 10000000000", query.Get("marketCapLowerThan"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		MarketCapMin: 1_000_000_000,
		MarketCapMax: 10_000_000_000,
	}

	_, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScreen_WithLimit(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("limit") != "50" {
			t.Errorf("limit = %s, want 50", query.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		Limit: 50,
	}

	_, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRatios_EmptyResponse(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.getRatios(ctx, "UNKNOWN")

	if err == nil {
		t.Error("expected error for empty ratios response")
	}
}

func TestGetRatios_NonOKStatus(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	_, err := service.getRatios(ctx, "AAPL")

	if err == nil {
		t.Error("expected error for non-OK status")
	}
}

func TestScreen_WithDividendYieldFilter(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/stock-screener" {
			w.Write([]byte(`[
				{
					"symbol": "JNJ",
					"companyName": "Johnson & Johnson",
					"marketCap": 400000000000,
					"sector": "Healthcare",
					"industry": "Pharmaceuticals",
					"beta": 0.65,
					"price": 155.00,
					"volume": 10000000,
					"exchangeShortName": "NYSE",
					"country": "US",
					"isEtf": false,
					"isActivelyTrading": true
				},
				{
					"symbol": "NVDA",
					"companyName": "NVIDIA Corporation",
					"marketCap": 1200000000000,
					"sector": "Technology",
					"industry": "Semiconductors",
					"beta": 1.75,
					"price": 480.00,
					"volume": 50000000,
					"exchangeShortName": "NASDAQ",
					"country": "US",
					"isEtf": false,
					"isActivelyTrading": true
				}
			]`))
		} else if r.URL.Path == "/ratios-ttm/JNJ" {
			w.Write([]byte(`[{
				"symbol": "JNJ",
				"peRatioTTM": 15.0,
				"priceToBookRatioTTM": 5.5,
				"dividendYieldPercentageTTM": 3.0,
				"netIncomePerShareTTM": 10.0
			}]`))
		} else if r.URL.Path == "/ratios-ttm/NVDA" {
			w.Write([]byte(`[{
				"symbol": "NVDA",
				"peRatioTTM": 65.0,
				"priceToBookRatioTTM": 25.0,
				"dividendYieldPercentageTTM": 0.04,
				"netIncomePerShareTTM": 7.5
			}]`))
		}
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		DividendYieldMin: 2.0, // Should filter out NVDA
	}

	results, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Symbol != "JNJ" {
		t.Errorf("expected JNJ (dividend=3.0), got %s", results[0].Symbol)
	}
}

func TestScreen_WithEPSFilter(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.URL.Path == "/stock-screener" {
			w.Write([]byte(`[
				{
					"symbol": "PROFIT",
					"companyName": "Profitable Co",
					"marketCap": 100000000000,
					"sector": "Technology",
					"industry": "Software",
					"beta": 1.0,
					"price": 100.00,
					"volume": 1000000,
					"exchangeShortName": "NASDAQ",
					"country": "US",
					"isEtf": false,
					"isActivelyTrading": true
				},
				{
					"symbol": "LOSS",
					"companyName": "Money Loser Inc",
					"marketCap": 50000000000,
					"sector": "Technology",
					"industry": "Software",
					"beta": 2.0,
					"price": 50.00,
					"volume": 500000,
					"exchangeShortName": "NASDAQ",
					"country": "US",
					"isEtf": false,
					"isActivelyTrading": true
				}
			]`))
		} else if r.URL.Path == "/ratios-ttm/PROFIT" {
			w.Write([]byte(`[{
				"symbol": "PROFIT",
				"peRatioTTM": 20.0,
				"priceToBookRatioTTM": 5.0,
				"dividendYieldPercentageTTM": 1.0,
				"netIncomePerShareTTM": 5.0
			}]`))
		} else if r.URL.Path == "/ratios-ttm/LOSS" {
			w.Write([]byte(`[{
				"symbol": "LOSS",
				"peRatioTTM": -10.0,
				"priceToBookRatioTTM": 3.0,
				"dividendYieldPercentageTTM": 0,
				"netIncomePerShareTTM": -5.0
			}]`))
		}
	}))
	defer server.Close()

	service := NewFMPService("test-key")
	service.baseURL = server.URL

	ctx := context.Background()
	criteria := ScreenCriteria{
		EPSMin: 0.01, // Positive earnings only (must be > 0)
	}

	results, err := service.Screen(ctx, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].Symbol != "PROFIT" {
		t.Errorf("expected PROFIT, got %s", results[0].Symbol)
	}
}
