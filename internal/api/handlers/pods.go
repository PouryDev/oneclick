package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/domain"
)

type PodHandler struct {
	podService services.PodService
	logger     *zap.Logger
	upgrader   websocket.Upgrader
}

func NewPodHandler(podService services.PodService, logger *zap.Logger) *PodHandler {
	return &PodHandler{
		podService: podService,
		logger:     logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// GetPodsByApp godoc
// @Summary Get pods for an application
// @Description Get a list of pods for a specific application
// @Tags pods
// @Produce json
// @Security BearerAuth
// @Param appId path string true "Application ID"
// @Success 200 {object} domain.PodListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apps/{appId}/pods [get]
func (h *PodHandler) GetPodsByApp(c *gin.Context) {
	appIDStr := c.Param("appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		h.logger.Warn("Invalid application ID format", zap.String("appID", appIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetPodsByApp")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	pods, err := h.podService.GetPodsByApp(c.Request.Context(), userIDUUID, appID)
	if err != nil {
		h.logger.Error("Failed to get pods by application ID", zap.Error(err), zap.String("appID", appIDStr))
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pods"})
		return
	}

	// Create audit log entry
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	go func() {
		h.podService.CreateAuditLog(c.Request.Context(), userIDUUID, "pod_list", "", "", uuid.Nil, appID, ipAddress, userAgent)
	}()

	c.JSON(http.StatusOK, domain.PodListResponse{
		Pods:  pods,
		Total: len(pods),
	})
}

// GetPodDetail godoc
// @Summary Get pod details
// @Description Get detailed information about a specific pod
// @Tags pods
// @Produce json
// @Security BearerAuth
// @Param podId path string true "Pod ID"
// @Param namespace query string true "Pod namespace"
// @Success 200 {object} domain.PodDetail
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pods/{podId} [get]
func (h *PodHandler) GetPodDetail(c *gin.Context) {
	podIDStr := c.Param("podId")
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace parameter is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetPodDetail")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	podDetail, err := h.podService.GetPodDetail(c.Request.Context(), userIDUUID, podIDStr, namespace)
	if err != nil {
		h.logger.Error("Failed to get pod detail", zap.Error(err), zap.String("podID", podIDStr))
		if strings.Contains(err.Error(), "pod not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pod details"})
		return
	}

	// Create audit log entry
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	go func() {
		h.podService.CreateAuditLog(c.Request.Context(), userIDUUID, "pod_detail", podIDStr, namespace, uuid.Nil, uuid.Nil, ipAddress, userAgent)
	}()

	c.JSON(http.StatusOK, podDetail)
}

// GetPodLogs godoc
// @Summary Get pod logs
// @Description Get logs from a specific pod
// @Tags pods
// @Produce json
// @Security BearerAuth
// @Param podId path string true "Pod ID"
// @Param namespace query string true "Pod namespace"
// @Param container query string false "Container name"
// @Param follow query bool false "Follow logs"
// @Param tailLines query int false "Number of lines to tail"
// @Success 200 {object} domain.PodLogsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pods/{podId}/logs [get]
func (h *PodHandler) GetPodLogs(c *gin.Context) {
	podIDStr := c.Param("podId")
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace parameter is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetPodLogs")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	// Parse query parameters
	var req domain.PodLogsRequest
	req.Container = c.Query("container")
	req.Follow = c.Query("follow") == "true"

	if tailLinesStr := c.Query("tailLines"); tailLinesStr != "" {
		tailLines, err := strconv.ParseInt(tailLinesStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tailLines parameter"})
			return
		}
		req.TailLines = tailLines
	} else {
		req.TailLines = 100 // Default
	}

	podLogs, err := h.podService.GetPodLogs(c.Request.Context(), userIDUUID, podIDStr, namespace, req)
	if err != nil {
		h.logger.Error("Failed to get pod logs", zap.Error(err), zap.String("podID", podIDStr))
		if strings.Contains(err.Error(), "pod not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pod logs"})
		return
	}

	// Create audit log entry
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	go func() {
		h.podService.CreateAuditLog(c.Request.Context(), userIDUUID, "pod_logs", podIDStr, namespace, uuid.Nil, uuid.Nil, ipAddress, userAgent)
	}()

	c.JSON(http.StatusOK, podLogs)
}

// GetPodDescribe godoc
// @Summary Get pod describe information
// @Description Get kubectl describe style information for a pod
// @Tags pods
// @Produce json
// @Security BearerAuth
// @Param podId path string true "Pod ID"
// @Param namespace query string true "Pod namespace"
// @Success 200 {object} domain.PodDescribeResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pods/{podId}/describe [get]
func (h *PodHandler) GetPodDescribe(c *gin.Context) {
	podIDStr := c.Param("podId")
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace parameter is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for GetPodDescribe")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	podDescribe, err := h.podService.GetPodDescribe(c.Request.Context(), userIDUUID, podIDStr, namespace)
	if err != nil {
		h.logger.Error("Failed to get pod describe", zap.Error(err), zap.String("podID", podIDStr))
		if strings.Contains(err.Error(), "pod not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pod not found"})
			return
		}
		if strings.Contains(err.Error(), "user is not a member") || strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pod describe information"})
		return
	}

	// Create audit log entry
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	go func() {
		h.podService.CreateAuditLog(c.Request.Context(), userIDUUID, "pod_describe", podIDStr, namespace, uuid.Nil, uuid.Nil, ipAddress, userAgent)
	}()

	c.JSON(http.StatusOK, podDescribe)
}

// ExecInPod godoc
// @Summary Execute command in pod terminal
// @Description Execute a command in a pod and establish websocket connection for terminal
// @Tags pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param podId path string true "Pod ID"
// @Param namespace query string true "Pod namespace"
// @Param request body domain.PodExecRequest true "Exec request"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pods/{podId}/terminal [post]
func (h *PodHandler) ExecInPod(c *gin.Context) {
	podIDStr := c.Param("podId")
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace parameter is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context for ExecInPod")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		h.logger.Error("Invalid user ID in context", zap.Any("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var req domain.PodExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for ExecInPod", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Upgrade to websocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to websocket", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to websocket"})
		return
	}
	defer conn.Close()

	// Create audit log entry
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	go func() {
		h.podService.CreateAuditLog(c.Request.Context(), userIDUUID, "pod_exec", podIDStr, namespace, uuid.Nil, uuid.Nil, ipAddress, userAgent)
	}()

	// Execute command in pod
	err = h.podService.ExecInPod(c.Request.Context(), userIDUUID, podIDStr, namespace, req, conn)
	if err != nil {
		h.logger.Error("Failed to exec in pod", zap.Error(err), zap.String("podID", podIDStr))
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}
}
