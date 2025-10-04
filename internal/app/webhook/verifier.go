package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// WebhookVerifier handles webhook signature verification for different Git providers
type WebhookVerifier struct{}

// NewWebhookVerifier creates a new webhook verifier
func NewWebhookVerifier() *WebhookVerifier {
	return &WebhookVerifier{}
}

// VerifyGitHubSignature verifies GitHub webhook signature using HMAC-SHA256
func (w *WebhookVerifier) VerifyGitHubSignature(payload []byte, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	// GitHub sends signature as "sha256=<hash>"
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature format")
	}

	expectedSignature := strings.TrimPrefix(signature, "sha256=")
	actualSignature := w.computeHMACSHA256(payload, secret)

	if !hmac.Equal([]byte(expectedSignature), []byte(actualSignature)) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// VerifyGitLabSignature verifies GitLab webhook signature using HMAC-SHA256
func (w *WebhookVerifier) VerifyGitLabSignature(payload []byte, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	// GitLab sends signature as "sha256=<hash>"
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature format")
	}

	expectedSignature := strings.TrimPrefix(signature, "sha256=")
	actualSignature := w.computeHMACSHA256(payload, secret)

	if !hmac.Equal([]byte(expectedSignature), []byte(actualSignature)) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// VerifyGiteaSignature verifies Gitea webhook signature using HMAC-SHA256
func (w *WebhookVerifier) VerifyGiteaSignature(payload []byte, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	// Gitea sends signature as "sha256=<hash>"
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature format")
	}

	expectedSignature := strings.TrimPrefix(signature, "sha256=")
	actualSignature := w.computeHMACSHA256(payload, secret)

	if !hmac.Equal([]byte(expectedSignature), []byte(actualSignature)) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// VerifyLegacyGitHubSignature verifies legacy GitHub webhook signature using HMAC-SHA1
func (w *WebhookVerifier) VerifyLegacyGitHubSignature(payload []byte, signature, secret string) error {
	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	// Legacy GitHub sends signature as "sha1=<hash>"
	if !strings.HasPrefix(signature, "sha1=") {
		return fmt.Errorf("invalid signature format")
	}

	expectedSignature := strings.TrimPrefix(signature, "sha1=")
	actualSignature := w.computeHMACSHA1(payload, secret)

	if !hmac.Equal([]byte(expectedSignature), []byte(actualSignature)) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// VerifySignature verifies webhook signature based on provider type
func (w *WebhookVerifier) VerifySignature(provider string, payload []byte, signature, secret string) error {
	switch strings.ToLower(provider) {
	case "github":
		// Try SHA256 first, fallback to SHA1 for legacy
		err := w.VerifyGitHubSignature(payload, signature, secret)
		if err != nil {
			// Try legacy SHA1
			return w.VerifyLegacyGitHubSignature(payload, signature, secret)
		}
		return nil
	case "gitlab":
		return w.VerifyGitLabSignature(payload, signature, secret)
	case "gitea":
		return w.VerifyGiteaSignature(payload, signature, secret)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// computeHMACSHA256 computes HMAC-SHA256 signature
func (w *WebhookVerifier) computeHMACSHA256(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// computeHMACSHA1 computes HMAC-SHA1 signature
func (w *WebhookVerifier) computeHMACSHA1(payload []byte, secret string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// ExtractSignatureFromHeader extracts signature from HTTP header
func (w *WebhookVerifier) ExtractSignatureFromHeader(header string) string {
	// Handle different header formats
	// GitHub: X-Hub-Signature-256 or X-Hub-Signature
	// GitLab: X-Gitlab-Token or X-Gitlab-Signature
	// Gitea: X-Gitea-Signature

	// For now, we'll use the standard X-Hub-Signature-256 format
	// In a real implementation, you'd check different headers based on provider
	return header
}
