package agents

import (
	"trade-machine/services"
)

// Type aliases for service interfaces - defined in services package
// These aliases allow agents to reference interfaces without importing concrete implementations
type LLMService = services.LLMService
type AlphaVantageServiceInterface = services.AlphaVantageServiceInterface
type NewsAPIServiceInterface = services.NewsAPIServiceInterface
type AlpacaServiceInterface = services.AlpacaServiceInterface
