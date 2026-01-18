package models

import "github.com/shopspring/decimal"

// Account represents trading account information
type Account struct {
	ID                string          `json:"id"`
	Currency          string          `json:"currency"`
	BuyingPower       decimal.Decimal `json:"buying_power"`
	Cash              decimal.Decimal `json:"cash"`
	PortfolioValue    decimal.Decimal `json:"portfolio_value"`
	Equity            decimal.Decimal `json:"equity"`
	LongMarketValue   decimal.Decimal `json:"long_market_value"`
	ShortMarketValue  decimal.Decimal `json:"short_market_value"`
	InitialMargin     decimal.Decimal `json:"initial_margin"`
	MaintenanceMargin decimal.Decimal `json:"maintenance_margin"`
	TradingBlocked    bool            `json:"trading_blocked"`
	AccountBlocked    bool            `json:"account_blocked"`
	ShortingEnabled   bool            `json:"shorting_enabled"`
}

// AvailableForTrading returns the amount available for new positions
// This is a conservative estimate using buying power
func (a *Account) AvailableForTrading() decimal.Decimal {
	return a.BuyingPower
}
