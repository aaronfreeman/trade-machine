package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Quote represents real-time quote data for a stock
type Quote struct {
	Symbol    string          `json:"symbol"`
	Bid       decimal.Decimal `json:"bid"`
	Ask       decimal.Decimal `json:"ask"`
	BidSize   int64           `json:"bid_size"`
	AskSize   int64           `json:"ask_size"`
	Last      decimal.Decimal `json:"last"`
	Volume    int64           `json:"volume"`
	Timestamp time.Time       `json:"timestamp"`
}

// Bar represents OHLCV price data for a time period
type Bar struct {
	Symbol    string          `json:"symbol"`
	Timestamp time.Time       `json:"timestamp"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    int64           `json:"volume"`
	VWAP      decimal.Decimal `json:"vwap"`
}

// Fundamentals represents key fundamental data for a stock
type Fundamentals struct {
	Symbol        string          `json:"symbol"`
	MarketCap     decimal.Decimal `json:"market_cap"`
	PERatio       float64         `json:"pe_ratio"`
	EPS           decimal.Decimal `json:"eps"`
	DividendYield float64         `json:"dividend_yield"`
	Week52High    decimal.Decimal `json:"week52_high"`
	Week52Low     decimal.Decimal `json:"week52_low"`
	Beta          float64         `json:"beta"`
	Revenue       decimal.Decimal `json:"revenue"`
	GrossProfit   decimal.Decimal `json:"gross_profit"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// NewsArticle represents a news article about a stock
type NewsArticle struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Author      string    `json:"author,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}

// TechnicalIndicators holds computed technical analysis indicators
type TechnicalIndicators struct {
	Symbol         string          `json:"symbol"`
	RSI            float64         `json:"rsi"`
	MACD           float64         `json:"macd"`
	MACDSignal     float64         `json:"macd_signal"`
	MACDHistogram  float64         `json:"macd_histogram"`
	SMA20          decimal.Decimal `json:"sma_20"`
	SMA50          decimal.Decimal `json:"sma_50"`
	SMA200         decimal.Decimal `json:"sma_200"`
	EMA12          decimal.Decimal `json:"ema_12"`
	EMA26          decimal.Decimal `json:"ema_26"`
	BollingerUpper decimal.Decimal `json:"bollinger_upper"`
	BollingerLower decimal.Decimal `json:"bollinger_lower"`
	ATR            decimal.Decimal `json:"atr"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
