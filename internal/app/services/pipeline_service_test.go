package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/PouryDev/oneclick/internal/domain"
)

// MockPipelineRepository is a mock implementation of PipelineRepository
type MockPipelineRepository struct {
	mock.Mock
}

func (m *MockPipelineRepository) CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) (*domain.Pipeline, error) {
	args := m.Called(ctx, pipeline)
	return args.Get(0).(*domain.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) GetPipelineByID(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) GetPipelinesByAppID(ctx context.Context, appID uuid.UUID, limit, offset int) ([]domain.PipelineSummary, error) {
	args := m.Called(ctx, appID, limit, offset)
	return args.Get(0).([]domain.PipelineSummary), args.Error(1)
}

func (m *MockPipelineRepository) UpdatePipelineStatus(ctx context.Context, id uuid.UUID, status domain.PipelineStatus) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(*domain.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) UpdatePipelineStarted(ctx context.Context, id uuid.UUID, status domain.PipelineStatus, startedAt *time.Time) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, status, startedAt)
	return args.Get(0).(*domain.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) UpdatePipelineFinished(ctx context.Context, id uuid.UUID, status domain.PipelineStatus, finishedAt *time.Time) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, status, finishedAt)
	return args.Get(0).(*domain.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) UpdatePipelineLogsURL(ctx context.Context, id uuid.UUID, logsURL string) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, logsURL)
	return args.Get(0).(*domain.Pipeline), args.Error(1)
}

func (m *MockPipelineRepository) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPipelineStepRepository is a mock implementation of PipelineStepRepository
type MockPipelineStepRepository struct {
	mock.Mock
}

func (m *MockPipelineStepRepository) CreatePipelineStep(ctx context.Context, step *domain.PipelineStep) (*domain.PipelineStep, error) {
	args := m.Called(ctx, step)
	return args.Get(0).(*domain.PipelineStep), args.Error(1)
}

func (m *MockPipelineStepRepository) GetPipelineStepsByPipelineID(ctx context.Context, pipelineID uuid.UUID) ([]domain.PipelineStep, error) {
	args := m.Called(ctx, pipelineID)
	return args.Get(0).([]domain.PipelineStep), args.Error(1)
}

func (m *MockPipelineStepRepository) UpdatePipelineStepStatus(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus) (*domain.PipelineStep, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(*domain.PipelineStep), args.Error(1)
}

func (m *MockPipelineStepRepository) UpdatePipelineStepStarted(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus, startedAt *time.Time) (*domain.PipelineStep, error) {
	args := m.Called(ctx, id, status, startedAt)
	return args.Get(0).(*domain.PipelineStep), args.Error(1)
}

func (m *MockPipelineStepRepository) UpdatePipelineStepFinished(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus, finishedAt *time.Time) (*domain.PipelineStep, error) {
	args := m.Called(ctx, id, status, finishedAt)
	return args.Get(0).(*domain.PipelineStep), args.Error(1)
}

func (m *MockPipelineStepRepository) UpdatePipelineStepLogs(ctx context.Context, id uuid.UUID, logs string) (*domain.PipelineStep, error) {
	args := m.Called(ctx, id, logs)
	return args.Get(0).(*domain.PipelineStep), args.Error(1)
}

func (m *MockPipelineStepRepository) DeletePipelineStep(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockRepositoryRepository is a mock implementation of RepositoryRepository
type MockRepositoryRepository struct {
	mock.Mock
}

func (m *MockRepositoryRepository) GetRepositoryByID(ctx context.Context, id uuid.UUID) (*domain.Repository, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryRepository) GetRepositoryByURL(ctx context.Context, orgID uuid.UUID, url string) (*domain.Repository, error) {
	args := m.Called(ctx, orgID, url)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryRepository) GetRepositoriesByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.RepositorySummary, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.RepositorySummary), args.Error(1)
}

func (m *MockRepositoryRepository) CreateRepository(ctx context.Context, repo *domain.Repository) (*domain.Repository, error) {
	args := m.Called(ctx, repo)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryRepository) UpdateRepositoryConfig(ctx context.Context, id uuid.UUID, config []byte) (*domain.Repository, error) {
	args := m.Called(ctx, id, config)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryRepository) DeleteRepository(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockJobRepository is a mock implementation of JobRepository
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) (*domain.Job, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) StartJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetPendingJobs(ctx context.Context) ([]domain.Job, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetJobsByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.Job, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetJobByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	args := m.Called(ctx, job)
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) CompleteJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobRepository) FailJob(ctx context.Context, id uuid.UUID, reason string) (*domain.Job, error) {
	args := m.Called(ctx, id, reason)
	return args.Get(0).(*domain.Job), args.Error(1)
}

func TestPipelineService_TriggerPipeline(t *testing.T) {
	// Setup mocks
	pipelineRepo := &MockPipelineRepository{}
	pipelineStepRepo := &MockPipelineStepRepository{}
	appRepo := &MockApplicationRepository{}
	repositoryRepo := &MockRepositoryRepository{}
	jobRepo := &MockJobRepository{}
	orgRepo := &MockOrganizationRepository{}

	service := NewPipelineService(pipelineRepo, pipelineStepRepo, appRepo, repositoryRepo, orgRepo, jobRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()
	repoID := uuid.New()

	// Mock application exists
	orgID := uuid.New()
	app := &domain.Application{
		ID:     appID,
		RepoID: repoID,
		OrgID:  orgID,
		Name:   "test-app",
	}
	appRepo.On("GetApplicationByID", ctx, appID).Return(app, nil)

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock repository exists
	repository := &domain.Repository{
		ID:  repoID,
		URL: "https://github.com/test/repo",
	}
	repositoryRepo.On("GetRepositoryByID", ctx, repoID).Return(repository, nil)

	// Mock pipeline creation
	pipelineID := uuid.New()
	createdPipeline := &domain.Pipeline{
		ID:          pipelineID,
		AppID:       appID,
		RepoID:      repoID,
		CommitSHA:   "abc123",
		Status:      domain.PipelineStatusPending,
		TriggeredBy: userID,
		Meta:        map[string]interface{}{"ref": "main"},
	}
	pipelineRepo.On("CreatePipeline", ctx, mock.AnythingOfType("*domain.Pipeline")).Return(createdPipeline, nil)

	// Mock pipeline steps (empty for new pipeline)
	pipelineStepRepo.On("GetPipelineStepsByPipelineID", ctx, pipelineID).Return([]domain.PipelineStep{}, nil)

	// Mock job creation
	jobID := uuid.New()
	createdJob := &domain.Job{
		ID:        jobID,
		Type:      domain.JobTypePipelineRun,
		Status:    domain.JobStatusPending,
		CreatedAt: time.Now(),
	}
	jobRepo.On("CreateJob", ctx, mock.AnythingOfType("*domain.Job")).Return(createdJob, nil)

	// Test
	req := domain.CreatePipelineRequest{
		Ref:       "main",
		CommitSHA: "abc123",
	}

	result, err := service.TriggerPipeline(ctx, userID, appID, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, pipelineID, result.ID)
	assert.Equal(t, domain.PipelineStatusPending, result.Status)

	// Verify mocks
	pipelineRepo.AssertExpectations(t)
	pipelineStepRepo.AssertExpectations(t)
	appRepo.AssertExpectations(t)
	repositoryRepo.AssertExpectations(t)
	jobRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestPipelineService_GetPipelinesByApp(t *testing.T) {
	// Setup mocks
	pipelineRepo := &MockPipelineRepository{}
	pipelineStepRepo := &MockPipelineStepRepository{}
	appRepo := &MockApplicationRepository{}
	repositoryRepo := &MockRepositoryRepository{}
	jobRepo := &MockJobRepository{}
	orgRepo := &MockOrganizationRepository{}

	service := NewPipelineService(pipelineRepo, pipelineStepRepo, appRepo, repositoryRepo, orgRepo, jobRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	appID := uuid.New()

	// Mock application exists
	orgID := uuid.New()
	app := &domain.Application{
		ID:    appID,
		OrgID: orgID,
		Name:  "test-app",
	}
	appRepo.On("GetApplicationByID", ctx, appID).Return(app, nil)

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock pipelines
	pipelines := []domain.PipelineSummary{
		{
			ID:          uuid.New(),
			AppID:       appID,
			Status:      domain.PipelineStatusSuccess,
			TriggeredBy: userID,
			CommitSHA:   "abc123",
		},
		{
			ID:          uuid.New(),
			AppID:       appID,
			Status:      domain.PipelineStatusFailed,
			TriggeredBy: userID,
			CommitSHA:   "def456",
		},
	}
	pipelineRepo.On("GetPipelinesByAppID", ctx, appID, 20, 0).Return(pipelines, nil)

	// Test
	result, err := service.GetPipelinesByApp(ctx, userID, appID, 20, 0)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, domain.PipelineStatusSuccess, result[0].Status)
	assert.Equal(t, domain.PipelineStatusFailed, result[1].Status)

	// Verify mocks
	pipelineRepo.AssertExpectations(t)
	appRepo.AssertExpectations(t)
}

func TestPipelineService_GetPipeline(t *testing.T) {
	// Setup mocks
	pipelineRepo := &MockPipelineRepository{}
	pipelineStepRepo := &MockPipelineStepRepository{}
	appRepo := &MockApplicationRepository{}
	repositoryRepo := &MockRepositoryRepository{}
	jobRepo := &MockJobRepository{}
	orgRepo := &MockOrganizationRepository{}

	service := NewPipelineService(pipelineRepo, pipelineStepRepo, appRepo, repositoryRepo, orgRepo, jobRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	pipelineID := uuid.New()
	appID := uuid.New()

	// Mock pipeline exists
	pipeline := &domain.Pipeline{
		ID:          pipelineID,
		AppID:       appID,
		Status:      domain.PipelineStatusRunning,
		TriggeredBy: userID,
		CommitSHA:   "abc123",
	}
	pipelineRepo.On("GetPipelineByID", ctx, pipelineID).Return(pipeline, nil)

	// Mock application exists
	orgID := uuid.New()
	app := &domain.Application{
		ID:    appID,
		OrgID: orgID,
		Name:  "test-app",
	}
	appRepo.On("GetApplicationByID", ctx, appID).Return(app, nil)

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock pipeline steps
	steps := []domain.PipelineStep{
		{
			ID:         uuid.New(),
			PipelineID: pipelineID,
			Name:       "checkout",
			Status:     domain.PipelineStepStatusSuccess,
		},
		{
			ID:         uuid.New(),
			PipelineID: pipelineID,
			Name:       "build",
			Status:     domain.PipelineStepStatusRunning,
		},
	}
	pipelineStepRepo.On("GetPipelineStepsByPipelineID", ctx, pipelineID).Return(steps, nil)

	// Test
	result, err := service.GetPipelineByID(ctx, userID, pipelineID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, pipelineID, result.ID)
	assert.Equal(t, domain.PipelineStatusRunning, result.Status)
	assert.Len(t, result.Steps, 2)

	// Verify mocks
	pipelineRepo.AssertExpectations(t)
	pipelineStepRepo.AssertExpectations(t)
	appRepo.AssertExpectations(t)
}
