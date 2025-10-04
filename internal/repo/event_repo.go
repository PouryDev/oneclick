package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

// EventRepository defines the interface for event logging operations
type EventRepository interface {
	CreateEventLog(ctx context.Context, event *domain.EventLog) (*domain.EventLog, error)
	GetEventLogsByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.EventLog, error)
	GetEventLogsByOrgIDAndAction(ctx context.Context, orgID uuid.UUID, action domain.EventAction, limit, offset int) ([]*domain.EventLog, error)
	GetEventLogsByOrgIDAndResourceType(ctx context.Context, orgID uuid.UUID, resourceType domain.ResourceType, limit, offset int) ([]*domain.EventLog, error)
	GetEventLogByID(ctx context.Context, id uuid.UUID) (*domain.EventLog, error)
	DeleteEventLog(ctx context.Context, id uuid.UUID) error
}

// DashboardCountsRepository defines the interface for dashboard counts operations
type DashboardCountsRepository interface {
	GetDashboardCounts(ctx context.Context, orgID uuid.UUID) (*domain.DashboardCounts, error)
	UpdateDashboardCounts(ctx context.Context, counts *domain.DashboardCounts) (*domain.DashboardCounts, error)
	DeleteDashboardCounts(ctx context.Context, orgID uuid.UUID) error
}

// ReadModelProjectRepository defines the interface for read model projections
type ReadModelProjectRepository interface {
	CreateReadModelProject(ctx context.Context, project *domain.ReadModelProject) (*domain.ReadModelProject, error)
	GetReadModelProject(ctx context.Context, orgID uuid.UUID, key string) (*domain.ReadModelProject, error)
	GetReadModelProjectsByOrgID(ctx context.Context, orgID uuid.UUID) ([]*domain.ReadModelProject, error)
	DeleteReadModelProject(ctx context.Context, orgID uuid.UUID, key string) error
}

// eventRepo implements EventRepository
type eventRepo struct {
	db *sql.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepo{db: db}
}

func (r *eventRepo) CreateEventLog(ctx context.Context, event *domain.EventLog) (*domain.EventLog, error) {
	query := `
		INSERT INTO event_logs (org_id, user_id, action, resource_type, resource_id, details)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, org_id, user_id, action, resource_type, resource_id, details, created_at
	`

	var createdEvent domain.EventLog
	err := r.db.QueryRowContext(ctx, query,
		event.OrgID,
		event.UserID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		event.Details,
	).Scan(
		&createdEvent.ID,
		&createdEvent.OrgID,
		&createdEvent.UserID,
		&createdEvent.Action,
		&createdEvent.ResourceType,
		&createdEvent.ResourceID,
		&createdEvent.Details,
		&createdEvent.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdEvent, nil
}

func (r *eventRepo) GetEventLogsByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.EventLog, error) {
	query := `
		SELECT id, org_id, user_id, action, resource_type, resource_id, details, created_at
		FROM event_logs
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.EventLog
	for rows.Next() {
		var event domain.EventLog
		err := rows.Scan(
			&event.ID,
			&event.OrgID,
			&event.UserID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&event.Details,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, nil
}

func (r *eventRepo) GetEventLogsByOrgIDAndAction(ctx context.Context, orgID uuid.UUID, action domain.EventAction, limit, offset int) ([]*domain.EventLog, error) {
	query := `
		SELECT id, org_id, user_id, action, resource_type, resource_id, details, created_at
		FROM event_logs
		WHERE org_id = $1 AND action = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, action, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.EventLog
	for rows.Next() {
		var event domain.EventLog
		err := rows.Scan(
			&event.ID,
			&event.OrgID,
			&event.UserID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&event.Details,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, nil
}

func (r *eventRepo) GetEventLogsByOrgIDAndResourceType(ctx context.Context, orgID uuid.UUID, resourceType domain.ResourceType, limit, offset int) ([]*domain.EventLog, error) {
	query := `
		SELECT id, org_id, user_id, action, resource_type, resource_id, details, created_at
		FROM event_logs
		WHERE org_id = $1 AND resource_type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, resourceType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.EventLog
	for rows.Next() {
		var event domain.EventLog
		err := rows.Scan(
			&event.ID,
			&event.OrgID,
			&event.UserID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&event.Details,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, nil
}

func (r *eventRepo) GetEventLogByID(ctx context.Context, id uuid.UUID) (*domain.EventLog, error) {
	query := `
		SELECT id, org_id, user_id, action, resource_type, resource_id, details, created_at
		FROM event_logs
		WHERE id = $1
	`

	var event domain.EventLog
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.OrgID,
		&event.UserID,
		&event.Action,
		&event.ResourceType,
		&event.ResourceID,
		&event.Details,
		&event.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &event, nil
}

func (r *eventRepo) DeleteEventLog(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM event_logs WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// dashboardCountsRepo implements DashboardCountsRepository
type dashboardCountsRepo struct {
	db *sql.DB
}

// NewDashboardCountsRepository creates a new dashboard counts repository
func NewDashboardCountsRepository(db *sql.DB) DashboardCountsRepository {
	return &dashboardCountsRepo{db: db}
}

func (r *dashboardCountsRepo) GetDashboardCounts(ctx context.Context, orgID uuid.UUID) (*domain.DashboardCounts, error) {
	query := `
		SELECT org_id, apps_count, clusters_count, running_pipelines, updated_at
		FROM dashboard_counts
		WHERE org_id = $1
	`

	var counts domain.DashboardCounts
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&counts.OrgID,
		&counts.AppsCount,
		&counts.ClustersCount,
		&counts.RunningPipelines,
		&counts.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &counts, nil
}

func (r *dashboardCountsRepo) UpdateDashboardCounts(ctx context.Context, counts *domain.DashboardCounts) (*domain.DashboardCounts, error) {
	query := `
		INSERT INTO dashboard_counts (org_id, apps_count, clusters_count, running_pipelines, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (org_id) DO UPDATE SET
			apps_count = EXCLUDED.apps_count,
			clusters_count = EXCLUDED.clusters_count,
			running_pipelines = EXCLUDED.running_pipelines,
			updated_at = EXCLUDED.updated_at
		RETURNING org_id, apps_count, clusters_count, running_pipelines, updated_at
	`

	var updatedCounts domain.DashboardCounts
	err := r.db.QueryRowContext(ctx, query,
		counts.OrgID,
		counts.AppsCount,
		counts.ClustersCount,
		counts.RunningPipelines,
	).Scan(
		&updatedCounts.OrgID,
		&updatedCounts.AppsCount,
		&updatedCounts.ClustersCount,
		&updatedCounts.RunningPipelines,
		&updatedCounts.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &updatedCounts, nil
}

func (r *dashboardCountsRepo) DeleteDashboardCounts(ctx context.Context, orgID uuid.UUID) error {
	query := `DELETE FROM dashboard_counts WHERE org_id = $1`
	_, err := r.db.ExecContext(ctx, query, orgID)
	return err
}

// readModelProjectRepo implements ReadModelProjectRepository
type readModelProjectRepo struct {
	db *sql.DB
}

// NewReadModelProjectRepository creates a new read model project repository
func NewReadModelProjectRepository(db *sql.DB) ReadModelProjectRepository {
	return &readModelProjectRepo{db: db}
}

func (r *readModelProjectRepo) CreateReadModelProject(ctx context.Context, project *domain.ReadModelProject) (*domain.ReadModelProject, error) {
	query := `
		INSERT INTO read_model_projects (org_id, key, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_id, key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = NOW()
		RETURNING id, org_id, key, value, created_at, updated_at
	`

	var createdProject domain.ReadModelProject
	err := r.db.QueryRowContext(ctx, query,
		project.OrgID,
		project.Key,
		project.Value,
	).Scan(
		&createdProject.ID,
		&createdProject.OrgID,
		&createdProject.Key,
		&createdProject.Value,
		&createdProject.CreatedAt,
		&createdProject.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdProject, nil
}

func (r *readModelProjectRepo) GetReadModelProject(ctx context.Context, orgID uuid.UUID, key string) (*domain.ReadModelProject, error) {
	query := `
		SELECT id, org_id, key, value, created_at, updated_at
		FROM read_model_projects
		WHERE org_id = $1 AND key = $2
	`

	var project domain.ReadModelProject
	err := r.db.QueryRowContext(ctx, query, orgID, key).Scan(
		&project.ID,
		&project.OrgID,
		&project.Key,
		&project.Value,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

func (r *readModelProjectRepo) GetReadModelProjectsByOrgID(ctx context.Context, orgID uuid.UUID) ([]*domain.ReadModelProject, error) {
	query := `
		SELECT id, org_id, key, value, created_at, updated_at
		FROM read_model_projects
		WHERE org_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.ReadModelProject
	for rows.Next() {
		var project domain.ReadModelProject
		err := rows.Scan(
			&project.ID,
			&project.OrgID,
			&project.Key,
			&project.Value,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func (r *readModelProjectRepo) DeleteReadModelProject(ctx context.Context, orgID uuid.UUID, key string) error {
	query := `DELETE FROM read_model_projects WHERE org_id = $1 AND key = $2`
	_, err := r.db.ExecContext(ctx, query, orgID, key)
	return err
}
