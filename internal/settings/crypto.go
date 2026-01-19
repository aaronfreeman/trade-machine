package settings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize   = 16
	keySize    = 32 // AES-256
	iterations = 100000
)

// Crypto handles encryption and decryption of settings
type Crypto struct {
	passphrase string
}

// NewCrypto creates a new Crypto instance
func NewCrypto(passphrase string) (*Crypto, error) {
	if passphrase == "" {
		// Use a default passphrase based on machine-specific data
		// In production, this should be derived from OS keychain or similar
		passphrase = getDefaultPassphrase()
	}
	return &Crypto{passphrase: passphrase}, nil
}

// getDefaultPassphrase generates a default passphrase
// This provides basic protection for local files while allowing app to work without setup
func getDefaultPassphrase() string {
	// Use a fixed default that provides basic obfuscation
	// For stronger security, users should provide their own passphrase
	return "trade-machine-default-key-2024"
}

// deriveKey derives an AES key from passphrase and salt using PBKDF2
func (c *Crypto) deriveKey(salt []byte) []byte {
	return pbkdf2.Key([]byte(c.passphrase), salt, iterations, keySize, sha256.New)
}

// Encrypt encrypts plaintext using AES-256-GCM
func (c *Crypto) Encrypt(plaintext []byte) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Derive key from passphrase
	key := c.deriveKey(salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Prepend salt to ciphertext
	result := make([]byte, saltSize+len(ciphertext))
	copy(result, salt)
	copy(result[saltSize:], ciphertext)

	return result, nil
}

// Decrypt decrypts ciphertext encrypted with Encrypt
func (c *Crypto) Decrypt(data []byte) ([]byte, error) {
	if len(data) < saltSize {
		return nil, errors.New("ciphertext too short")
	}

	// Extract salt
	salt := data[:saltSize]
	ciphertext := data[saltSize:]

	// Derive key from passphrase
	key := c.deriveKey(salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	// Extract nonce and decrypt
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid passphrase or corrupted data")
	}

	return plaintext, nil
}
