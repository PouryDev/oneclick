package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
)

// MockEventRepository is a mock implementation of EventRepository
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) CreateEventLog(ctx context.Context, event *domain.EventLog) (*domain.EventLog, error) {
	args := m.Called(ctx, event)
	return args.Get(0).(*domain.EventLog), args.Error(1)
}

func (m *MockEventRepository) GetEventLogsByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.EventLog, error) {
	args := m.Called(ctx, orgID, limit, offset)
	return args.Get(0).([]*domain.EventLog), args.Error(1)
}

func (m *MockEventRepository) GetEventLogsByOrgIDAndAction(ctx context.Context, orgID uuid.UUID, action domain.EventAction, limit, offset int) ([]*domain.EventLog, error) {
	args := m.Called(ctx, orgID, action, limit, offset)
	return args.Get(0).([]*domain.EventLog), args.Error(1)
}

func (m *MockEventRepository) GetEventLogsByOrgIDAndResourceType(ctx context.Context, orgID uuid.UUID, resourceType domain.ResourceType, limit, offset int) ([]*domain.EventLog, error) {
	args := m.Called(ctx, orgID, resourceType, limit, offset)
	return args.Get(0).([]*domain.EventLog), args.Error(1)
}

func (m *MockEventRepository) GetEventLogByID(ctx context.Context, id uuid.UUID) (*domain.EventLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.EventLog), args.Error(1)
}

func (m *MockEventRepository) DeleteEventLog(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockDashboardCountsRepository is a mock implementation of DashboardCountsRepository
type MockDashboardCountsRepository struct {
	mock.Mock
}

func (m *MockDashboardCountsRepository) GetDashboardCounts(ctx context.Context, orgID uuid.UUID) (*domain.DashboardCounts, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).(*domain.DashboardCounts), args.Error(1)
}

func (m *MockDashboardCountsRepository) UpdateDashboardCounts(ctx context.Context, counts *domain.DashboardCounts) (*domain.DashboardCounts, error) {
	args := m.Called(ctx, counts)
	return args.Get(0).(*domain.DashboardCounts), args.Error(1)
}

func (m *MockDashboardCountsRepository) DeleteDashboardCounts(ctx context.Context, orgID uuid.UUID) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

// MockReadModelProjectRepository is a mock implementation of ReadModelProjectRepository
type MockReadModelProjectRepository struct {
	mock.Mock
}

func (m *MockReadModelProjectRepository) CreateReadModelProject(ctx context.Context, project *domain.ReadModelProject) (*domain.ReadModelProject, error) {
	args := m.Called(ctx, project)
	return args.Get(0).(*domain.ReadModelProject), args.Error(1)
}

func (m *MockReadModelProjectRepository) GetReadModelProject(ctx context.Context, orgID uuid.UUID, key string) (*domain.ReadModelProject, error) {
	args := m.Called(ctx, orgID, key)
	return args.Get(0).(*domain.ReadModelProject), args.Error(1)
}

func (m *MockReadModelProjectRepository) GetReadModelProjectsByOrgID(ctx context.Context, orgID uuid.UUID) ([]*domain.ReadModelProject, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*domain.ReadModelProject), args.Error(1)
}

func (m *MockReadModelProjectRepository) DeleteReadModelProject(ctx context.Context, orgID uuid.UUID, key string) error {
	args := m.Called(ctx, orgID, key)
	return args.Error(0)
}

func TestEventLoggerService_LogEvent(t *testing.T) {
	// Setup mocks
	eventRepo := &MockEventRepository{}
	orgRepo := &MockOrganizationRepository{}
	logger := zap.NewNop()

	service := NewEventLoggerService(eventRepo, orgRepo, logger)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	appID := uuid.New()

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock event creation
	createdEvent := &domain.EventLog{
		ID:           uuid.New(),
		OrgID:        orgID,
		UserID:       userID,
		Action:       domain.EventActionAppCreated,
		ResourceType: domain.ResourceTypeApp,
		ResourceID:   appID,
		Details:      map[string]interface{}{"name": "test-app"},
		CreatedAt:    time.Now(),
	}
	eventRepo.On("CreateEventLog", ctx, mock.AnythingOfType("*domain.EventLog")).Return(createdEvent, nil)

	// Test
	req := domain.CreateEventRequest{
		OrgID:        orgID,
		UserID:       userID,
		Action:       domain.EventActionAppCreated,
		ResourceType: domain.ResourceTypeApp,
		ResourceID:   appID,
		Details:      map[string]interface{}{"name": "test-app"},
	}

	result, err := service.LogEvent(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, createdEvent.ID, result.ID)
	assert.Equal(t, domain.EventActionAppCreated, result.Action)
	assert.Equal(t, domain.ResourceTypeApp, result.ResourceType)

	// Verify mocks
	eventRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestEventLoggerService_GetEventsByOrgID(t *testing.T) {
	// Setup mocks
	eventRepo := &MockEventRepository{}
	orgRepo := &MockOrganizationRepository{}
	logger := zap.NewNop()

	service := NewEventLoggerService(eventRepo, orgRepo, logger)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock events
	events := []*domain.EventLog{
		{
			ID:           uuid.New(),
			OrgID:        orgID,
			UserID:       userID,
			Action:       domain.EventActionAppCreated,
			ResourceType: domain.ResourceTypeApp,
			ResourceID:   uuid.New(),
			Details:      map[string]interface{}{"name": "test-app"},
			CreatedAt:    time.Now(),
		},
	}
	eventRepo.On("GetEventLogsByOrgID", ctx, orgID, 50, 0).Return(events, nil)

	// Test
	result, err := service.GetEventsByOrgID(ctx, userID, orgID, 50, 0)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, events[0].ID, result[0].ID)

	// Verify mocks
	eventRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestEventLoggerService_GetEventByID(t *testing.T) {
	// Setup mocks
	eventRepo := &MockEventRepository{}
	orgRepo := &MockOrganizationRepository{}
	logger := zap.NewNop()

	service := NewEventLoggerService(eventRepo, orgRepo, logger)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	eventID := uuid.New()

	// Mock event retrieval
	event := &domain.EventLog{
		ID:           eventID,
		OrgID:        orgID,
		UserID:       userID,
		Action:       domain.EventActionAppCreated,
		ResourceType: domain.ResourceTypeApp,
		ResourceID:   uuid.New(),
		Details:      map[string]interface{}{"name": "test-app"},
		CreatedAt:    time.Now(),
	}
	eventRepo.On("GetEventLogByID", ctx, eventID).Return(event, nil)

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Test
	result, err := service.GetEventByID(ctx, userID, eventID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, eventID, result.ID)
	assert.Equal(t, orgID, result.OrgID)

	// Verify mocks
	eventRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestDashboardService_GetDashboardCounts(t *testing.T) {
	// Setup mocks
	dashboardCountsRepo := &MockDashboardCountsRepository{}
	appRepo := &MockApplicationRepository{}
	clusterRepo := &MockClusterRepository{}
	pipelineRepo := &MockPipelineRepository{}
	orgRepo := &MockOrganizationRepository{}
	logger := zap.NewNop()

	service := NewDashboardService(dashboardCountsRepo, appRepo, clusterRepo, pipelineRepo, orgRepo, logger)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock dashboard counts
	counts := &domain.DashboardCounts{
		OrgID:            orgID,
		AppsCount:        5,
		ClustersCount:    2,
		RunningPipelines: 1,
		UpdatedAt:        time.Now(),
	}
	dashboardCountsRepo.On("GetDashboardCounts", ctx, orgID).Return(counts, nil)

	// Test
	result, err := service.GetDashboardCounts(ctx, userID, orgID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orgID, result.OrgID)
	assert.Equal(t, 5, result.AppsCount)
	assert.Equal(t, 2, result.ClustersCount)
	assert.Equal(t, 1, result.RunningPipelines)

	// Verify mocks
	dashboardCountsRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestReadModelService_CreateReadModelProject(t *testing.T) {
	// Setup mocks
	readModelRepo := &MockReadModelProjectRepository{}
	orgRepo := &MockOrganizationRepository{}
	logger := zap.NewNop()

	service := NewReadModelService(readModelRepo, orgRepo, logger)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	key := "test_key"
	value := map[string]interface{}{"data": "test"}

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock project creation
	createdProject := &domain.ReadModelProject{
		ID:    uuid.New(),
		OrgID: orgID,
		Key:   key,
		Value: value,
	}
	readModelRepo.On("CreateReadModelProject", ctx, mock.AnythingOfType("*domain.ReadModelProject")).Return(createdProject, nil)

	// Test
	result, err := service.CreateReadModelProject(ctx, userID, orgID, key, value)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, createdProject.ID, result.ID)
	assert.Equal(t, orgID, result.OrgID)
	assert.Equal(t, key, result.Key)
	assert.Equal(t, value, result.Value)

	// Verify mocks
	readModelRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}

func TestReadModelService_GetReadModelProject(t *testing.T) {
	// Setup mocks
	readModelRepo := &MockReadModelProjectRepository{}
	orgRepo := &MockOrganizationRepository{}
	logger := zap.NewNop()

	service := NewReadModelService(readModelRepo, orgRepo, logger)

	ctx := context.Background()
	userID := uuid.New()
	orgID := uuid.New()
	key := "test_key"

	// Mock user role check
	orgRepo.On("GetUserRoleInOrganization", ctx, userID, orgID).Return("admin", nil)

	// Mock project retrieval
	project := &domain.ReadModelProject{
		ID:    uuid.New(),
		OrgID: orgID,
		Key:   key,
		Value: map[string]interface{}{"data": "test"},
	}
	readModelRepo.On("GetReadModelProject", ctx, orgID, key).Return(project, nil)

	// Test
	result, err := service.GetReadModelProject(ctx, userID, orgID, key)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project.ID, result.ID)
	assert.Equal(t, orgID, result.OrgID)
	assert.Equal(t, key, result.Key)

	// Verify mocks
	readModelRepo.AssertExpectations(t)
	orgRepo.AssertExpectations(t)
}
