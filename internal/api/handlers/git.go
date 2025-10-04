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

type GitServerHandler struct {
	gitServerService services.GitServerService
	logger           *zap.Logger
	validator        *validator.Validate
}

type RunnerHandler struct {
	runnerService services.RunnerService
	logger        *zap.Logger
	validator     *validator.Validate
}

type JobHandler struct {
	jobService services.JobService
	logger     *zap.Logger
}

func NewGitServerHandler(gitServerService services.GitServerService, logger *zap.Logger) *GitServerHandler {
	return &GitServerHandler{
		gitServerService: gitServerService,
		logger:           logger,
		validator:        validator.New(),
	}
}

func NewRunnerHandler(runnerService services.RunnerService, logger *zap.Logger) *RunnerHandler {
	return &RunnerHandler{
		runnerService: runnerService,
		logger:        logger,
		validator:     validator.New(),
	}
}

func NewJobHandler(jobService services.JobService, logger *zap.Logger) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		logger:     logger,
	}
}

// GitServer handlers

// CreateGitServer godoc
// @Summary Create a git server
// @Description Create a new self-hosted git server (Gitea) for the organization
// @Tags git-servers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body domain.CreateGitServerRequest true "Git server creation request"
// @Success 201 {object} domain.GitServerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/gitservers [post]
func (h *GitServerHandler) CreateGitServer(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Warn("Invalid organization ID format", zap.String("orgID", orgIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for CreateGitServer")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var req domain.CreateGitServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for CreateGitServer", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for CreateGitServer", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.gitServerService.CreateGitServer(c.Request.Context(), userIDUUID, orgID, req)
	if err != nil {
		h.logger.Error("Failed to create git server", zap.Error(err), zap.String("orgID", orgIDStr))
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create git server"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetGitServersByOrg godoc
// @Summary Get git servers for an organization
// @Description Get a list of git servers for a specific organization
// @Tags git-servers
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {array} domain.GitServerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/gitservers [get]
func (h *GitServerHandler) GetGitServersByOrg(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Warn("Invalid organization ID format", zap.String("orgID", orgIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetGitServersByOrg")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	gitServers, err := h.gitServerService.GetGitServersByOrg(c.Request.Context(), userIDUUID, orgID)
	if err != nil {
		h.logger.Error("Failed to get git servers by organization ID", zap.Error(err), zap.String("orgID", orgIDStr))
		if strings.Contains(err.Error(), "user does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve git servers"})
		return
	}

	c.JSON(http.StatusOK, gitServers)
}

// GetGitServer godoc
// @Summary Get git server details
// @Description Get detailed information about a specific git server
// @Tags git-servers
// @Produce json
// @Security BearerAuth
// @Param gitServerId path string true "Git Server ID"
// @Success 200 {object} domain.GitServerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /gitservers/{gitServerId} [get]
func (h *GitServerHandler) GetGitServer(c *gin.Context) {
	gitServerIDStr := c.Param("gitServerId")
	gitServerID, err := uuid.Parse(gitServerIDStr)
	if err != nil {
		h.logger.Warn("Invalid git server ID format", zap.String("gitServerID", gitServerIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid git server ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetGitServer")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	gitServer, err := h.gitServerService.GetGitServer(c.Request.Context(), userIDUUID, gitServerID)
	if err != nil {
		h.logger.Error("Failed to get git server", zap.Error(err), zap.String("gitServerID", gitServerIDStr))
		if strings.Contains(err.Error(), "git server not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Git server not found"})
			return
		}
		if strings.Contains(err.Error(), "user does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve git server"})
		return
	}

	c.JSON(http.StatusOK, gitServer)
}

// DeleteGitServer godoc
// @Summary Delete a git server
// @Description Delete a git server and its resources
// @Tags git-servers
// @Security BearerAuth
// @Param gitServerId path string true "Git Server ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /gitservers/{gitServerId} [delete]
func (h *GitServerHandler) DeleteGitServer(c *gin.Context) {
	gitServerIDStr := c.Param("gitServerId")
	gitServerID, err := uuid.Parse(gitServerIDStr)
	if err != nil {
		h.logger.Warn("Invalid git server ID format", zap.String("gitServerID", gitServerIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid git server ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for DeleteGitServer")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	err = h.gitServerService.DeleteGitServer(c.Request.Context(), userIDUUID, gitServerID)
	if err != nil {
		h.logger.Error("Failed to delete git server", zap.Error(err), zap.String("gitServerID", gitServerIDStr))
		if strings.Contains(err.Error(), "git server not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Git server not found"})
			return
		}
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete git server"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Runner handlers

// CreateRunner godoc
// @Summary Create a CI runner
// @Description Create a new CI runner for the organization
// @Tags runners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body domain.CreateRunnerRequest true "Runner creation request"
// @Success 201 {object} domain.RunnerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/runners [post]
func (h *RunnerHandler) CreateRunner(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Warn("Invalid organization ID format", zap.String("orgID", orgIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for CreateRunner")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var req domain.CreateRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for CreateRunner", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("Validation failed for CreateRunner", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.runnerService.CreateRunner(c.Request.Context(), userIDUUID, orgID, req)
	if err != nil {
		h.logger.Error("Failed to create runner", zap.Error(err), zap.String("orgID", orgIDStr))
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create runner"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetRunnersByOrg godoc
// @Summary Get runners for an organization
// @Description Get a list of CI runners for a specific organization
// @Tags runners
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {array} domain.RunnerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/runners [get]
func (h *RunnerHandler) GetRunnersByOrg(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Warn("Invalid organization ID format", zap.String("orgID", orgIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetRunnersByOrg")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	runners, err := h.runnerService.GetRunnersByOrg(c.Request.Context(), userIDUUID, orgID)
	if err != nil {
		h.logger.Error("Failed to get runners by organization ID", zap.Error(err), zap.String("orgID", orgIDStr))
		if strings.Contains(err.Error(), "user does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runners"})
		return
	}

	c.JSON(http.StatusOK, runners)
}

// GetRunner godoc
// @Summary Get runner details
// @Description Get detailed information about a specific CI runner
// @Tags runners
// @Produce json
// @Security BearerAuth
// @Param runnerId path string true "Runner ID"
// @Success 200 {object} domain.RunnerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /runners/{runnerId} [get]
func (h *RunnerHandler) GetRunner(c *gin.Context) {
	runnerIDStr := c.Param("runnerId")
	runnerID, err := uuid.Parse(runnerIDStr)
	if err != nil {
		h.logger.Warn("Invalid runner ID format", zap.String("runnerID", runnerIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetRunner")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	runner, err := h.runnerService.GetRunner(c.Request.Context(), userIDUUID, runnerID)
	if err != nil {
		h.logger.Error("Failed to get runner", zap.Error(err), zap.String("runnerID", runnerIDStr))
		if strings.Contains(err.Error(), "runner not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
		if strings.Contains(err.Error(), "user does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve runner"})
		return
	}

	c.JSON(http.StatusOK, runner)
}

// DeleteRunner godoc
// @Summary Delete a CI runner
// @Description Delete a CI runner and its resources
// @Tags runners
// @Security BearerAuth
// @Param runnerId path string true "Runner ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /runners/{runnerId} [delete]
func (h *RunnerHandler) DeleteRunner(c *gin.Context) {
	runnerIDStr := c.Param("runnerId")
	runnerID, err := uuid.Parse(runnerIDStr)
	if err != nil {
		h.logger.Warn("Invalid runner ID format", zap.String("runnerID", runnerIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for DeleteRunner")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	err = h.runnerService.DeleteRunner(c.Request.Context(), userIDUUID, runnerID)
	if err != nil {
		h.logger.Error("Failed to delete runner", zap.Error(err), zap.String("runnerID", runnerIDStr))
		if strings.Contains(err.Error(), "runner not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
		if strings.Contains(err.Error(), "user does not have access") || strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete runner"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Job handlers

// GetJobsByOrg godoc
// @Summary Get jobs for an organization
// @Description Get a list of background jobs for a specific organization
// @Tags jobs
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {array} domain.JobResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/jobs [get]
func (h *JobHandler) GetJobsByOrg(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Warn("Invalid organization ID format", zap.String("orgID", orgIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetJobsByOrg")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	jobs, err := h.jobService.GetJobsByOrg(c.Request.Context(), userIDUUID, orgID)
	if err != nil {
		h.logger.Error("Failed to get jobs by organization ID", zap.Error(err), zap.String("orgID", orgIDStr))
		if strings.Contains(err.Error(), "user does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve jobs"})
		return
	}

	c.JSON(http.StatusOK, jobs)
}
