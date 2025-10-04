package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/domain"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	eventLoggerService services.EventLoggerService
	dashboardService   services.DashboardService
	readModelService   services.ReadModelService
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	eventLoggerService services.EventLoggerService,
	dashboardService services.DashboardService,
	readModelService services.ReadModelService,
) *EventHandler {
	return &EventHandler{
		eventLoggerService: eventLoggerService,
		dashboardService:   dashboardService,
		readModelService:   readModelService,
	}
}

// GetEventsByOrgID handles GET /orgs/{orgId}/events
func (h *EventHandler) GetEventsByOrgID(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get optional filters
	action := c.Query("action")
	resourceType := c.Query("resource_type")

	var events []*domain.EventLog
	var serviceErr error

	if action != "" {
		events, serviceErr = h.eventLoggerService.GetEventsByOrgIDAndAction(
			c.Request.Context(),
			userUUID,
			orgID,
			domain.EventAction(action),
			limit,
			offset,
		)
	} else if resourceType != "" {
		events, serviceErr = h.eventLoggerService.GetEventsByOrgIDAndResourceType(
			c.Request.Context(),
			userUUID,
			orgID,
			domain.ResourceType(resourceType),
			limit,
			offset,
		)
	} else {
		events, serviceErr = h.eventLoggerService.GetEventsByOrgID(
			c.Request.Context(),
			userUUID,
			orgID,
			limit,
			offset,
		)
	}

	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}

	// Convert to response format
	var eventResponses []*domain.EventResponse
	for _, event := range events {
		eventResponses = append(eventResponses, event.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"events": eventResponses,
		"limit":  limit,
		"offset": offset,
		"count":  len(eventResponses),
	})
}

// GetEventByID handles GET /events/{eventId}
func (h *EventHandler) GetEventByID(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get event ID from URL
	eventIDStr := c.Param("eventId")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get event
	event, err := h.eventLoggerService.GetEventByID(c.Request.Context(), userUUID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event.ToResponse())
}

// GetDashboardCounts handles GET /orgs/{orgId}/dashboard
func (h *EventHandler) GetDashboardCounts(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get dashboard counts
	counts, err := h.dashboardService.GetDashboardCounts(c.Request.Context(), userUUID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if counts == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard counts not found"})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// UpdateDashboardCounts handles POST /orgs/{orgId}/dashboard/refresh
func (h *EventHandler) UpdateDashboardCounts(c *gin.Context) {
	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Update dashboard counts
	counts, err := h.dashboardService.UpdateDashboardCounts(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// GetReadModelProject handles GET /orgs/{orgId}/readmodel/{key}
func (h *EventHandler) GetReadModelProject(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get key from URL
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	// Get read model project
	project, err := h.readModelService.GetReadModelProject(c.Request.Context(), userUUID, orgID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Read model project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// GetReadModelProjects handles GET /orgs/{orgId}/readmodel
func (h *EventHandler) GetReadModelProjects(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get read model projects
	projects, err := h.readModelService.GetReadModelProjectsByOrgID(c.Request.Context(), userUUID, orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"count":    len(projects),
	})
}

// CreateReadModelProject handles POST /orgs/{orgId}/readmodel
func (h *EventHandler) CreateReadModelProject(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse request body
	var req struct {
		Key   string                 `json:"key" binding:"required"`
		Value map[string]interface{} `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create read model project
	project, err := h.readModelService.CreateReadModelProject(
		c.Request.Context(),
		userUUID,
		orgID,
		req.Key,
		req.Value,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// DeleteReadModelProject handles DELETE /orgs/{orgId}/readmodel/{key}
func (h *EventHandler) DeleteReadModelProject(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get organization ID from URL
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Get key from URL
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	// Delete read model project
	err = h.readModelService.DeleteReadModelProject(c.Request.Context(), userUUID, orgID, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
