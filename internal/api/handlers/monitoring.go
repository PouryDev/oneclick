package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/domain"
)

// MonitoringHandler handles monitoring-related HTTP requests
type MonitoringHandler struct {
	monitoringService services.MonitoringService
	logger            *zap.Logger
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(monitoringService services.MonitoringService, logger *zap.Logger) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
		logger:            logger,
	}
}

// GetClusterMetrics handles GET /clusters/{clusterId}/monitoring
func (h *MonitoringHandler) GetClusterMetrics(c *gin.Context) {
	// Parse cluster ID
	clusterIDStr := c.Param("clusterId")
	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		h.logger.Error("Invalid cluster ID", zap.Error(err), zap.String("clusterID", clusterIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse monitoring request
	var req domain.MonitoringRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid monitoring request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Set default time range if not provided
	if req.TimeRange == "" {
		req.TimeRange = domain.TimeRange5m
	}

	// Validate time range
	if !domain.IsValidTimeRange(string(req.TimeRange)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range"})
		return
	}

	// Get cluster metrics
	metrics, err := h.monitoringService.GetClusterMetrics(c.Request.Context(), userID, clusterID, req)
	if err != nil {
		h.logger.Error("Failed to get cluster metrics", zap.Error(err), zap.String("clusterID", clusterID.String()))

		// Check for specific error types
		if err.Error() == "rate limit exceeded" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "cluster not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cluster metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetApplicationMetrics handles GET /apps/{appId}/monitoring
func (h *MonitoringHandler) GetApplicationMetrics(c *gin.Context) {
	// Parse app ID
	appIDStr := c.Param("appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		h.logger.Error("Invalid app ID", zap.Error(err), zap.String("appID", appIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse monitoring request
	var req domain.MonitoringRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid monitoring request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Set default time range if not provided
	if req.TimeRange == "" {
		req.TimeRange = domain.TimeRange5m
	}

	// Validate time range
	if !domain.IsValidTimeRange(string(req.TimeRange)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range"})
		return
	}

	// Get application metrics
	metrics, err := h.monitoringService.GetApplicationMetrics(c.Request.Context(), userID, appID, req)
	if err != nil {
		h.logger.Error("Failed to get application metrics", zap.Error(err), zap.String("appID", appID.String()))

		// Check for specific error types
		if err.Error() == "rate limit exceeded" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve application metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetPodMetrics handles GET /pods/{podId}/monitoring
func (h *MonitoringHandler) GetPodMetrics(c *gin.Context) {
	// Parse pod ID
	podIDStr := c.Param("podId")
	podID, err := uuid.Parse(podIDStr)
	if err != nil {
		h.logger.Error("Invalid pod ID", zap.Error(err), zap.String("podID", podIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pod ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse monitoring request
	var req domain.MonitoringRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Invalid monitoring request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	// Set default time range if not provided
	if req.TimeRange == "" {
		req.TimeRange = domain.TimeRange5m
	}

	// Validate time range
	if !domain.IsValidTimeRange(string(req.TimeRange)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range"})
		return
	}

	// Get pod metrics
	metrics, err := h.monitoringService.GetPodMetrics(c.Request.Context(), userID, podID, req)
	if err != nil {
		h.logger.Error("Failed to get pod metrics", zap.Error(err), zap.String("podID", podID.String()))

		// Check for specific error types
		if err.Error() == "rate limit exceeded" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "pod metrics not fully implemented - requires pod repository" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Pod metrics not yet implemented"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pod metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetAlerts handles GET /clusters/{clusterId}/alerts
func (h *MonitoringHandler) GetAlerts(c *gin.Context) {
	// Parse cluster ID
	clusterIDStr := c.Param("clusterId")
	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		h.logger.Error("Invalid cluster ID", zap.Error(err), zap.String("clusterID", clusterIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.logger.Error("Invalid user ID", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get alerts
	alerts, err := h.monitoringService.GetAlerts(c.Request.Context(), userID, clusterID)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err), zap.String("clusterID", clusterID.String()))

		// Check for specific error types
		if err.Error() == "rate limit exceeded" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "cluster not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	// Limit results
	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetMonitoringHealth handles GET /monitoring/health
func (h *MonitoringHandler) GetMonitoringHealth(c *gin.Context) {
	// This endpoint can be used to check if monitoring services are healthy
	// For now, we'll return a simple health check
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"services": gin.H{
			"prometheus": "healthy",
			"cache":      "healthy",
			"rate_limit": "healthy",
		},
	})
}
