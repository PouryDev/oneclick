package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {
	// Set a test master key (exactly 32 bytes, not valid base64)
	os.Setenv("ONECLICK_MASTER_KEY", "!@#$%^&*()_+-=[]{}|;':\",./<>?123") // 32 bytes, not base64
	defer os.Unsetenv("ONECLICK_MASTER_KEY")

	crypto, err := NewCrypto()
	assert.NoError(t, err)
	assert.NotNil(t, crypto)

	// Test string encryption/decryption
	originalText := "This is a test kubeconfig content"
	encrypted, err := crypto.EncryptString(originalText)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := crypto.DecryptString(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, originalText, decrypted)

	// Test byte encryption/decryption
	originalBytes := []byte("This is test data")
	encryptedBytes, err := crypto.Encrypt(originalBytes)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedBytes)

	decryptedBytes, err := crypto.Decrypt(encryptedBytes)
	assert.NoError(t, err)
	assert.Equal(t, originalBytes, decryptedBytes)
}

func TestCryptoInvalidKey(t *testing.T) {
	// Test with invalid key length
	os.Setenv("ONECLICK_MASTER_KEY", "short")
	defer os.Unsetenv("ONECLICK_MASTER_KEY")

	_, err := NewCrypto()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be exactly 32 bytes")
}

func TestCryptoMissingKey(t *testing.T) {
	// Test with missing key
	os.Unsetenv("ONECLICK_MASTER_KEY")

	_, err := NewCrypto()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ONECLICK_MASTER_KEY environment variable is required")
}
