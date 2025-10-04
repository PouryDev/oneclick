package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// Error constants
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

// EventLoggerService defines the interface for event logging operations
type EventLoggerService interface {
	LogEvent(ctx context.Context, req domain.CreateEventRequest) (*domain.EventLog, error)
	GetEventsByOrgID(ctx context.Context, userID, orgID uuid.UUID, limit, offset int) ([]*domain.EventLog, error)
	GetEventsByOrgIDAndAction(ctx context.Context, userID, orgID uuid.UUID, action domain.EventAction, limit, offset int) ([]*domain.EventLog, error)
	GetEventsByOrgIDAndResourceType(ctx context.Context, userID, orgID uuid.UUID, resourceType domain.ResourceType, limit, offset int) ([]*domain.EventLog, error)
	GetEventByID(ctx context.Context, userID, eventID uuid.UUID) (*domain.EventLog, error)
}

// DashboardService defines the interface for dashboard operations
type DashboardService interface {
	GetDashboardCounts(ctx context.Context, userID, orgID uuid.UUID) (*domain.DashboardCounts, error)
	UpdateDashboardCounts(ctx context.Context, orgID uuid.UUID) (*domain.DashboardCounts, error)
}

// ReadModelService defines the interface for read model operations
type ReadModelService interface {
	CreateReadModelProject(ctx context.Context, userID, orgID uuid.UUID, key string, value map[string]interface{}) (*domain.ReadModelProject, error)
	GetReadModelProject(ctx context.Context, userID, orgID uuid.UUID, key string) (*domain.ReadModelProject, error)
	GetReadModelProjectsByOrgID(ctx context.Context, userID, orgID uuid.UUID) ([]*domain.ReadModelProject, error)
	DeleteReadModelProject(ctx context.Context, userID, orgID uuid.UUID, key string) error
}

// eventLoggerService implements EventLoggerService
type eventLoggerService struct {
	eventRepo repo.EventRepository
	orgRepo   repo.OrganizationRepository
	logger    *zap.Logger
}

// NewEventLoggerService creates a new event logger service
func NewEventLoggerService(eventRepo repo.EventRepository, orgRepo repo.OrganizationRepository, logger *zap.Logger) EventLoggerService {
	return &eventLoggerService{
		eventRepo: eventRepo,
		orgRepo:   orgRepo,
		logger:    logger,
	}
}

func (s *eventLoggerService) LogEvent(ctx context.Context, req domain.CreateEventRequest) (*domain.EventLog, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, req.UserID, req.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", req.UserID.String()),
			zap.String("orgID", req.OrgID.String()))
		return nil, ErrUnauthorized
	}

	// Create event log
	event := &domain.EventLog{
		ID:           uuid.New(),
		OrgID:        req.OrgID,
		UserID:       req.UserID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Details:      req.Details,
		CreatedAt:    time.Now(),
	}

	createdEvent, err := s.eventRepo.CreateEventLog(ctx, event)
	if err != nil {
		s.logger.Error("Failed to create event log", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Event logged successfully",
		zap.String("eventID", createdEvent.ID.String()),
		zap.String("action", string(createdEvent.Action)),
		zap.String("resourceType", string(createdEvent.ResourceType)),
		zap.String("orgID", createdEvent.OrgID.String()))

	return createdEvent, nil
}

func (s *eventLoggerService) GetEventsByOrgID(ctx context.Context, userID, orgID uuid.UUID, limit, offset int) ([]*domain.EventLog, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	events, err := s.eventRepo.GetEventLogsByOrgID(ctx, orgID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get events by org ID", zap.Error(err))
		return nil, err
	}

	return events, nil
}

func (s *eventLoggerService) GetEventsByOrgIDAndAction(ctx context.Context, userID, orgID uuid.UUID, action domain.EventAction, limit, offset int) ([]*domain.EventLog, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	events, err := s.eventRepo.GetEventLogsByOrgIDAndAction(ctx, orgID, action, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get events by org ID and action", zap.Error(err))
		return nil, err
	}

	return events, nil
}

func (s *eventLoggerService) GetEventsByOrgIDAndResourceType(ctx context.Context, userID, orgID uuid.UUID, resourceType domain.ResourceType, limit, offset int) ([]*domain.EventLog, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	events, err := s.eventRepo.GetEventLogsByOrgIDAndResourceType(ctx, orgID, resourceType, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get events by org ID and resource type", zap.Error(err))
		return nil, err
	}

	return events, nil
}

func (s *eventLoggerService) GetEventByID(ctx context.Context, userID, eventID uuid.UUID) (*domain.EventLog, error) {
	// Get the event first to check organization access
	event, err := s.eventRepo.GetEventLogByID(ctx, eventID)
	if err != nil {
		s.logger.Error("Failed to get event by ID", zap.Error(err))
		return nil, err
	}

	if event == nil {
		return nil, ErrNotFound
	}

	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, event.OrgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", event.OrgID.String()))
		return nil, ErrUnauthorized
	}

	return event, nil
}

// dashboardService implements DashboardService
type dashboardService struct {
	dashboardCountsRepo repo.DashboardCountsRepository
	appRepo             repo.ApplicationRepository
	clusterRepo         repo.ClusterRepository
	pipelineRepo        repo.PipelineRepository
	orgRepo             repo.OrganizationRepository
	logger              *zap.Logger
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	dashboardCountsRepo repo.DashboardCountsRepository,
	appRepo repo.ApplicationRepository,
	clusterRepo repo.ClusterRepository,
	pipelineRepo repo.PipelineRepository,
	orgRepo repo.OrganizationRepository,
	logger *zap.Logger,
) DashboardService {
	return &dashboardService{
		dashboardCountsRepo: dashboardCountsRepo,
		appRepo:             appRepo,
		clusterRepo:         clusterRepo,
		pipelineRepo:        pipelineRepo,
		orgRepo:             orgRepo,
		logger:              logger,
	}
}

func (s *dashboardService) GetDashboardCounts(ctx context.Context, userID, orgID uuid.UUID) (*domain.DashboardCounts, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	counts, err := s.dashboardCountsRepo.GetDashboardCounts(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get dashboard counts", zap.Error(err))
		return nil, err
	}

	// If no counts exist, calculate them
	if counts == nil {
		return s.UpdateDashboardCounts(ctx, orgID)
	}

	return counts, nil
}

func (s *dashboardService) UpdateDashboardCounts(ctx context.Context, orgID uuid.UUID) (*domain.DashboardCounts, error) {
	// Get counts from actual data
	// For apps, we need to get clusters first, then apps from each cluster
	clusters, err := s.clusterRepo.GetClustersByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get clusters by org ID", zap.Error(err))
		return nil, err
	}

	var totalApps int
	for _, cluster := range clusters {
		apps, err := s.appRepo.GetApplicationsByClusterID(ctx, cluster.ID)
		if err != nil {
			s.logger.Error("Failed to get applications by cluster ID", zap.Error(err))
			return nil, err
		}
		totalApps += len(apps)
	}

	// Get running pipelines (this would need to be implemented in pipeline repo)
	// For now, we'll set it to 0
	runningPipelines := 0

	counts := &domain.DashboardCounts{
		OrgID:            orgID,
		AppsCount:        totalApps,
		ClustersCount:    len(clusters),
		RunningPipelines: runningPipelines,
		UpdatedAt:        time.Now(),
	}

	updatedCounts, err := s.dashboardCountsRepo.UpdateDashboardCounts(ctx, counts)
	if err != nil {
		s.logger.Error("Failed to update dashboard counts", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Dashboard counts updated",
		zap.String("orgID", orgID.String()),
		zap.Int("appsCount", updatedCounts.AppsCount),
		zap.Int("clustersCount", updatedCounts.ClustersCount),
		zap.Int("runningPipelines", updatedCounts.RunningPipelines))

	return updatedCounts, nil
}

// readModelService implements ReadModelService
type readModelService struct {
	readModelRepo repo.ReadModelProjectRepository
	orgRepo       repo.OrganizationRepository
	logger        *zap.Logger
}

// NewReadModelService creates a new read model service
func NewReadModelService(readModelRepo repo.ReadModelProjectRepository, orgRepo repo.OrganizationRepository, logger *zap.Logger) ReadModelService {
	return &readModelService{
		readModelRepo: readModelRepo,
		orgRepo:       orgRepo,
		logger:        logger,
	}
}

func (s *readModelService) CreateReadModelProject(ctx context.Context, userID, orgID uuid.UUID, key string, value map[string]interface{}) (*domain.ReadModelProject, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	project := &domain.ReadModelProject{
		ID:    uuid.New(),
		OrgID: orgID,
		Key:   key,
		Value: value,
	}

	createdProject, err := s.readModelRepo.CreateReadModelProject(ctx, project)
	if err != nil {
		s.logger.Error("Failed to create read model project", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Read model project created",
		zap.String("projectID", createdProject.ID.String()),
		zap.String("key", createdProject.Key),
		zap.String("orgID", createdProject.OrgID.String()))

	return createdProject, nil
}

func (s *readModelService) GetReadModelProject(ctx context.Context, userID, orgID uuid.UUID, key string) (*domain.ReadModelProject, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	project, err := s.readModelRepo.GetReadModelProject(ctx, orgID, key)
	if err != nil {
		s.logger.Error("Failed to get read model project", zap.Error(err))
		return nil, err
	}

	return project, nil
}

func (s *readModelService) GetReadModelProjectsByOrgID(ctx context.Context, userID, orgID uuid.UUID) ([]*domain.ReadModelProject, error) {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return nil, err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return nil, ErrUnauthorized
	}

	projects, err := s.readModelRepo.GetReadModelProjectsByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get read model projects by org ID", zap.Error(err))
		return nil, err
	}

	return projects, nil
}

func (s *readModelService) DeleteReadModelProject(ctx context.Context, userID, orgID uuid.UUID, key string) error {
	// Verify user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		s.logger.Error("Failed to get user role in organization", zap.Error(err))
		return err
	}

	if role == "" {
		s.logger.Warn("User does not have access to organization",
			zap.String("userID", userID.String()),
			zap.String("orgID", orgID.String()))
		return ErrUnauthorized
	}

	err = s.readModelRepo.DeleteReadModelProject(ctx, orgID, key)
	if err != nil {
		s.logger.Error("Failed to delete read model project", zap.Error(err))
		return err
	}

	s.logger.Info("Read model project deleted",
		zap.String("key", key),
		zap.String("orgID", orgID.String()))

	return nil
}
