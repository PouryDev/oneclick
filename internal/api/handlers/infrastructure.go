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

type InfrastructureHandler struct {
	infrastructureService services.InfrastructureService
	logger                *zap.Logger
	validator             *validator.Validate
}

func NewInfrastructureHandler(infrastructureService services.InfrastructureService, logger *zap.Logger) *InfrastructureHandler {
	return &InfrastructureHandler{
		infrastructureService: infrastructureService,
		logger:                logger,
		validator:             validator.New(),
	}
}

// ProvisionServices godoc
// @Summary Provision services for an application
// @Description Parse infra-config.yml and provision services using Helm
// @Tags infrastructure
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appId path string true "Application ID"
// @Param request body domain.ProvisionServiceRequest true "Service provisioning request"
// @Success 202 {object} domain.ProvisionServiceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{appId}/infra/provision [post]
func (h *InfrastructureHandler) ProvisionServices(c *gin.Context) {
	appIDStr := c.Param("appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		h.logger.Warn("Invalid application ID format", zap.String("appID", appIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for ProvisionServices")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var req domain.ProvisionServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for ProvisionServices", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for ProvisionServices", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.InfraConfig == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "infra_config is required"})
		return
	}

	response, err := h.infrastructureService.ProvisionServices(c.Request.Context(), userIDUUID, appID, req.InfraConfig)
	if err != nil {
		h.logger.Error("Failed to provision services", zap.Error(err), zap.String("appID", appIDStr))
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "application not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}
		if strings.Contains(err.Error(), "failed to parse") || strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to provision services"})
		return
	}

	c.JSON(http.StatusAccepted, response) // 202 Accepted as provisioning is async
}

// GetServicesByApp godoc
// @Summary Get services for an application
// @Description Get a list of provisioned services for a specific application
// @Tags infrastructure
// @Produce json
// @Security BearerAuth
// @Param appId path string true "Application ID"
// @Success 200 {array} domain.ServiceDetail
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{appId}/infra/services [get]
func (h *InfrastructureHandler) GetServicesByApp(c *gin.Context) {
	appIDStr := c.Param("appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		h.logger.Warn("Invalid application ID format", zap.String("appID", appIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetServicesByApp")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	services, err := h.infrastructureService.GetServicesByApp(c.Request.Context(), userIDUUID, appID)
	if err != nil {
		h.logger.Error("Failed to get services by application ID", zap.Error(err), zap.String("appID", appIDStr))
		if strings.Contains(err.Error(), "user does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "application not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve services"})
		return
	}

	c.JSON(http.StatusOK, services)
}

// GetServiceConfig godoc
// @Summary Get service configuration details
// @Description Get detailed configuration for a specific service config (reveals secrets for authorized users)
// @Tags infrastructure
// @Produce json
// @Security BearerAuth
// @Param configId path string true "Service Config ID"
// @Success 200 {object} domain.ServiceConfigRevealResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/{configId}/config [get]
func (h *InfrastructureHandler) GetServiceConfig(c *gin.Context) {
	configIDStr := c.Param("configId")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		h.logger.Warn("Invalid service config ID format", zap.String("configID", configIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service config ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetServiceConfig")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	response, err := h.infrastructureService.GetServiceConfig(c.Request.Context(), userIDUUID, configID)
	if err != nil {
		h.logger.Error("Failed to get service configuration", zap.Error(err), zap.String("configID", configIDStr))
		if strings.Contains(err.Error(), "service configuration not found") || strings.Contains(err.Error(), "service not found") || strings.Contains(err.Error(), "application not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service configuration not found"})
			return
		}
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve service configuration"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UnprovisionService godoc
// @Summary Unprovision a service
// @Description Remove a provisioned service and its resources
// @Tags infrastructure
// @Security BearerAuth
// @Param serviceId path string true "Service ID"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /services/{serviceId} [delete]
func (h *InfrastructureHandler) UnprovisionService(c *gin.Context) {
	serviceIDStr := c.Param("serviceId")
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		h.logger.Warn("Invalid service ID format", zap.String("serviceID", serviceIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for UnprovisionService")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	err = h.infrastructureService.UnprovisionService(c.Request.Context(), userIDUUID, serviceID)
	if err != nil {
		h.logger.Error("Failed to unprovision service", zap.Error(err), zap.String("serviceID", serviceIDStr))
		if strings.Contains(err.Error(), "service not found") || strings.Contains(err.Error(), "application not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unprovision service"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Service unprovisioning initiated"}) // 202 Accepted as unprovisioning is async
}
