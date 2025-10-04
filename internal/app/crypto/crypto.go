package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// Crypto provides AES-GCM encryption/decryption functionality
type Crypto struct {
	key []byte
}

// NewCrypto creates a new Crypto instance with the master key from environment
func NewCrypto() (*Crypto, error) {
	masterKey := os.Getenv("ONECLICK_MASTER_KEY")
	if masterKey == "" {
		return nil, fmt.Errorf("ONECLICK_MASTER_KEY environment variable is required")
	}

	// Decode base64 key if provided, otherwise use raw key
	key, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		// If base64 decode fails, use the raw string
		key = []byte(masterKey)
	}

	// Ensure key is exactly 32 bytes for AES-256
	if len(key) != 32 {
		return nil, fmt.Errorf("master key must be exactly 32 bytes, got %d bytes", len(key))
	}

	return &Crypto{key: key}, nil
}

// Encrypt encrypts the given data using AES-GCM
func (c *Crypto) Encrypt(data []byte) ([]byte, error) {
	// Create a new AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts the given encrypted data using AES-GCM
func (c *Crypto) Decrypt(encryptedData []byte) ([]byte, error) {
	// Create a new AES cipher
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check if we have enough data for nonce + ciphertext
	if len(encryptedData) < gcm.NonceSize() {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract nonce and ciphertext
	nonce := encryptedData[:gcm.NonceSize()]
	ciphertext := encryptedData[gcm.NonceSize():]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (c *Crypto) EncryptString(data string) (string, error) {
	encrypted, err := c.Encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64 encoded encrypted string
func (c *Crypto) DecryptString(encryptedData string) (string, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decrypt
	decrypted, err := c.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
