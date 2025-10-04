package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/app/deployment"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type ApplicationService interface {
	CreateApplication(ctx context.Context, userID, clusterID uuid.UUID, req *domain.CreateApplicationRequest) (*domain.ApplicationResponse, error)
	GetApplicationsByCluster(ctx context.Context, userID, clusterID uuid.UUID) ([]domain.ApplicationSummary, error)
	GetApplication(ctx context.Context, userID, appID uuid.UUID) (*domain.ApplicationDetail, error)
	DeleteApplication(ctx context.Context, userID, appID uuid.UUID) error
	DeployApplication(ctx context.Context, userID, appID uuid.UUID, req *domain.DeployApplicationRequest) (*domain.DeployApplicationResponse, error)
	RollbackApplication(ctx context.Context, userID, appID, releaseID uuid.UUID) (*domain.DeployApplicationResponse, error)
	GetReleasesByApplication(ctx context.Context, userID, appID uuid.UUID) ([]domain.ReleaseSummary, error)
}

type applicationService struct {
	appRepo     repo.ApplicationRepository
	releaseRepo repo.ReleaseRepository
	clusterRepo repo.ClusterRepository
	repoRepo    repo.RepositoryRepository
	orgRepo     repo.OrganizationRepository
	deployer    *deployment.DeploymentGenerator
}

func NewApplicationService(
	appRepo repo.ApplicationRepository,
	releaseRepo repo.ReleaseRepository,
	clusterRepo repo.ClusterRepository,
	repoRepo repo.RepositoryRepository,
	orgRepo repo.OrganizationRepository,
) ApplicationService {
	return &applicationService{
		appRepo:     appRepo,
		releaseRepo: releaseRepo,
		clusterRepo: clusterRepo,
		repoRepo:    repoRepo,
		orgRepo:     orgRepo,
		deployer:    deployment.NewDeploymentGenerator(),
	}
}

func (s *applicationService) CreateApplication(ctx context.Context, userID, clusterID uuid.UUID, req *domain.CreateApplicationRequest) (*domain.ApplicationResponse, error) {
	// Get cluster to verify it exists and user has access
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Parse repository ID
	repoID, err := uuid.Parse(req.RepoID)
	if err != nil {
		return nil, errors.New("invalid repository ID")
	}

	// Verify repository exists and belongs to the same organization
	repository, err := s.repoRepo.GetRepositoryByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	if repository == nil {
		return nil, errors.New("repository not found")
	}
	if repository.OrgID != cluster.OrgID {
		return nil, errors.New("repository does not belong to the same organization")
	}

	// Check if application name already exists in cluster
	existingApp, err := s.appRepo.GetApplicationByNameInCluster(ctx, clusterID, req.Name)
	if err != nil {
		return nil, err
	}
	if existingApp != nil {
		return nil, errors.New("application name already exists in this cluster")
	}

	// Create application
	application := &domain.Application{
		OrgID:         cluster.OrgID,
		ClusterID:     clusterID,
		Name:          req.Name,
		RepoID:        repoID,
		DefaultBranch: req.DefaultBranch,
	}

	if req.Path != "" {
		application.Path = &req.Path
	}

	createdApp, err := s.appRepo.CreateApplication(ctx, application)
	if err != nil {
		return nil, err
	}

	response := createdApp.ToResponse()
	return &response, nil
}

func (s *applicationService) GetApplicationsByCluster(ctx context.Context, userID, clusterID uuid.UUID) ([]domain.ApplicationSummary, error) {
	// Get cluster to verify it exists and user has access
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	apps, err := s.appRepo.GetApplicationsByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (s *applicationService) GetApplication(ctx context.Context, userID, appID uuid.UUID) (*domain.ApplicationDetail, error) {
	// Get application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Get latest release
	latestRelease, err := s.releaseRepo.GetLatestReleaseByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// Get all releases for count
	releases, err := s.releaseRepo.GetReleasesByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	detail := &domain.ApplicationDetail{
		Application:  *app,
		ReleaseCount: len(releases),
		Status:       "unknown",
	}

	if latestRelease != nil {
		releaseSummary := latestRelease.ToSummary()
		detail.CurrentRelease = &releaseSummary
		detail.Status = string(latestRelease.Status)
	}

	return detail, nil
}

func (s *applicationService) DeleteApplication(ctx context.Context, userID, appID uuid.UUID) error {
	// Get application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return errors.New("application not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		return err
	}
	if role == "" {
		return errors.New("user does not have access to this organization")
	}

	// Only allow owners and admins to delete applications
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return errors.New("insufficient permissions to delete application")
	}

	err = s.appRepo.DeleteApplication(ctx, appID)
	if err != nil {
		return err
	}

	return nil
}

func (s *applicationService) DeployApplication(ctx context.Context, userID, appID uuid.UUID, req *domain.DeployApplicationRequest) (*domain.DeployApplicationResponse, error) {
	// Get application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Validate image and tag
	if req.Image == "" {
		return nil, errors.New("image is required")
	}
	if req.Tag == "" {
		return nil, errors.New("tag is required")
	}

	// Create release record
	release := &domain.Release{
		AppID:     appID,
		Image:     req.Image,
		Tag:       req.Tag,
		CreatedBy: userID,
		Status:    domain.ReleaseStatusPending,
		Meta:      json.RawMessage("{}"),
	}

	createdRelease, err := s.releaseRepo.CreateRelease(ctx, release)
	if err != nil {
		return nil, err
	}

	// TODO: In a real implementation, this would publish to a message queue
	// For now, we'll simulate the deployment process
	go s.processDeployment(context.Background(), createdRelease)

	response := &domain.DeployApplicationResponse{
		ReleaseID: createdRelease.ID,
		Status:    string(createdRelease.Status),
		Message:   "Deployment initiated",
	}

	return response, nil
}

func (s *applicationService) RollbackApplication(ctx context.Context, userID, appID, releaseID uuid.UUID) (*domain.DeployApplicationResponse, error) {
	// Get application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Get the release to rollback to
	rollbackRelease, err := s.releaseRepo.GetReleaseByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if rollbackRelease == nil {
		return nil, errors.New("release not found")
	}
	if rollbackRelease.AppID != appID {
		return nil, errors.New("release does not belong to this application")
	}

	// Create new release with the same image/tag
	newRelease := &domain.Release{
		AppID:     appID,
		Image:     rollbackRelease.Image,
		Tag:       rollbackRelease.Tag,
		CreatedBy: userID,
		Status:    domain.ReleaseStatusPending,
		Meta:      rollbackRelease.Meta,
	}

	createdRelease, err := s.releaseRepo.CreateRelease(ctx, newRelease)
	if err != nil {
		return nil, err
	}

	// TODO: In a real implementation, this would publish to a message queue
	// For now, we'll simulate the deployment process
	go s.processDeployment(context.Background(), createdRelease)

	response := &domain.DeployApplicationResponse{
		ReleaseID: createdRelease.ID,
		Status:    string(createdRelease.Status),
		Message:   "Rollback initiated",
	}

	return response, nil
}

func (s *applicationService) GetReleasesByApplication(ctx context.Context, userID, appID uuid.UUID) ([]domain.ReleaseSummary, error) {
	// Get application
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	releases, err := s.releaseRepo.GetReleasesByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

// processDeployment simulates the deployment process
func (s *applicationService) processDeployment(ctx context.Context, release *domain.Release) {
	// Update status to running
	now := time.Now()
	_, err := s.releaseRepo.UpdateReleaseStatus(ctx, release.ID, domain.ReleaseStatusRunning, &now, nil)
	if err != nil {
		fmt.Printf("Failed to update release status to running: %v\n", err)
		return
	}

	// Simulate deployment time
	time.Sleep(5 * time.Second)

	// Simulate success (in real implementation, this would check actual deployment status)
	finishedAt := time.Now()
	_, err = s.releaseRepo.UpdateReleaseStatus(ctx, release.ID, domain.ReleaseStatusSucceeded, nil, &finishedAt)
	if err != nil {
		fmt.Printf("Failed to update release status to succeeded: %v\n", err)
		return
	}

	fmt.Printf("Deployment completed for release %s\n", release.ID)
}
