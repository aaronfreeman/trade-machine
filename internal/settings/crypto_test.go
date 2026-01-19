package settings

import (
	"bytes"
	"testing"
)

func TestCryptoEncryptDecrypt(t *testing.T) {
	crypto, err := NewCrypto("test-passphrase")
	if err != nil {
		t.Fatalf("NewCrypto() error = %v", err)
	}

	plaintext := []byte("Hello, World! This is a test message for encryption.")

	// Encrypt
	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Ciphertext should be different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Encrypt() ciphertext equals plaintext")
	}

	// Ciphertext should be longer than plaintext (due to salt, nonce, and auth tag)
	if len(ciphertext) <= len(plaintext) {
		t.Error("Encrypt() ciphertext not longer than plaintext")
	}

	// Decrypt
	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Decrypted should match original
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestCryptoDefaultPassphrase(t *testing.T) {
	crypto, err := NewCrypto("")
	if err != nil {
		t.Fatalf("NewCrypto() with empty passphrase error = %v", err)
	}

	plaintext := []byte("test data")

	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestCryptoWrongPassphrase(t *testing.T) {
	crypto1, _ := NewCrypto("passphrase1")
	crypto2, _ := NewCrypto("passphrase2")

	plaintext := []byte("secret data")

	// Encrypt with first passphrase
	ciphertext, err := crypto1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with second passphrase - should fail
	_, err = crypto2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong passphrase should return error")
	}
}

func TestCryptoEmptyData(t *testing.T) {
	crypto, _ := NewCrypto("test")

	// Empty plaintext
	ciphertext, err := crypto.Encrypt([]byte{})
	if err != nil {
		t.Fatalf("Encrypt() empty data error = %v", err)
	}

	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() empty data error = %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("Decrypt() = %q, want empty", decrypted)
	}
}

func TestCryptoLargeData(t *testing.T) {
	crypto, _ := NewCrypto("test")

	// 1MB of data
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() large data error = %v", err)
	}

	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() large data error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypt() large data mismatch")
	}
}

func TestCryptoDecryptInvalidData(t *testing.T) {
	crypto, _ := NewCrypto("test")

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"too short", []byte{1, 2, 3, 4, 5}},
		{"salt only", make([]byte, saltSize)},
		{"random garbage", []byte("this is not encrypted data at all")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := crypto.Decrypt(tt.data)
			if err == nil {
				t.Errorf("Decrypt(%s) should return error", tt.name)
			}
		})
	}
}

func TestCryptoMultipleEncryptions(t *testing.T) {
	crypto, _ := NewCrypto("test")

	plaintext := []byte("test data")

	// Encrypt the same data multiple times - should get different ciphertexts
	// (due to random salt and nonce)
	ciphertext1, _ := crypto.Encrypt(plaintext)
	ciphertext2, _ := crypto.Encrypt(plaintext)

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Multiple encryptions of same data produced identical ciphertext")
	}

	// Both should decrypt to the same plaintext
	decrypted1, _ := crypto.Decrypt(ciphertext1)
	decrypted2, _ := crypto.Decrypt(ciphertext2)

	if !bytes.Equal(decrypted1, plaintext) {
		t.Error("First decryption failed")
	}
	if !bytes.Equal(decrypted2, plaintext) {
		t.Error("Second decryption failed")
	}
}

func TestCryptoTampering(t *testing.T) {
	crypto, _ := NewCrypto("test")

	plaintext := []byte("test data")
	ciphertext, _ := crypto.Encrypt(plaintext)

	// Tamper with the ciphertext
	tampered := make([]byte, len(ciphertext))
	copy(tampered, ciphertext)
	tampered[len(tampered)-1] ^= 0xFF // Flip bits in last byte

	_, err := crypto.Decrypt(tampered)
	if err == nil {
		t.Error("Decrypt() of tampered data should return error")
	}
}
