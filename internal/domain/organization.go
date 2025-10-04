package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization entity
type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserOrganization represents the relationship between a user and an organization
type UserOrganization struct {
	UserID    uuid.UUID `json:"user_id"`
	OrgID     uuid.UUID `json:"org_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrganizationWithMembers represents an organization with its members
type OrganizationWithMembers struct {
	Organization
	Members       []OrganizationMember `json:"members"`
	ClustersCount int                  `json:"clusters_count"`
	AppsCount     int                  `json:"apps_count"`
}

// UserOrganizationSummary represents a user's view of an organization
type UserOrganizationSummary struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
	ClustersCount int       `json:"clusters_count"`
	AppsCount     int       `json:"apps_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Request/Response DTOs
type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

type AddMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=owner admin member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin member"`
}

type OrganizationResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Valid roles
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// ValidRoles returns a list of valid organization roles
func ValidRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleMember}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	for _, validRole := range ValidRoles() {
		if role == validRole {
			return true
		}
	}
	return false
}

// CanManageMembers checks if a role can manage organization members
func CanManageMembers(role string) bool {
	return role == RoleOwner || role == RoleAdmin
}

// CanDeleteOrganization checks if a role can delete the organization
func CanDeleteOrganization(role string) bool {
	return role == RoleOwner
}
