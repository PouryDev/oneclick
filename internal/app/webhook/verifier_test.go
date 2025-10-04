package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookVerifier_VerifyGitHubSignature(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Test valid signature
	err := verifier.VerifyGitHubSignature(payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test invalid signature
	err = verifier.VerifyGitHubSignature(payload, "sha256=invalid", secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")

	// Test missing signature
	err = verifier.VerifyGitHubSignature(payload, "", secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing signature")

	// Test invalid signature format
	err = verifier.VerifyGitHubSignature(payload, "invalid-format", secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature format")
}

func TestWebhookVerifier_VerifyLegacyGitHubSignature(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute expected signature using SHA1
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha1=" + hex.EncodeToString(mac.Sum(nil))

	// Test valid signature
	err := verifier.VerifyLegacyGitHubSignature(payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test invalid signature
	err = verifier.VerifyLegacyGitHubSignature(payload, "sha1=invalid", secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")
}

func TestWebhookVerifier_VerifyGitLabSignature(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Test valid signature
	err := verifier.VerifyGitLabSignature(payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test invalid signature
	err = verifier.VerifyGitLabSignature(payload, "sha256=invalid", secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")
}

func TestWebhookVerifier_VerifyGiteaSignature(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Test valid signature
	err := verifier.VerifyGiteaSignature(payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test invalid signature
	err = verifier.VerifyGiteaSignature(payload, "sha256=invalid", secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")
}

func TestWebhookVerifier_VerifySignature(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Test GitHub
	err := verifier.VerifySignature("github", payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test GitLab
	err = verifier.VerifySignature("gitlab", payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test Gitea
	err = verifier.VerifySignature("gitea", payload, expectedSignature, secret)
	assert.NoError(t, err)

	// Test unsupported provider
	err = verifier.VerifySignature("bitbucket", payload, expectedSignature, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")

	// Test GitHub with legacy SHA1
	macSha1 := hmac.New(sha1.New, []byte(secret))
	macSha1.Write(payload)
	expectedSignatureSha1 := "sha1=" + hex.EncodeToString(macSha1.Sum(nil))

	err = verifier.VerifySignature("github", payload, expectedSignatureSha1, secret)
	assert.NoError(t, err)
}

func TestWebhookVerifier_ComputeHMACSHA256(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute signature manually
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	// Test the method
	result := verifier.computeHMACSHA256(payload, secret)
	assert.Equal(t, expected, result)
}

func TestWebhookVerifier_ComputeHMACSHA1(t *testing.T) {
	verifier := NewWebhookVerifier()
	payload := []byte("test payload")
	secret := "test-secret"

	// Compute signature manually
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	// Test the method
	result := verifier.computeHMACSHA1(payload, secret)
	assert.Equal(t, expected, result)
}
