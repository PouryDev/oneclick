package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// OrganizationAccessMiddleware checks if the user has access to the organization
func OrganizationAccessMiddleware(orgRepo repo.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		userUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		orgIDStr := c.Param("orgId")
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			c.Abort()
			return
		}

		// Check if user has access to the organization
		role, err := orgRepo.GetUserRoleInOrganization(c.Request.Context(), userUUID, orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check organization access"})
			c.Abort()
			return
		}

		if role == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to organization"})
			c.Abort()
			return
		}

		// Set user role in context for use in handlers
		c.Set("user_role", role)
		c.Set("org_id", orgID)
		c.Next()
	}
}

// RequireRoleMiddleware checks if the user has the required role
func RequireRoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnerMiddleware checks if the user is an owner
func RequireOwnerMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware(domain.RoleOwner)
}

// RequireAdminOrOwnerMiddleware checks if the user is an admin or owner
func RequireAdminOrOwnerMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware(domain.RoleAdmin, domain.RoleOwner)
}

// RequireMemberMiddleware checks if the user is a member (any role)
func RequireMemberMiddleware() gin.HandlerFunc {
	return RequireRoleMiddleware(domain.RoleMember, domain.RoleAdmin, domain.RoleOwner)
}
