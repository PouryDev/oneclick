package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/domain"
)

type ClusterHandler struct {
	clusterService services.ClusterService
	validator      *validator.Validate
}

func NewClusterHandler(clusterService services.ClusterService) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
		validator:      validator.New(),
	}
}

// CreateCluster godoc
// @Summary Create a new cluster
// @Description Create a new cluster in the organization
// @Tags clusters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body domain.CreateClusterRequest true "Cluster creation data"
// @Success 201 {object} domain.ClusterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/clusters [post]
func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req domain.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster, err := h.clusterService.CreateCluster(c.Request.Context(), userUUID, orgID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "invalid kubeconfig") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cluster"})
		return
	}

	c.JSON(http.StatusCreated, cluster)
}

// ImportCluster godoc
// @Summary Import a cluster
// @Description Import an existing cluster by uploading kubeconfig
// @Tags clusters
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param name formData string true "Cluster name"
// @Param kubeconfig formData file true "Kubeconfig file"
// @Success 201 {object} domain.ClusterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/clusters/import [post]
func (h *ClusterHandler) ImportCluster(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get form data
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster name is required"})
		return
	}

	// Get uploaded file
	file, _, err := c.Request.FormFile("kubeconfig")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kubeconfig file is required"})
		return
	}
	defer file.Close()

	// Read file content
	kubeconfigData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read kubeconfig file"})
		return
	}

	cluster, err := h.clusterService.ImportCluster(c.Request.Context(), userUUID, orgID, name, kubeconfigData)
	if err != nil {
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "invalid kubeconfig") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import cluster"})
		return
	}

	c.JSON(http.StatusCreated, cluster)
}

// GetClustersByOrg godoc
// @Summary Get organization clusters
// @Description Get list of clusters in the organization
// @Tags clusters
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {array} domain.ClusterSummary
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/clusters [get]
func (h *ClusterHandler) GetClustersByOrg(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	clusters, err := h.clusterService.GetClustersByOrg(c.Request.Context(), userUUID, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get clusters"})
		return
	}

	c.JSON(http.StatusOK, clusters)
}

// GetCluster godoc
// @Summary Get cluster details
// @Description Get detailed information about a cluster
// @Tags clusters
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "Cluster ID"
// @Success 200 {object} domain.ClusterDetailResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clusters/{clusterId} [get]
func (h *ClusterHandler) GetCluster(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	clusterIDStr := c.Param("clusterId")
	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), userUUID, clusterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
			return
		}
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cluster"})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

// GetClusterHealth godoc
// @Summary Get cluster health status
// @Description Get real-time cluster health information
// @Tags clusters
// @Produce json
// @Security BearerAuth
// @Param clusterId path string true "Cluster ID"
// @Success 200 {object} domain.ClusterHealth
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clusters/{clusterId}/status [get]
func (h *ClusterHandler) GetClusterHealth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	clusterIDStr := c.Param("clusterId")
	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	health, err := h.clusterService.GetClusterHealth(c.Request.Context(), userUUID, clusterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
			return
		}
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "does not have kubeconfig") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cluster does not have kubeconfig"})
			return
		}
		if strings.Contains(err.Error(), "failed to get cluster health") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to cluster"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cluster health"})
		return
	}

	c.JSON(http.StatusOK, health)
}

// DeleteCluster godoc
// @Summary Delete cluster
// @Description Delete a cluster (only admins and owners can delete)
// @Tags clusters
// @Security BearerAuth
// @Param clusterId path string true "Cluster ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clusters/{clusterId} [delete]
func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	clusterIDStr := c.Param("clusterId")
	clusterID, err := uuid.Parse(clusterIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cluster ID"})
		return
	}

	err = h.clusterService.DeleteCluster(c.Request.Context(), userUUID, clusterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
			return
		}
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete cluster"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cluster"})
		return
	}

	c.Status(http.StatusNoContent)
}
