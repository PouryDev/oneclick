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

// PipelineHandler handles pipeline-related HTTP requests
type PipelineHandler struct {
	pipelineService services.PipelineService
	logger          *zap.Logger
}

// NewPipelineHandler creates a new pipeline handler
func NewPipelineHandler(pipelineService services.PipelineService, logger *zap.Logger) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
		logger:          logger,
	}
}

// CreatePipeline handles POST /apps/{appId}/pipelines
func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
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

	// Parse request body
	var req domain.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Ref == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ref is required"})
		return
	}
	if req.CommitSHA == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "commit_sha is required"})
		return
	}

	// Create pipeline
	pipeline, err := h.pipelineService.CreatePipeline(c.Request.Context(), userID, appID, req)
	if err != nil {
		h.logger.Error("Failed to create pipeline", zap.Error(err), zap.String("appID", appID.String()))

		// Check for specific error types
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}

	c.JSON(http.StatusCreated, pipeline)
}

// GetPipelinesByApp handles GET /apps/{appId}/pipelines
func (h *PipelineHandler) GetPipelinesByApp(c *gin.Context) {
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

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get pipelines
	pipelines, err := h.pipelineService.GetPipelinesByApp(c.Request.Context(), userID, appID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get pipelines", zap.Error(err), zap.String("appID", appID.String()))

		// Check for specific error types
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pipelines"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pipelines": pipelines,
		"count":     len(pipelines),
		"limit":     limit,
		"offset":    offset,
	})
}

// GetPipelineByID handles GET /pipelines/{pipelineId}
func (h *PipelineHandler) GetPipelineByID(c *gin.Context) {
	// Parse pipeline ID
	pipelineIDStr := c.Param("pipelineId")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.logger.Error("Invalid pipeline ID", zap.Error(err), zap.String("pipelineID", pipelineIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
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

	// Get pipeline
	pipeline, err := h.pipelineService.GetPipelineByID(c.Request.Context(), userID, pipelineID)
	if err != nil {
		h.logger.Error("Failed to get pipeline", zap.Error(err), zap.String("pipelineID", pipelineID.String()))

		// Check for specific error types
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "pipeline not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pipeline"})
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

// GetPipelineLogs handles GET /pipelines/{pipelineId}/logs
func (h *PipelineHandler) GetPipelineLogs(c *gin.Context) {
	// Parse pipeline ID
	pipelineIDStr := c.Param("pipelineId")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.logger.Error("Invalid pipeline ID", zap.Error(err), zap.String("pipelineID", pipelineIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
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

	// Get pipeline logs
	logs, err := h.pipelineService.GetPipelineLogs(c.Request.Context(), userID, pipelineID)
	if err != nil {
		h.logger.Error("Failed to get pipeline logs", zap.Error(err), zap.String("pipelineID", pipelineID.String()))

		// Check for specific error types
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "pipeline not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pipeline logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// TriggerPipeline handles POST /apps/{appId}/pipelines/trigger
func (h *PipelineHandler) TriggerPipeline(c *gin.Context) {
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

	// Parse request body
	var req domain.CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Ref == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ref is required"})
		return
	}
	if req.CommitSHA == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "commit_sha is required"})
		return
	}

	// Trigger pipeline
	pipeline, err := h.pipelineService.TriggerPipeline(c.Request.Context(), userID, appID, req)
	if err != nil {
		h.logger.Error("Failed to trigger pipeline", zap.Error(err), zap.String("appID", appID.String()))

		// Check for specific error types
		if err.Error() == "unauthorized: user is not a member of the organization" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger pipeline"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Pipeline triggered successfully",
		"pipeline": pipeline,
	})
}
