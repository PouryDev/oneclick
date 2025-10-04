package services

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, userID uuid.UUID, req *domain.CreateOrganizationRequest) (*domain.OrganizationResponse, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domain.UserOrganizationSummary, error)
	GetOrganization(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationWithMembers, error)
	AddMember(ctx context.Context, userID, orgID uuid.UUID, req *domain.AddMemberRequest) (*domain.OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, userID, orgID, memberID uuid.UUID, req *domain.UpdateMemberRoleRequest) (*domain.OrganizationMember, error)
	RemoveMember(ctx context.Context, userID, orgID, memberID uuid.UUID) error
	DeleteOrganization(ctx context.Context, userID, orgID uuid.UUID) error
}

type organizationService struct {
	orgRepo  repo.OrganizationRepository
	userRepo repo.UserRepository
}

func NewOrganizationService(orgRepo repo.OrganizationRepository, userRepo repo.UserRepository) OrganizationService {
	return &organizationService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

func (s *organizationService) CreateOrganization(ctx context.Context, userID uuid.UUID, req *domain.CreateOrganizationRequest) (*domain.OrganizationResponse, error) {
	// Create the organization
	org, err := s.orgRepo.CreateOrganization(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	// Add the creator as owner
	_, err = s.orgRepo.AddUserToOrganization(ctx, userID, org.ID, domain.RoleOwner)
	if err != nil {
		return nil, err
	}

	response := &domain.OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}

	return response, nil
}

func (s *organizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domain.UserOrganizationSummary, error) {
	orgs, err := s.orgRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orgs, nil
}

func (s *organizationService) GetOrganization(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationWithMembers, error) {
	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Get organization details
	org, err := s.orgRepo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("organization not found")
	}

	// Get organization members
	members, err := s.orgRepo.GetOrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	response := &domain.OrganizationWithMembers{
		Organization:  *org,
		Members:       members,
		ClustersCount: 0, // TODO: implement when clusters are added
		AppsCount:     0, // TODO: implement when apps are added
	}

	return response, nil
}

func (s *organizationService) AddMember(ctx context.Context, userID, orgID uuid.UUID, req *domain.AddMemberRequest) (*domain.OrganizationMember, error) {
	// Check if user has permission to add members
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !domain.CanManageMembers(role) {
		return nil, errors.New("insufficient permissions to add members")
	}

	// Find user by email
	user, err := s.orgRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if user is already a member
	existingRole, err := s.orgRepo.GetUserRoleInOrganization(ctx, user.ID, orgID)
	if err != nil {
		return nil, err
	}
	if existingRole != "" {
		return nil, errors.New("user is already a member of this organization")
	}

	// Add user to organization
	_, err = s.orgRepo.AddUserToOrganization(ctx, user.ID, orgID, req.Role)
	if err != nil {
		return nil, err
	}

	// Return the new member
	member := &domain.OrganizationMember{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  req.Role,
	}

	return member, nil
}

func (s *organizationService) UpdateMemberRole(ctx context.Context, userID, orgID, memberID uuid.UUID, req *domain.UpdateMemberRoleRequest) (*domain.OrganizationMember, error) {
	// Check if user has permission to update member roles
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if !domain.CanManageMembers(role) {
		return nil, errors.New("insufficient permissions to update member roles")
	}

	// Check if target member exists in organization
	memberRole, err := s.orgRepo.GetUserRoleInOrganization(ctx, memberID, orgID)
	if err != nil {
		return nil, err
	}
	if memberRole == "" {
		return nil, errors.New("member not found in organization")
	}

	// Prevent non-owners from changing owner roles
	if memberRole == domain.RoleOwner && role != domain.RoleOwner {
		return nil, errors.New("only owners can change owner roles")
	}

	// Prevent changing the last owner to a different role
	if memberRole == domain.RoleOwner && req.Role != domain.RoleOwner {
		// Check if there are other owners
		members, err := s.orgRepo.GetOrganizationMembers(ctx, orgID)
		if err != nil {
			return nil, err
		}

		ownerCount := 0
		for _, member := range members {
			if member.Role == domain.RoleOwner {
				ownerCount++
			}
		}

		if ownerCount <= 1 {
			return nil, errors.New("cannot remove the last owner from the organization")
		}
	}

	// Update the role
	_, err = s.orgRepo.UpdateUserRoleInOrganization(ctx, memberID, orgID, req.Role)
	if err != nil {
		return nil, err
	}

	// Get updated member info
	user, err := s.userRepo.GetUserByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	member := &domain.OrganizationMember{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  req.Role,
	}

	return member, nil
}

func (s *organizationService) RemoveMember(ctx context.Context, userID, orgID, memberID uuid.UUID) error {
	// Check if user has permission to remove members
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if !domain.CanManageMembers(role) {
		return errors.New("insufficient permissions to remove members")
	}

	// Check if target member exists in organization
	memberRole, err := s.orgRepo.GetUserRoleInOrganization(ctx, memberID, orgID)
	if err != nil {
		return err
	}
	if memberRole == "" {
		return errors.New("member not found in organization")
	}

	// Prevent non-owners from removing owners
	if memberRole == domain.RoleOwner && role != domain.RoleOwner {
		return errors.New("only owners can remove other owners")
	}

	// Prevent removing the last owner
	if memberRole == domain.RoleOwner {
		members, err := s.orgRepo.GetOrganizationMembers(ctx, orgID)
		if err != nil {
			return err
		}

		ownerCount := 0
		for _, member := range members {
			if member.Role == domain.RoleOwner {
				ownerCount++
			}
		}

		if ownerCount <= 1 {
			return errors.New("cannot remove the last owner from the organization")
		}
	}

	// Remove user from organization
	err = s.orgRepo.RemoveUserFromOrganization(ctx, memberID, orgID)
	if err != nil {
		return err
	}

	return nil
}

func (s *organizationService) DeleteOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	// Check if user has permission to delete organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return err
	}
	if !domain.CanDeleteOrganization(role) {
		return errors.New("insufficient permissions to delete organization")
	}

	// Delete the organization (cascade will handle user_organizations)
	err = s.orgRepo.DeleteOrganization(ctx, orgID)
	if err != nil {
		return err
	}

	return nil
}
