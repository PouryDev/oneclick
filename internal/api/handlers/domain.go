package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/domain"
)

type DomainHandler struct {
	domainService services.DomainService
	logger        *zap.Logger
	validator     *validator.Validate
}

func NewDomainHandler(domainService services.DomainService, logger *zap.Logger) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
		logger:        logger,
		validator:     validator.New(),
	}
}

// CreateDomain godoc
// @Summary Create a domain
// @Description Create a new domain for an application with DNS provider configuration
// @Tags domains
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appId path string true "Application ID"
// @Param request body domain.CreateDomainRequest true "Domain creation request"
// @Success 201 {object} domain.DomainResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{appId}/domains [post]
func (h *DomainHandler) CreateDomain(c *gin.Context) {
	appIDStr := c.Param("appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		h.logger.Warn("Invalid application ID format", zap.String("appID", appIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for CreateDomain")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var req domain.CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for CreateDomain", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for CreateDomain", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.domainService.CreateDomain(c.Request.Context(), userIDUUID, appID, req)
	if err != nil {
		h.logger.Error("Failed to create domain", zap.Error(err), zap.String("appID", appIDStr))
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create domain"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetDomainsByApp godoc
// @Summary Get domains for an application
// @Description Get a list of domains for a specific application
// @Tags domains
// @Produce json
// @Security BearerAuth
// @Param appId path string true "Application ID"
// @Success 200 {array} domain.DomainResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{appId}/domains [get]
func (h *DomainHandler) GetDomainsByApp(c *gin.Context) {
	appIDStr := c.Param("appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		h.logger.Warn("Invalid application ID format", zap.String("appID", appIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetDomainsByApp")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	domains, err := h.domainService.GetDomainsByApp(c.Request.Context(), userIDUUID, appID)
	if err != nil {
		h.logger.Error("Failed to get domains by application ID", zap.Error(err), zap.String("appID", appIDStr))
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve domains"})
		return
	}

	c.JSON(http.StatusOK, domains)
}

// GetDomain godoc
// @Summary Get domain details
// @Description Get detailed information about a specific domain
// @Tags domains
// @Produce json
// @Security BearerAuth
// @Param domainId path string true "Domain ID"
// @Success 200 {object} domain.DomainResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /domains/{domainId} [get]
func (h *DomainHandler) GetDomain(c *gin.Context) {
	domainIDStr := c.Param("domainId")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		h.logger.Warn("Invalid domain ID format", zap.String("domainID", domainIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetDomain")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	domain, err := h.domainService.GetDomain(c.Request.Context(), userIDUUID, domainID)
	if err != nil {
		h.logger.Error("Failed to get domain", zap.Error(err), zap.String("domainID", domainIDStr))
		if strings.Contains(err.Error(), "domain not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve domain"})
		return
	}

	c.JSON(http.StatusOK, domain)
}

// RequestCertificate godoc
// @Summary Request certificate for a domain
// @Description Request a certificate for a domain using cert-manager
// @Tags domains
// @Produce json
// @Security BearerAuth
// @Param domainId path string true "Domain ID"
// @Success 202 {object} domain.CertificateRequestResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /domains/{domainId}/certificates [post]
func (h *DomainHandler) RequestCertificate(c *gin.Context) {
	domainIDStr := c.Param("domainId")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		h.logger.Warn("Invalid domain ID format", zap.String("domainID", domainIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for RequestCertificate")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	response, err := h.domainService.RequestCertificate(c.Request.Context(), userIDUUID, domainID)
	if err != nil {
		h.logger.Error("Failed to request certificate", zap.Error(err), zap.String("domainID", domainIDStr))
		if strings.Contains(err.Error(), "domain not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request certificate"})
		return
	}

	c.JSON(http.StatusAccepted, response)
}

// GetCertificateStatus godoc
// @Summary Get certificate status for a domain
// @Description Get the current certificate status and information for a domain
// @Tags domains
// @Produce json
// @Security BearerAuth
// @Param domainId path string true "Domain ID"
// @Success 200 {object} domain.CertificateInfo
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /domains/{domainId}/certificates [get]
func (h *DomainHandler) GetCertificateStatus(c *gin.Context) {
	domainIDStr := c.Param("domainId")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		h.logger.Warn("Invalid domain ID format", zap.String("domainID", domainIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetCertificateStatus")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	certInfo, err := h.domainService.GetCertificateStatus(c.Request.Context(), userIDUUID, domainID)
	if err != nil {
		h.logger.Error("Failed to get certificate status", zap.Error(err), zap.String("domainID", domainIDStr))
		if strings.Contains(err.Error(), "domain not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve certificate status"})
		return
	}

	c.JSON(http.StatusOK, certInfo)
}

// DeleteDomain godoc
// @Summary Delete a domain
// @Description Delete a domain and its associated resources
// @Tags domains
// @Security BearerAuth
// @Param domainId path string true "Domain ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /domains/{domainId} [delete]
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	domainIDStr := c.Param("domainId")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		h.logger.Warn("Invalid domain ID format", zap.String("domainID", domainIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for DeleteDomain")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	err = h.domainService.DeleteDomain(c.Request.Context(), userIDUUID, domainID)
	if err != nil {
		h.logger.Error("Failed to delete domain", zap.Error(err), zap.String("domainID", domainIDStr))
		if strings.Contains(err.Error(), "domain not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Domain not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete domain"})
		return
	}

	c.Status(http.StatusNoContent)
}
