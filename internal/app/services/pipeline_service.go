package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// PipelineService defines the interface for pipeline operations
type PipelineService interface {
	CreatePipeline(ctx context.Context, userID, appID uuid.UUID, req domain.CreatePipelineRequest) (*domain.PipelineResponse, error)
	GetPipelinesByApp(ctx context.Context, userID, appID uuid.UUID, limit, offset int) ([]domain.PipelineSummary, error)
	GetPipelineByID(ctx context.Context, userID, pipelineID uuid.UUID) (*domain.PipelineResponse, error)
	GetPipelineLogs(ctx context.Context, userID, pipelineID uuid.UUID) (*domain.PipelineLogsResponse, error)
	TriggerPipeline(ctx context.Context, userID, appID uuid.UUID, req domain.CreatePipelineRequest) (*domain.PipelineResponse, error)
}

// pipelineService implements PipelineService
type pipelineService struct {
	pipelineRepo     repo.PipelineRepository
	pipelineStepRepo repo.PipelineStepRepository
	appRepo          repo.ApplicationRepository
	repoRepo         repo.RepositoryRepository
	orgRepo          repo.OrganizationRepository
	jobRepo          repo.JobRepository
	logger           *zap.Logger
}

// NewPipelineService creates a new pipeline service
func NewPipelineService(
	pipelineRepo repo.PipelineRepository,
	pipelineStepRepo repo.PipelineStepRepository,
	appRepo repo.ApplicationRepository,
	repoRepo repo.RepositoryRepository,
	orgRepo repo.OrganizationRepository,
	jobRepo repo.JobRepository,
	logger *zap.Logger,
) PipelineService {
	return &pipelineService{
		pipelineRepo:     pipelineRepo,
		pipelineStepRepo: pipelineStepRepo,
		appRepo:          appRepo,
		repoRepo:         repoRepo,
		orgRepo:          orgRepo,
		jobRepo:          jobRepo,
		logger:           logger,
	}
}

// CreatePipeline creates a new pipeline
func (s *pipelineService) CreatePipeline(ctx context.Context, userID, appID uuid.UUID, req domain.CreatePipelineRequest) (*domain.PipelineResponse, error) {
	// Get application details
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", appID.String()))
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Get repository details
	repo, err := s.repoRepo.GetRepositoryByID(ctx, app.RepoID)
	if err != nil {
		s.logger.Error("Failed to get repository", zap.Error(err), zap.String("repoID", app.RepoID.String()))
		return nil, fmt.Errorf("repository not found: %w", err)
	}
	if repo == nil {
		return nil, errors.New("repository not found")
	}

	// Create pipeline
	pipeline := &domain.Pipeline{
		AppID:       appID,
		RepoID:      app.RepoID,
		CommitSHA:   req.CommitSHA,
		Status:      domain.PipelineStatusPending,
		TriggeredBy: userID,
		Meta: map[string]interface{}{
			"ref":          req.Ref,
			"triggered_by": "manual",
			"repository":   repo.URL,
			"app_name":     app.Name,
		},
	}

	createdPipeline, err := s.pipelineRepo.CreatePipeline(ctx, pipeline)
	if err != nil {
		s.logger.Error("Failed to create pipeline", zap.Error(err))
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Enqueue pipeline job
	jobPayload := domain.JobPayload{
		PipelineID: &createdPipeline.ID,
		Config: map[string]interface{}{
			"app_id":       appID.String(),
			"repo_id":      app.RepoID.String(),
			"commit_sha":   req.CommitSHA,
			"ref":          req.Ref,
			"triggered_by": userID.String(),
		},
	}

	job := &domain.Job{
		Type:    domain.JobTypePipelineRun,
		Status:  domain.JobStatusPending,
		Payload: jobPayload,
	}

	_, err = s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		s.logger.Error("Failed to create pipeline job", zap.Error(err))
		// Don't fail the pipeline creation if job creation fails
		s.logger.Warn("Pipeline created but job queue failed", zap.String("pipelineID", createdPipeline.ID.String()))
	}

	// Get pipeline steps (empty for now)
	steps, err := s.pipelineStepRepo.GetPipelineStepsByPipelineID(ctx, createdPipeline.ID)
	if err != nil {
		s.logger.Error("Failed to get pipeline steps", zap.Error(err))
		steps = []domain.PipelineStep{} // Return empty steps if error
	}

	response := createdPipeline.ToResponse(steps)
	return &response, nil
}

// GetPipelinesByApp retrieves pipelines for an application
func (s *pipelineService) GetPipelinesByApp(ctx context.Context, userID, appID uuid.UUID, limit, offset int) ([]domain.PipelineSummary, error) {
	// Get application details
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", appID.String()))
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get pipelines
	pipelines, err := s.pipelineRepo.GetPipelinesByAppID(ctx, appID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get pipelines", zap.Error(err), zap.String("appID", appID.String()))
		return nil, fmt.Errorf("failed to get pipelines: %w", err)
	}

	return pipelines, nil
}

// GetPipelineByID retrieves a pipeline by ID
func (s *pipelineService) GetPipelineByID(ctx context.Context, userID, pipelineID uuid.UUID) (*domain.PipelineResponse, error) {
	// Get pipeline
	pipeline, err := s.pipelineRepo.GetPipelineByID(ctx, pipelineID)
	if err != nil {
		s.logger.Error("Failed to get pipeline", zap.Error(err), zap.String("pipelineID", pipelineID.String()))
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}
	if pipeline == nil {
		return nil, errors.New("pipeline not found")
	}

	// Get application details for authorization
	app, err := s.appRepo.GetApplicationByID(ctx, pipeline.AppID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", pipeline.AppID.String()))
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Get pipeline steps
	steps, err := s.pipelineStepRepo.GetPipelineStepsByPipelineID(ctx, pipelineID)
	if err != nil {
		s.logger.Error("Failed to get pipeline steps", zap.Error(err), zap.String("pipelineID", pipelineID.String()))
		return nil, fmt.Errorf("failed to get pipeline steps: %w", err)
	}

	response := pipeline.ToResponse(steps)
	return &response, nil
}

// GetPipelineLogs retrieves logs for a pipeline
func (s *pipelineService) GetPipelineLogs(ctx context.Context, userID, pipelineID uuid.UUID) (*domain.PipelineLogsResponse, error) {
	// Get pipeline
	pipeline, err := s.pipelineRepo.GetPipelineByID(ctx, pipelineID)
	if err != nil {
		s.logger.Error("Failed to get pipeline", zap.Error(err), zap.String("pipelineID", pipelineID.String()))
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}
	if pipeline == nil {
		return nil, errors.New("pipeline not found")
	}

	// Get application details for authorization
	app, err := s.appRepo.GetApplicationByID(ctx, pipeline.AppID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", pipeline.AppID.String()))
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Get pipeline steps with logs
	steps, err := s.pipelineStepRepo.GetPipelineStepsByPipelineID(ctx, pipelineID)
	if err != nil {
		s.logger.Error("Failed to get pipeline steps", zap.Error(err), zap.String("pipelineID", pipelineID.String()))
		return nil, fmt.Errorf("failed to get pipeline steps: %w", err)
	}

	// Combine all step logs
	var allLogs string
	for _, step := range steps {
		if step.Logs != "" {
			allLogs += fmt.Sprintf("=== %s ===\n%s\n\n", step.Name, step.Logs)
		}
	}

	return &domain.PipelineLogsResponse{
		PipelineID: pipelineID,
		Logs:       allLogs,
		Steps:      steps,
	}, nil
}

// TriggerPipeline is an alias for CreatePipeline for consistency with API naming
func (s *pipelineService) TriggerPipeline(ctx context.Context, userID, appID uuid.UUID, req domain.CreatePipelineRequest) (*domain.PipelineResponse, error) {
	return s.CreatePipeline(ctx, userID, appID, req)
}
