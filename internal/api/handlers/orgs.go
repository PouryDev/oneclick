package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/app/services"
	"github.com/PouryDev/oneclick/internal/domain"
)

type OrganizationHandler struct {
	orgService services.OrganizationService
	validator  *validator.Validate
}

func NewOrganizationHandler(orgService services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
		validator:  validator.New(),
	}
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Create a new organization and add the creator as owner
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateOrganizationRequest true "Organization creation data"
// @Success 201 {object} domain.OrganizationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
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

	var req domain.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Additional validation
	if len(strings.TrimSpace(req.Name)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name cannot be empty"})
		return
	}

	org, err := h.orgService.CreateOrganization(c.Request.Context(), userUUID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// GetUserOrganizations godoc
// @Summary Get user's organizations
// @Description Get list of organizations the current user belongs to
// @Tags organizations
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.UserOrganizationSummary
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs [get]
func (h *OrganizationHandler) GetUserOrganizations(c *gin.Context) {
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

	orgs, err := h.orgService.GetUserOrganizations(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organizations"})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

// GetOrganization godoc
// @Summary Get organization details
// @Description Get organization details including members and counts
// @Tags organizations
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 200 {object} domain.OrganizationWithMembers
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
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

	org, err := h.orgService.GetOrganization(c.Request.Context(), userUUID, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "does not have access") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// AddMember godoc
// @Summary Add member to organization
// @Description Add a new member to the organization by email
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param request body domain.AddMemberRequest true "Member data"
// @Success 201 {object} domain.OrganizationMember
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/members [post]
func (h *OrganizationHandler) AddMember(c *gin.Context) {
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

	var req domain.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.orgService.AddMember(c.Request.Context(), userUUID, orgID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if strings.Contains(err.Error(), "already a member") {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// UpdateMemberRole godoc
// @Summary Update member role
// @Description Update a member's role in the organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param userId path string true "User ID"
// @Param request body domain.UpdateMemberRoleRequest true "Role update data"
// @Success 200 {object} domain.OrganizationMember
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/members/{userId} [patch]
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
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

	memberIDStr := c.Param("userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req domain.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.orgService.UpdateMemberRole(c.Request.Context(), userUUID, orgID, memberID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		if strings.Contains(err.Error(), "only owners can") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "cannot remove the last owner") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// RemoveMember godoc
// @Summary Remove member from organization
// @Description Remove a member from the organization
// @Tags organizations
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Param userId path string true "User ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId}/members/{userId} [delete]
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
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

	memberIDStr := c.Param("userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.orgService.RemoveMember(c.Request.Context(), userUUID, orgID, memberID)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		if strings.Contains(err.Error(), "only owners can") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "cannot remove the last owner") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteOrganization godoc
// @Summary Delete organization
// @Description Delete the organization (only owners can delete)
// @Tags organizations
// @Security BearerAuth
// @Param orgId path string true "Organization ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orgs/{orgId} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
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

	err = h.orgService.DeleteOrganization(c.Request.Context(), userUUID, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "insufficient permissions") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization"})
		return
	}

	c.Status(http.StatusNoContent)
}
