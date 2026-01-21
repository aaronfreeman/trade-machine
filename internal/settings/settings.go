package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// ServiceName represents a configurable service
type ServiceName string

const (
	ServiceOpenAI       ServiceName = "openai"
	ServiceAlpaca       ServiceName = "alpaca"
	ServiceAlphaVantage ServiceName = "alpha_vantage"
	ServiceNewsAPI      ServiceName = "newsapi"
	ServiceFMP          ServiceName = "fmp"
)

// APIKeyConfig represents configuration for a single API key
type APIKeyConfig struct {
	ServiceName ServiceName `json:"service_name"`
	APIKey      string      `json:"api_key,omitempty"`
	APISecret   string      `json:"api_secret,omitempty"` // For services like Alpaca that need both
	BaseURL     string      `json:"base_url,omitempty"`   // Optional base URL override
	Region      string      `json:"region,omitempty"`     // For AWS services
	ModelID     string      `json:"model_id,omitempty"`   // For AI services
}

// Settings holds all user-configurable settings
type Settings struct {
	APIKeys map[ServiceName]*APIKeyConfig `json:"api_keys"`
}

// MaskedAPIKeyConfig represents an API key config with masked secrets
type MaskedAPIKeyConfig struct {
	ServiceName  ServiceName `json:"service_name"`
	APIKey       string      `json:"api_key,omitempty"`
	APISecret    string      `json:"api_secret,omitempty"`
	BaseURL      string      `json:"base_url,omitempty"`
	Region       string      `json:"region,omitempty"`
	ModelID      string      `json:"model_id,omitempty"`
	IsConfigured bool        `json:"is_configured"`
}

// RepositoryInterface defines the database operations needed by Store
type RepositoryInterface interface {
	GetAPIKey(ctx context.Context, serviceName string) (*APIKeyModel, error)
	GetAllAPIKeys(ctx context.Context) ([]APIKeyModel, error)
	UpsertAPIKey(ctx context.Context, apiKey *APIKeyModel) error
	DeleteAPIKey(ctx context.Context, serviceName string) error
}

// APIKeyModel represents the database model for API keys
type APIKeyModel struct {
	ID                 uuid.UUID
	ServiceName        string
	APIKeyEncrypted    []byte
	APISecretEncrypted []byte
	BaseURL            string
	Region             string
	ModelID            string
}

// Store manages persistent storage of settings
type Store struct {
	mu         sync.RWMutex
	filePath   string
	settings   *Settings
	crypto     *Crypto
	passphrase string
	repo       RepositoryInterface
	ctx        context.Context
}

// NewStore creates a new settings store
// Repository is required for database storage
func NewStore(dataDir string, passphrase string, repo RepositoryInterface) (*Store, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required for settings storage")
	}

	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dataDir = filepath.Join(homeDir, ".trade-machine")
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create settings directory: %w", err)
	}

	crypto, err := NewCrypto(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crypto: %w", err)
	}

	store := &Store{
		filePath:   filepath.Join(dataDir, "settings.enc"),
		crypto:     crypto,
		passphrase: passphrase,
		settings:   newDefaultSettings(),
		repo:       repo,
		ctx:        context.Background(),
	}

	// Try to load existing settings from database
	if err := store.loadFromDB(); err != nil {
		fmt.Printf("info: no settings found in database, checking file: %v\n", err)
		// Try to migrate from file (one-time migration)
		if err := store.migrateFromFile(); err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("warning: failed to migrate settings from file: %v\n", err)
		}
	}

	return store, nil
}

// newDefaultSettings creates empty default settings
func newDefaultSettings() *Settings {
	return &Settings{
		APIKeys: make(map[ServiceName]*APIKeyConfig),
	}
}

// load reads settings from encrypted file
func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	decrypted, err := s.crypto.Decrypt(data)
	if err != nil {
		return fmt.Errorf("failed to decrypt settings: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(decrypted, &settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	s.settings = &settings
	return nil
}

// Save persists settings to database
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveToDB()
}

// saveToFile persists settings to encrypted file (legacy)
func (s *Store) saveToFile() error {
	data, err := json.Marshal(s.settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	encrypted, err := s.crypto.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt settings: %w", err)
	}

	if err := os.WriteFile(s.filePath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// saveToDB persists settings to database
func (s *Store) saveToDB() error {
	// This is called with lock already held
	for serviceName, config := range s.settings.APIKeys {
		apiKeyEncrypted, err := s.crypto.Encrypt([]byte(config.APIKey))
		if err != nil {
			return fmt.Errorf("failed to encrypt api key for %s: %w", serviceName, err)
		}

		var apiSecretEncrypted []byte
		if config.APISecret != "" {
			apiSecretEncrypted, err = s.crypto.Encrypt([]byte(config.APISecret))
			if err != nil {
				return fmt.Errorf("failed to encrypt api secret for %s: %w", serviceName, err)
			}
		}

		dbModel := &APIKeyModel{
			ServiceName:        string(serviceName),
			APIKeyEncrypted:    apiKeyEncrypted,
			APISecretEncrypted: apiSecretEncrypted,
			BaseURL:            config.BaseURL,
			Region:             config.Region,
			ModelID:            config.ModelID,
		}

		if err := s.repo.UpsertAPIKey(s.ctx, dbModel); err != nil {
			return fmt.Errorf("failed to save api key for %s: %w", serviceName, err)
		}
	}

	return nil
}

// loadFromDB loads settings from database
func (s *Store) loadFromDB() error {
	apiKeys, err := s.repo.GetAllAPIKeys(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to load api keys from database: %w", err)
	}

	s.settings.APIKeys = make(map[ServiceName]*APIKeyConfig)

	for _, dbModel := range apiKeys {
		config := &APIKeyConfig{
			ServiceName: ServiceName(dbModel.ServiceName),
			BaseURL:     dbModel.BaseURL,
			Region:      dbModel.Region,
			ModelID:     dbModel.ModelID,
		}

		// Decrypt API key
		if len(dbModel.APIKeyEncrypted) > 0 {
			decrypted, err := s.crypto.Decrypt(dbModel.APIKeyEncrypted)
			if err != nil {
				fmt.Printf("warning: failed to decrypt api key for %s: %v\n", dbModel.ServiceName, err)
				continue
			}
			config.APIKey = string(decrypted)
		}

		// Decrypt API secret
		if len(dbModel.APISecretEncrypted) > 0 {
			decrypted, err := s.crypto.Decrypt(dbModel.APISecretEncrypted)
			if err != nil {
				fmt.Printf("warning: failed to decrypt api secret for %s: %v\n", dbModel.ServiceName, err)
				continue
			}
			config.APISecret = string(decrypted)
		}

		s.settings.APIKeys[ServiceName(dbModel.ServiceName)] = config
	}

	return nil
}

// migrateFromFile migrates settings from file to database
func (s *Store) migrateFromFile() error {
	// Load from file
	if err := s.load(); err != nil {
		return err
	}

	// If we have settings and a repo, migrate to database
	if len(s.settings.APIKeys) > 0 && s.repo != nil {
		fmt.Printf("migrating %d API keys from file to database\n", len(s.settings.APIKeys))
		if err := s.saveToDB(); err != nil {
			return fmt.Errorf("failed to migrate to database: %w", err)
		}
		fmt.Println("migration complete")
	}

	return nil
}

// GetAPIKey returns the API key config for a service (unmasked)
func (s *Store) GetAPIKey(service ServiceName) *APIKeyConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if config, ok := s.settings.APIKeys[service]; ok {
		// Return a copy to prevent external modification
		configCopy := *config
		return &configCopy
	}
	return nil
}

// SetAPIKey stores an API key configuration
func (s *Store) SetAPIKey(config *APIKeyConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}
	if config.ServiceName == "" {
		return errors.New("service name is required")
	}

	s.mu.Lock()
	s.settings.APIKeys[config.ServiceName] = config
	s.mu.Unlock()

	return s.Save()
}

// DeleteAPIKey removes an API key configuration
func (s *Store) DeleteAPIKey(service ServiceName) error {
	s.mu.Lock()
	delete(s.settings.APIKeys, service)
	s.mu.Unlock()

	// Delete from database
	if err := s.repo.DeleteAPIKey(s.ctx, string(service)); err != nil {
		return fmt.Errorf("failed to delete from database: %w", err)
	}
	return nil
}

// GetMaskedSettings returns all settings with API keys masked
func (s *Store) GetMaskedSettings() map[ServiceName]*MaskedAPIKeyConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[ServiceName]*MaskedAPIKeyConfig)

	// Include all known services
	for _, service := range []ServiceName{ServiceOpenAI, ServiceAlpaca, ServiceAlphaVantage, ServiceNewsAPI, ServiceFMP} {
		masked := &MaskedAPIKeyConfig{
			ServiceName:  service,
			IsConfigured: false,
		}

		if config, ok := s.settings.APIKeys[service]; ok {
			masked.APIKey = maskString(config.APIKey)
			masked.APISecret = maskString(config.APISecret)
			masked.BaseURL = config.BaseURL
			masked.Region = config.Region
			masked.ModelID = config.ModelID
			masked.IsConfigured = config.APIKey != "" || config.APISecret != ""
		}

		result[service] = masked
	}

	return result
}

// IsConfigured checks if a service has API keys configured
func (s *Store) IsConfigured(service ServiceName) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, ok := s.settings.APIKeys[service]
	if !ok {
		return false
	}

	return config.APIKey != ""
}

// maskString masks a string showing only last 4 characters
func maskString(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}

// GetAllAPIKeys returns all API keys (unmasked) - use with caution
func (s *Store) GetAllAPIKeys() map[ServiceName]*APIKeyConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[ServiceName]*APIKeyConfig)
	for k, v := range s.settings.APIKeys {
		configCopy := *v
		result[k] = &configCopy
	}
	return result
}

// ResetAll removes all API keys (for testing purposes)
func (s *Store) ResetAll() error {
	s.mu.Lock()
	s.settings.APIKeys = make(map[ServiceName]*APIKeyConfig)
	s.mu.Unlock()

	return s.Save()
}

// ServiceDisplayName returns a human-readable name for a service
func ServiceDisplayName(service ServiceName) string {
	switch service {
	case ServiceOpenAI:
		return "OpenAI"
	case ServiceAlpaca:
		return "Alpaca Markets"
	case ServiceAlphaVantage:
		return "Alpha Vantage"
	case ServiceNewsAPI:
		return "NewsAPI"
	case ServiceFMP:
		return "Financial Modeling Prep"
	default:
		return string(service)
	}
}

// ServiceDescription returns a description for a service
func ServiceDescription(service ServiceName) string {
	switch service {
	case ServiceOpenAI:
		return "AI model for stock analysis and recommendations"
	case ServiceAlpaca:
		return "Market data and paper/live trading"
	case ServiceAlphaVantage:
		return "Fundamental company data and financials"
	case ServiceNewsAPI:
		return "News articles for sentiment analysis"
	case ServiceFMP:
		return "Stock screening and additional fundamentals"
	default:
		return ""
	}
}
