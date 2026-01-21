package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestDatabaseStorage tests the database storage path
func TestDatabaseStorage(t *testing.T) {
	tmpDir := t.TempDir()
	repo := newMockRepository()

	store, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set an API key
	config := &APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-database-test",
		BaseURL:     "https://api.openai.com",
	}

	err = store.SetAPIKey(config)
	if err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	// Verify it was saved to the repository
	if len(repo.apiKeys) != 1 {
		t.Errorf("Expected 1 key in repository, got %d", len(repo.apiKeys))
	}

	dbKey := repo.apiKeys[string(ServiceOpenAI)]
	if dbKey == nil {
		t.Fatal("Key not found in repository")
	}

	// Verify the key is encrypted
	if string(dbKey.APIKeyEncrypted) == "sk-database-test" {
		t.Error("API key should be encrypted in database")
	}

	// Verify base URL is stored unencrypted
	if dbKey.BaseURL != "https://api.openai.com" {
		t.Errorf("BaseURL = %v, want https://api.openai.com", dbKey.BaseURL)
	}

	// Get the key back
	retrieved := store.GetAPIKey(ServiceOpenAI)
	if retrieved == nil {
		t.Fatal("GetAPIKey() returned nil")
	}

	// Verify it was decrypted correctly
	if retrieved.APIKey != "sk-database-test" {
		t.Errorf("GetAPIKey() APIKey = %v, want sk-database-test", retrieved.APIKey)
	}
}

// TestDatabasePersistence tests that keys persist across store instances
func TestDatabasePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	repo := newMockRepository()

	// Create first store and save data
	store1, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	store1.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-persistent-db-test",
	})
	store1.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceAlpaca,
		APIKey:      "AKTEST-DB",
		APISecret:   "secret-db",
	})

	// Create second store with same repository - should load from DB
	store2, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() second load error = %v", err)
	}

	// Verify data loaded from database
	openAI := store2.GetAPIKey(ServiceOpenAI)
	if openAI == nil || openAI.APIKey != "sk-persistent-db-test" {
		t.Error("Persisted OpenAI key not loaded from database correctly")
	}

	alpaca := store2.GetAPIKey(ServiceAlpaca)
	if alpaca == nil || alpaca.APIKey != "AKTEST-DB" || alpaca.APISecret != "secret-db" {
		t.Error("Persisted Alpaca key not loaded from database correctly")
	}
}

// TestDatabaseDelete tests deletion from database
func TestDatabaseDelete(t *testing.T) {
	tmpDir := t.TempDir()
	repo := newMockRepository()

	store, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Add a key
	store.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test",
	})

	// Verify it's in the repository
	if len(repo.apiKeys) != 1 {
		t.Errorf("Expected 1 key in repository, got %d", len(repo.apiKeys))
	}

	// Delete it
	err = store.DeleteAPIKey(ServiceOpenAI)
	if err != nil {
		t.Fatalf("DeleteAPIKey() error = %v", err)
	}

	// Verify it's gone from repository
	if len(repo.apiKeys) != 0 {
		t.Errorf("Expected 0 keys in repository after delete, got %d", len(repo.apiKeys))
	}

	// Verify it's gone from store
	if store.IsConfigured(ServiceOpenAI) {
		t.Error("IsConfigured() = true after delete")
	}
}

// TestDatabaseWithSecret tests storing keys with secrets
func TestDatabaseWithSecret(t *testing.T) {
	tmpDir := t.TempDir()
	repo := newMockRepository()

	store, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set Alpaca with both key and secret
	config := &APIKeyConfig{
		ServiceName: ServiceAlpaca,
		APIKey:      "AKTEST123",
		APISecret:   "secret456",
		BaseURL:     "https://paper-api.alpaca.markets",
		Region:      "us-east-1",
	}

	err = store.SetAPIKey(config)
	if err != nil {
		t.Fatalf("SetAPIKey() error = %v", err)
	}

	// Verify both are encrypted in database
	dbKey := repo.apiKeys[string(ServiceAlpaca)]
	if string(dbKey.APIKeyEncrypted) == "AKTEST123" {
		t.Error("API key should be encrypted")
	}
	if string(dbKey.APISecretEncrypted) == "secret456" {
		t.Error("API secret should be encrypted")
	}

	// Verify unencrypted fields
	if dbKey.BaseURL != "https://paper-api.alpaca.markets" {
		t.Errorf("BaseURL = %v, want https://paper-api.alpaca.markets", dbKey.BaseURL)
	}
	if dbKey.Region != "us-east-1" {
		t.Errorf("Region = %v, want us-east-1", dbKey.Region)
	}

	// Retrieve and verify decryption
	retrieved := store.GetAPIKey(ServiceAlpaca)
	if retrieved.APIKey != "AKTEST123" {
		t.Errorf("APIKey = %v, want AKTEST123", retrieved.APIKey)
	}
	if retrieved.APISecret != "secret456" {
		t.Errorf("APISecret = %v, want secret456", retrieved.APISecret)
	}
}

// TestFileMigrationToDatabase tests automatic migration from file to database
func TestFileMigrationToDatabase(t *testing.T) {
	tmpDir := t.TempDir()

	// Manually create an encrypted settings file with some data
	crypto, err := NewCrypto("test-passphrase")
	if err != nil {
		t.Fatalf("Failed to create crypto: %v", err)
	}

	// Create settings data
	settings := &Settings{
		APIKeys: map[ServiceName]*APIKeyConfig{
			ServiceOpenAI: {
				ServiceName: ServiceOpenAI,
				APIKey:      "sk-migrate-test",
			},
			ServiceAlpaca: {
				ServiceName: ServiceAlpaca,
				APIKey:      "AKMIGRATE",
				APISecret:   "migrate-secret",
			},
		},
	}

	// Marshal and encrypt
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	encrypted, err := crypto.Encrypt(data)
	if err != nil {
		t.Fatalf("Failed to encrypt settings: %v", err)
	}

	// Write to file
	filePath := filepath.Join(tmpDir, "settings.enc")
	if err := os.WriteFile(filePath, encrypted, 0600); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}

	// Now simulate an empty database by creating a custom mock that returns
	// an error on first GetAllAPIKeys (simulating empty DB), then works normally
	repo := &mockRepositoryWithOnce{
		mockRepository: &mockRepository{
			apiKeys: make(map[string]*APIKeyModel),
		},
		firstGetAllKeysError: errors.New("no keys found"),
	}
	
	// Create database-backed store - should trigger migration
	dbStore, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() database-backed error = %v", err)
	}

	// Verify data was migrated to database
	if len(repo.apiKeys) != 2 {
		t.Errorf("Expected 2 keys migrated to database, got %d", len(repo.apiKeys))
	}

	// Verify we can retrieve the migrated keys
	openAI := dbStore.GetAPIKey(ServiceOpenAI)
	if openAI == nil || openAI.APIKey != "sk-migrate-test" {
		t.Error("Migrated OpenAI key not accessible")
	}

	alpaca := dbStore.GetAPIKey(ServiceAlpaca)
	if alpaca == nil || alpaca.APIKey != "AKMIGRATE" {
		t.Error("Migrated Alpaca key not accessible")
	}
}

// TestDatabaseError tests error handling
func TestDatabaseError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := newMockRepository()

	store, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set error on repository
	repo.err = errors.New("database connection lost")

	// Try to save - should get error
	err = store.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test",
	})
	if err == nil {
		t.Error("SetAPIKey() should return error when database fails")
	}

	// Reset error
	repo.err = nil

	// Should be able to save now
	err = store.SetAPIKey(&APIKeyConfig{
		ServiceName: ServiceOpenAI,
		APIKey:      "sk-test-2",
	})
	if err != nil {
		t.Errorf("SetAPIKey() after reset error = %v", err)
	}
}

// TestNoFileMigrationWhenDBHasData tests that we don't migrate when DB already has data
func TestNoFileMigrationWhenDBHasData(t *testing.T) {
	tmpDir := t.TempDir()

	// Manually create an encrypted settings file with some data
	crypto, err := NewCrypto("test-passphrase")
	if err != nil {
		t.Fatalf("Failed to create crypto: %v", err)
	}

	// Create settings data
	settings := &Settings{
		APIKeys: map[ServiceName]*APIKeyConfig{
			ServiceOpenAI: {
				ServiceName: ServiceOpenAI,
				APIKey:      "sk-old-file-key",
			},
		},
	}

	// Marshal and encrypt
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	encrypted, err := crypto.Encrypt(data)
	if err != nil {
		t.Fatalf("Failed to encrypt settings: %v", err)
	}

	// Write to file
	filePath := filepath.Join(tmpDir, "settings.enc")
	if err := os.WriteFile(filePath, encrypted, 0600); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}

	// Pre-populate database with different data
	repo := newMockRepository()
	repo.apiKeys[string(ServiceAlpaca)] = &APIKeyModel{
		ServiceName: string(ServiceAlpaca),
		APIKeyEncrypted: func() []byte {
			crypto, _ := NewCrypto("test-passphrase")
			encrypted, _ := crypto.Encrypt([]byte("sk-existing-db-key"))
			return encrypted
		}(),
	}

	// Create database-backed store - should load from DB, not migrate from file
	dbStore, err := NewStore(tmpDir, "test-passphrase", repo)
	if err != nil {
		t.Fatalf("NewStore() database-backed error = %v", err)
	}

	// Should have the DB key, not the file key
	if dbStore.IsConfigured(ServiceOpenAI) {
		t.Error("Should not have migrated file key when DB already has data")
	}
	if !dbStore.IsConfigured(ServiceAlpaca) {
		t.Error("Should have loaded existing DB key")
	}
}
