package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/app/webhook"
)

type WebhookHandler struct {
	repositoryService services.RepositoryService
	verifier          *webhook.WebhookVerifier
	logger            *zap.Logger
}

func NewWebhookHandler(repositoryService services.RepositoryService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		repositoryService: repositoryService,
		verifier:          webhook.NewWebhookVerifier(),
		logger:            logger,
	}
}

// GitWebhook godoc
// @Summary Git webhook endpoint
// @Description Public webhook endpoint for Git providers (GitHub, GitLab, Gitea)
// @Tags webhooks
// @Accept json
// @Produce json
// @Param X-Hub-Signature-256 header string false "GitHub signature"
// @Param X-Hub-Signature header string false "GitHub legacy signature"
// @Param X-Gitlab-Signature header string false "GitLab signature"
// @Param X-Gitea-Signature header string false "Gitea signature"
// @Param provider query string true "Git provider (github, gitlab, gitea)"
// @Param secret query string false "Webhook secret for signature verification"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /hooks/git [post]
func (h *WebhookHandler) GitWebhook(c *gin.Context) {
	// Get provider from query parameter
	provider := c.Query("provider")
	if provider == "" {
		h.logger.Error("Missing provider parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider parameter is required"})
		return
	}

	// Validate provider
	provider = strings.ToLower(provider)
	if provider != "github" && provider != "gitlab" && provider != "gitea" {
		h.logger.Error("Invalid provider", zap.String("provider", provider))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider. Must be github, gitlab, or gitea"})
		return
	}

	// Get webhook secret from query parameter
	secret := c.Query("secret")
	if secret == "" {
		h.logger.Warn("No webhook secret provided", zap.String("provider", provider))
		// Continue without signature verification
	}

	// Read the request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get signature from appropriate header based on provider
	var signature string
	switch provider {
	case "github":
		signature = c.GetHeader("X-Hub-Signature-256")
		if signature == "" {
			signature = c.GetHeader("X-Hub-Signature") // Legacy SHA1
		}
	case "gitlab":
		signature = c.GetHeader("X-Gitlab-Signature")
	case "gitea":
		signature = c.GetHeader("X-Gitea-Signature")
	}

	// Verify signature if secret is provided
	if secret != "" && signature != "" {
		err := h.verifier.VerifySignature(provider, payload, signature, secret)
		if err != nil {
			h.logger.Error("Signature verification failed",
				zap.String("provider", provider),
				zap.String("signature", signature),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}
		h.logger.Info("Signature verified successfully", zap.String("provider", provider))
	} else if secret != "" {
		h.logger.Warn("Secret provided but no signature found", zap.String("provider", provider))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Signature required when secret is provided"})
		return
	}

	// Process the webhook
	err = h.repositoryService.ProcessWebhook(c.Request.Context(), provider, payload, signature)
	if err != nil {
		h.logger.Error("Failed to process webhook",
			zap.String("provider", provider),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	h.logger.Info("Webhook processed successfully", zap.String("provider", provider))
	c.JSON(http.StatusAccepted, gin.H{"message": "Webhook processed successfully"})
}

// TestWebhook godoc
// @Summary Test webhook endpoint
// @Description Test endpoint for webhook functionality
// @Tags webhooks
// @Accept json
// @Produce json
// @Param provider query string true "Git provider (github, gitlab, gitea)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /hooks/test [get]
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider parameter is required"})
		return
	}

	provider = strings.ToLower(provider)
	if provider != "github" && provider != "gitlab" && provider != "gitea" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider. Must be github, gitlab, or gitea"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Webhook endpoint is working",
		"provider": provider,
		"status":   "ready",
	})
}
