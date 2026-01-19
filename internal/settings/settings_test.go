package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()

	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if store == nil {
		t.Fatal("NewStore() returned nil store")
	}

	if store.filePath != filepath.Join(tmpDir, "settings.enc") {
		t.Errorf("NewStore() filePath = %v, want %v", store.filePath, filepath.Join(tmpDir, "settings.enc"))
	}
}

func TestNewStoreWithDefaultDir(t *testing.T) {
	// Test with empty dataDir - should use home directory
	store, err := NewStore("", "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() with empty dir error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".trade-machine")
	expectedPath := filepath.Join(expectedDir, "settings.enc")

	if store.filePath != expectedPath {
		t.Errorf("NewStore() filePath = %v, want %v", store.filePath, expectedPath)
	}

	// Cleanup
	os.RemoveAll(expectedDir)
}

func TestSetAndGetAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set an API key
	config := &APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test123456789",
	}

	err = store.SetAPIKey(config)
	if err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	// Get it back
	retrieved := store.GetAPIKey(ServiceOpenAI)
	if retrieved == nil {
		t.Fatal("GetAPIKey() returned nil")
	}

	if retrieved.APIKey != config.APIKey {
		t.Errorf("GetAPIKey() APIKey = %v, want %v", retrieved.APIKey, config.APIKey)
	}
}

func TestSetAPIKeyWithSecret(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set Alpaca config with both key and secret
	config := &APIKeyConfig{
		ServiceName: ServiceAlpaca,
		APIKey:      "AKTEST123",
		APISecret:   "secret456",
		BaseURL:     "https://paper-api.alpaca.markets",
	}

	err = store.SetAPIKey(config)
	if err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	retrieved := store.GetAPIKey(ServiceAlpaca)
	if retrieved == nil {
		t.Fatal("GetAPIKey() returned nil")
	}

	if retrieved.APIKey != config.APIKey {
		t.Errorf("GetAPIKey() APIKey = %v, want %v", retrieved.APIKey, config.APIKey)
	}
	if retrieved.APISecret != config.APISecret {
		t.Errorf("GetAPIKey() APISecret = %v, want %v", retrieved.APISecret, config.APISecret)
	}
	if retrieved.BaseURL != config.BaseURL {
		t.Errorf("GetAPIKey() BaseURL = %v, want %v", retrieved.BaseURL, config.BaseURL)
	}
}

func TestDeleteAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set an API key
	config := &APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test123456789",
	}
	store.SetAPIKey(config)

	// Verify it exists
	if !store.IsConfigured(ServiceOpenAI) {
		t.Fatal("IsConfigured() = false, want true")
	}

	// Delete it
	err = store.DeleteAPIKey(ServiceOpenAI)
	if err != nil {
		t.Fatalf("DeleteAPIKey() error = %v", err)
	}

	// Verify it's gone
	if store.IsConfigured(ServiceOpenAI) {
		t.Error("IsConfigured() = true after delete, want false")
	}

	retrieved := store.GetAPIKey(ServiceOpenAI)
	if retrieved != nil {
		t.Error("GetAPIKey() returned non-nil after delete")
	}
}

func TestGetMaskedSettings(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set an API key
	config := &APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test123456789",
	}
	store.SetAPIKey(config)

	// Get masked settings
	masked := store.GetMaskedSettings()

	// Should have all services
	if len(masked) != 5 {
		t.Errorf("GetMaskedSettings() returned %d services, want 5", len(masked))
	}

	// OpenAI should be configured and masked
	openAI := masked[ServiceOpenAI]
	if openAI == nil {
		t.Fatal("GetMaskedSettings() missing OpenAI")
	}
	if !openAI.IsConfigured {
		t.Error("GetMaskedSettings() OpenAI.IsConfigured = false, want true")
	}
	if openAI.APIKey != "****6789" {
		t.Errorf("GetMaskedSettings() OpenAI.APIKey = %v, want ****6789", openAI.APIKey)
	}

	// Other services should not be configured
	alpaca := masked[ServiceAlpaca]
	if alpaca.IsConfigured {
		t.Error("GetMaskedSettings() Alpaca.IsConfigured = true, want false")
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "****bcde"},
		{"sk-test123456789", "****6789"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskString(tt.input)
			if result != tt.expected {
				t.Errorf("maskString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Initially not configured
	if store.IsConfigured(ServiceOpenAI) {
		t.Error("IsConfigured() = true for unconfigured service")
	}

	// Set config
	store.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test",
	})

	// Now configured
	if !store.IsConfigured(ServiceOpenAI) {
		t.Error("IsConfigured() = false for configured service")
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create store and save data
	store1, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	store1.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-persistent-test",
	})
	store1.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceAlpaca,
		APIKey:      "AKTEST",
		APISecret:   "secret",
	})

	// Create new store with same path - should load saved data
	store2, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() second load error = %v", err)
	}

	// Verify data persisted
	openAI := store2.GetAPIKey(ServiceOpenAI)
	if openAI == nil || openAI.APIKey != "sk-persistent-test" {
		t.Error("Persisted OpenAI key not loaded correctly")
	}

	alpaca := store2.GetAPIKey(ServiceAlpaca)
	if alpaca == nil || alpaca.APIKey != "AKTEST" || alpaca.APISecret != "secret" {
		t.Error("Persisted Alpaca key not loaded correctly")
	}
}

func TestWrongPassphrase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create store and save data
	store1, err := NewStore(tmpDir, "correct-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	store1.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test",
	})

	// Try to load with wrong passphrase - should fail gracefully
	// and return empty settings
	store2, err := NewStore(tmpDir, "wrong-passphrase")
	if err != nil {
		t.Fatalf("NewStore() with wrong passphrase error = %v", err)
	}

	// Should have empty settings (decryption failed)
	if store2.IsConfigured(ServiceOpenAI) {
		t.Error("IsConfigured() = true with wrong passphrase, expected empty settings")
	}
}

func TestSetAPIKeyValidation(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Test nil config
	err = store.SetAPIKey(nil)
	if err == nil {
		t.Error("SetAPIKey(nil) should return error")
	}

	// Test empty service name
	err = store.SetAPIKey(&APIKeyConfig{
		ServiceName: "",
		APIKey:      "test",
	})
	if err == nil {
		t.Error("SetAPIKey() with empty ServiceName should return error")
	}
}

func TestServiceDisplayName(t *testing.T) {
	tests := []struct {
		service  ServiceName
		expected string
	}{
		{ServiceOpenAI, "OpenAI"},
		{ServiceAlpaca, "Alpaca Markets"},
		{ServiceAlphaVantage, "Alpha Vantage"},
		{ServiceNewsAPI, "NewsAPI"},
		{ServiceFMP, "Financial Modeling Prep"},
		{ServiceName("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.service), func(t *testing.T) {
			result := ServiceDisplayName(tt.service)
			if result != tt.expected {
				t.Errorf("ServiceDisplayName(%v) = %v, want %v", tt.service, result, tt.expected)
			}
		})
	}
}

func TestServiceDescription(t *testing.T) {
	tests := []struct {
		service   ServiceName
		hasResult bool
	}{
		{ServiceOpenAI, true},
		{ServiceAlpaca, true},
		{ServiceAlphaVantage, true},
		{ServiceNewsAPI, true},
		{ServiceFMP, true},
		{ServiceName("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.service), func(t *testing.T) {
			result := ServiceDescription(tt.service)
			if tt.hasResult && result == "" {
				t.Errorf("ServiceDescription(%v) = empty, want non-empty", tt.service)
			}
			if !tt.hasResult && result != "" {
				t.Errorf("ServiceDescription(%v) = %v, want empty", tt.service, result)
			}
		})
	}
}

func TestGetAllAPIKeys(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir, "test-passphrase")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set multiple keys
	store.SetAPIKey(&APIKeyConfig{ServiceName: ServiceOpenAI, APIKey: "openai-key"})
	store.SetAPIKey(&APIKeyConfig{ServiceName: ServiceAlpaca, APIKey: "alpaca-key", APISecret: "alpaca-secret"})

	// Get all
	all := store.GetAllAPIKeys()

	if len(all) != 2 {
		t.Errorf("GetAllAPIKeys() returned %d keys, want 2", len(all))
	}

	if all[ServiceOpenAI] == nil || all[ServiceOpenAI].APIKey != "openai-key" {
		t.Error("GetAllAPIKeys() missing or incorrect OpenAI key")
	}
	if all[ServiceAlpaca] == nil || all[ServiceAlpaca].APIKey != "alpaca-key" {
		t.Error("GetAllAPIKeys() missing or incorrect Alpaca key")
	}
}
