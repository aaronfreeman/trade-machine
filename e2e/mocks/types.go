package mocks

// BedrockResponse represents the expected response from Claude analysis.
type BedrockResponse struct {
	Score      float64  `json:"score"`
	Confidence float64  `json:"confidence"`
	Reasoning  string   `json:"reasoning"`
	KeyFactors []string `json:"key_factors,omitempty"`
	Signals    []string `json:"signals,omitempty"`
	KeyThemes  []string `json:"key_themes,omitempty"`
}

// AlpacaBar represents OHLCV bar data from Alpaca.
type AlpacaBar struct {
	Timestamp string  `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    int64   `json:"v"`
}

// AlpacaAccount represents an Alpaca trading account.
type AlpacaAccount struct {
	ID               string `json:"id"`
	AccountNumber    string `json:"account_number"`
	Status           string `json:"status"`
	Currency         string `json:"currency"`
	Cash             string `json:"cash"`
	PortfolioValue   string `json:"portfolio_value"`
	BuyingPower      string `json:"buying_power"`
	DaytradingBuying string `json:"daytrading_buying_power"`
}

// AlpacaPosition represents a position in the Alpaca account.
type AlpacaPosition struct {
	AssetID        string `json:"asset_id"`
	Symbol         string `json:"symbol"`
	Exchange       string `json:"exchange"`
	AssetClass     string `json:"asset_class"`
	Qty            string `json:"qty"`
	AvgEntryPrice  string `json:"avg_entry_price"`
	Side           string `json:"side"`
	MarketValue    string `json:"market_value"`
	CostBasis      string `json:"cost_basis"`
	UnrealizedPL   string `json:"unrealized_pl"`
	UnrealizedPLPC string `json:"unrealized_plpc"`
	CurrentPrice   string `json:"current_price"`
}

// AlpacaOrder represents an order in Alpaca.
type AlpacaOrder struct {
	ID            string `json:"id"`
	ClientOrderID string `json:"client_order_id"`
	Status        string `json:"status"`
	Symbol        string `json:"symbol"`
	Qty           string `json:"qty"`
	FilledQty     string `json:"filled_qty"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	TimeInForce   string `json:"time_in_force"`
}

// AlphaVantageFundamentals represents company overview data from Alpha Vantage.
type AlphaVantageFundamentals struct {
	Symbol        string `json:"Symbol"`
	Name          string `json:"Name"`
	Exchange      string `json:"Exchange"`
	Currency      string `json:"Currency"`
	Country       string `json:"Country"`
	Sector        string `json:"Sector"`
	Industry      string `json:"Industry"`
	MarketCap     string `json:"MarketCapitalization"`
	PERatio       string `json:"PERatio"`
	EPS           string `json:"EPS"`
	Beta          string `json:"Beta"`
	WeekHigh52    string `json:"52WeekHigh"`
	WeekLow52     string `json:"52WeekLow"`
	DividendYield string `json:"DividendYield"`
}

// NewsArticle represents a news article from NewsAPI.
type NewsArticle struct {
	Source      map[string]string `json:"source"`
	Author      string            `json:"author"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	PublishedAt string            `json:"publishedAt"`
}

// FMPScreenerResult represents a stock from FMP screener.
type FMPScreenerResult struct {
	Symbol        string  `json:"symbol"`
	CompanyName   string  `json:"companyName"`
	MarketCap     int64   `json:"marketCap"`
	PERatio       float64 `json:"peRatio"`
	PBRatio       float64 `json:"priceToBookRatio"`
	EPS           float64 `json:"eps"`
	DividendYield float64 `json:"dividendYield"`
	Sector        string  `json:"sector"`
	Industry      string  `json:"industry"`
	Country       string  `json:"country"`
	Exchange      string  `json:"exchange"`
}
