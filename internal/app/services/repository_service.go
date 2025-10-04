package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type RepositoryService interface {
	CreateRepository(ctx context.Context, userID, orgID uuid.UUID, req *domain.CreateRepositoryRequest) (*domain.RepositoryResponse, error)
	GetRepositoriesByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.RepositorySummary, error)
	GetRepository(ctx context.Context, userID, repoID uuid.UUID) (*domain.RepositoryResponse, error)
	DeleteRepository(ctx context.Context, userID, repoID uuid.UUID) error
	ProcessWebhook(ctx context.Context, provider string, payload []byte, signature string) error
}

type repositoryService struct {
	repoRepo repo.RepositoryRepository
	orgRepo  repo.OrganizationRepository
	crypto   *crypto.Crypto
}

func NewRepositoryService(repoRepo repo.RepositoryRepository, orgRepo repo.OrganizationRepository, crypto *crypto.Crypto) RepositoryService {
	return &repositoryService{
		repoRepo: repoRepo,
		orgRepo:  orgRepo,
		crypto:   crypto,
	}
}

func (s *repositoryService) CreateRepository(ctx context.Context, userID, orgID uuid.UUID, req *domain.CreateRepositoryRequest) (*domain.RepositoryResponse, error) {
	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Validate repository type
	if !domain.IsValidRepositoryType(req.Type) {
		return nil, errors.New("invalid repository type")
	}

	// Check if repository already exists for this organization
	existingRepo, err := s.repoRepo.GetRepositoryByURL(ctx, orgID, req.URL)
	if err != nil {
		return nil, err
	}
	if existingRepo != nil {
		return nil, errors.New("repository already exists for this organization")
	}

	// Create repository config
	config := &domain.RepositoryConfig{}
	if req.Token != "" {
		// Encrypt the token
		encryptedToken, err := s.crypto.EncryptString(req.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt token: %w", err)
		}
		config.Token = encryptedToken
	}

	// Convert config to JSON
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	repository := &domain.Repository{
		OrgID:         orgID,
		Type:          req.Type,
		URL:           req.URL,
		DefaultBranch: req.DefaultBranch,
		Config:        configBytes,
	}

	createdRepo, err := s.repoRepo.CreateRepository(ctx, repository)
	if err != nil {
		return nil, err
	}

	response := createdRepo.ToResponse()
	return &response, nil
}

func (s *repositoryService) GetRepositoriesByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.RepositorySummary, error) {
	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	repos, err := s.repoRepo.GetRepositoriesByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (s *repositoryService) GetRepository(ctx context.Context, userID, repoID uuid.UUID) (*domain.RepositoryResponse, error) {
	// Get repository
	repo, err := s.repoRepo.GetRepositoryByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, errors.New("repository not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, repo.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	response := repo.ToResponse()
	return &response, nil
}

func (s *repositoryService) DeleteRepository(ctx context.Context, userID, repoID uuid.UUID) error {
	// Get repository
	repo, err := s.repoRepo.GetRepositoryByID(ctx, repoID)
	if err != nil {
		return err
	}
	if repo == nil {
		return errors.New("repository not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, repo.OrgID)
	if err != nil {
		return err
	}
	if role == "" {
		return errors.New("user does not have access to this organization")
	}

	// Only allow owners and admins to delete repositories
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return errors.New("insufficient permissions to delete repository")
	}

	err = s.repoRepo.DeleteRepository(ctx, repoID)
	if err != nil {
		return err
	}

	return nil
}

func (s *repositoryService) ProcessWebhook(ctx context.Context, provider string, payload []byte, signature string) error {
	// Parse the webhook payload to extract repository information
	var repoURL string
	var branch string

	switch strings.ToLower(provider) {
	case "github":
		var githubPayload domain.WebhookPayload
		if err := json.Unmarshal(payload, &githubPayload); err != nil {
			return fmt.Errorf("failed to parse GitHub webhook payload: %w", err)
		}
		repoURL = githubPayload.Repository.CloneURL
		branch = strings.TrimPrefix(githubPayload.Ref, "refs/heads/")
	case "gitlab":
		var gitlabPayload domain.GitLabWebhookPayload
		if err := json.Unmarshal(payload, &gitlabPayload); err != nil {
			return fmt.Errorf("failed to parse GitLab webhook payload: %w", err)
		}
		repoURL = gitlabPayload.Repository.URL
		branch = strings.TrimPrefix(gitlabPayload.Ref, "refs/heads/")
	case "gitea":
		var giteaPayload domain.GiteaWebhookPayload
		if err := json.Unmarshal(payload, &giteaPayload); err != nil {
			return fmt.Errorf("failed to parse Gitea webhook payload: %w", err)
		}
		repoURL = giteaPayload.Repository.CloneURL
		branch = strings.TrimPrefix(giteaPayload.Ref, "refs/heads/")
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Find the repository by URL
	// Note: This is a simplified approach. In a real implementation, you might need
	// to search across all organizations or maintain a mapping
	// For now, we'll assume we can find it by searching through organizations

	// TODO: Implement repository lookup by URL
	// This would require either:
	// 1. Adding a URL index to the repositories table
	// 2. Searching through organizations
	// 3. Maintaining a separate mapping table

	fmt.Printf("Webhook received for %s repository: %s, branch: %s\n", provider, repoURL, branch)

	// TODO: Create pipeline record when applications are implemented
	// For now, just log the webhook event

	return nil
}
