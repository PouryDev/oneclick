package worker

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// EventProjectorWorker handles projection of events to read models
type EventProjectorWorker struct {
	db                   *sql.DB
	eventRepo            repo.EventRepository
	dashboardCountsRepo  repo.DashboardCountsRepository
	readModelRepo        repo.ReadModelProjectRepository
	appRepo              repo.ApplicationRepository
	clusterRepo          repo.ClusterRepository
	pipelineRepo         repo.PipelineRepository
	logger               *zap.Logger
	projectionInterval   time.Duration
	lastProcessedEventID uuid.UUID
}

// NewEventProjectorWorker creates a new event projector worker
func NewEventProjectorWorker(
	db *sql.DB,
	eventRepo repo.EventRepository,
	dashboardCountsRepo repo.DashboardCountsRepository,
	readModelRepo repo.ReadModelProjectRepository,
	appRepo repo.ApplicationRepository,
	clusterRepo repo.ClusterRepository,
	pipelineRepo repo.PipelineRepository,
	logger *zap.Logger,
) *EventProjectorWorker {
	return &EventProjectorWorker{
		db:                  db,
		eventRepo:           eventRepo,
		dashboardCountsRepo: dashboardCountsRepo,
		readModelRepo:       readModelRepo,
		appRepo:             appRepo,
		clusterRepo:         clusterRepo,
		pipelineRepo:        pipelineRepo,
		logger:              logger,
		projectionInterval:  time.Minute, // Run every minute
	}
}

// Start starts the projector worker
func (w *EventProjectorWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting event projector worker")

	ticker := time.NewTicker(w.projectionInterval)
	defer ticker.Stop()

	// Run initial projection
	if err := w.runProjection(ctx); err != nil {
		w.logger.Error("Initial projection failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Event projector worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.runProjection(ctx); err != nil {
				w.logger.Error("Projection failed", zap.Error(err))
			}
		}
	}
}

// runProjection runs a single projection cycle
func (w *EventProjectorWorker) runProjection(ctx context.Context) error {
	w.logger.Debug("Running event projection")

	// Get all organizations
	orgs, err := w.getOrganizations(ctx)
	if err != nil {
		w.logger.Error("Failed to get organizations", zap.Error(err))
		return err
	}

	// Process each organization
	for _, orgID := range orgs {
		if err := w.projectOrganization(ctx, orgID); err != nil {
			w.logger.Error("Failed to project organization",
				zap.String("orgID", orgID.String()),
				zap.Error(err))
			// Continue with other organizations
		}
	}

	return nil
}

// getOrganizations gets all organization IDs from the database
func (w *EventProjectorWorker) getOrganizations(ctx context.Context) ([]uuid.UUID, error) {
	query := `SELECT id FROM organizations ORDER BY created_at`
	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []uuid.UUID
	for rows.Next() {
		var orgID uuid.UUID
		if err := rows.Scan(&orgID); err != nil {
			return nil, err
		}
		orgs = append(orgs, orgID)
	}

	return orgs, nil
}

// projectOrganization projects events for a specific organization
func (w *EventProjectorWorker) projectOrganization(ctx context.Context, orgID uuid.UUID) error {
	// Update dashboard counts
	if err := w.updateDashboardCounts(ctx, orgID); err != nil {
		w.logger.Error("Failed to update dashboard counts",
			zap.String("orgID", orgID.String()),
			zap.Error(err))
		return err
	}

	// Update read model projections
	if err := w.updateReadModelProjections(ctx, orgID); err != nil {
		w.logger.Error("Failed to update read model projections",
			zap.String("orgID", orgID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// updateDashboardCounts updates the dashboard counts for an organization
func (w *EventProjectorWorker) updateDashboardCounts(ctx context.Context, orgID uuid.UUID) error {
	// Get clusters count
	clusters, err := w.clusterRepo.GetClustersByOrgID(ctx, orgID)
	if err != nil {
		return err
	}

	// Get applications count by iterating through clusters
	var totalApps int
	for _, cluster := range clusters {
		apps, err := w.appRepo.GetApplicationsByClusterID(ctx, cluster.ID)
		if err != nil {
			return err
		}
		totalApps += len(apps)
	}

	// Get running pipelines count
	runningPipelines, err := w.getRunningPipelinesCount(ctx, orgID)
	if err != nil {
		return err
	}

	// Update dashboard counts
	counts := &domain.DashboardCounts{
		OrgID:            orgID,
		AppsCount:        totalApps,
		ClustersCount:    len(clusters),
		RunningPipelines: runningPipelines,
		UpdatedAt:        time.Now(),
	}

	_, err = w.dashboardCountsRepo.UpdateDashboardCounts(ctx, counts)
	if err != nil {
		return err
	}

	w.logger.Debug("Dashboard counts updated",
		zap.String("orgID", orgID.String()),
		zap.Int("appsCount", counts.AppsCount),
		zap.Int("clustersCount", counts.ClustersCount),
		zap.Int("runningPipelines", counts.RunningPipelines))

	return nil
}

// getRunningPipelinesCount gets the count of running pipelines for an organization
func (w *EventProjectorWorker) getRunningPipelinesCount(ctx context.Context, orgID uuid.UUID) (int, error) {
	// This is a simplified implementation
	// In a real scenario, you'd need to join pipelines with applications to get org-level counts
	query := `
		SELECT COUNT(*)
		FROM pipelines p
		JOIN applications a ON p.app_id = a.id
		WHERE a.org_id = $1 AND p.status IN ('pending', 'running')
	`

	var count int
	err := w.db.QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// updateReadModelProjections updates read model projections for an organization
func (w *EventProjectorWorker) updateReadModelProjections(ctx context.Context, orgID uuid.UUID) error {
	// Project recent failed pipelines
	if err := w.projectRecentFailedPipelines(ctx, orgID); err != nil {
		w.logger.Error("Failed to project recent failed pipelines", zap.Error(err))
	}

	// Project top apps by deployments
	if err := w.projectTopAppsByDeployments(ctx, orgID); err != nil {
		w.logger.Error("Failed to project top apps by deployments", zap.Error(err))
	}

	// Project cluster health summary
	if err := w.projectClusterHealthSummary(ctx, orgID); err != nil {
		w.logger.Error("Failed to project cluster health summary", zap.Error(err))
	}

	return nil
}

// projectRecentFailedPipelines projects recent failed pipelines
func (w *EventProjectorWorker) projectRecentFailedPipelines(ctx context.Context, orgID uuid.UUID) error {
	query := `
		SELECT p.id, p.app_id, p.commit_sha, p.finished_at, a.name as app_name
		FROM pipelines p
		JOIN applications a ON p.app_id = a.id
		WHERE a.org_id = $1 AND p.status = 'failed'
		ORDER BY p.finished_at DESC
		LIMIT 10
	`

	rows, err := w.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var failedPipelines []map[string]interface{}
	for rows.Next() {
		var pipelineID, appID uuid.UUID
		var commitSHA, appName string
		var finishedAt *time.Time

		err := rows.Scan(&pipelineID, &appID, &commitSHA, &finishedAt, &appName)
		if err != nil {
			return err
		}

		pipeline := map[string]interface{}{
			"id":          pipelineID.String(),
			"app_id":      appID.String(),
			"app_name":    appName,
			"commit_sha":  commitSHA,
			"finished_at": finishedAt,
		}
		failedPipelines = append(failedPipelines, pipeline)
	}

	// Store in read model
	project := &domain.ReadModelProject{
		ID:    uuid.New(),
		OrgID: orgID,
		Key:   "recent_failed_pipelines",
		Value: map[string]interface{}{
			"pipelines":  failedPipelines,
			"count":      len(failedPipelines),
			"updated_at": time.Now(),
		},
	}

	_, err = w.readModelRepo.CreateReadModelProject(ctx, project)
	return err
}

// projectTopAppsByDeployments projects top apps by deployment count
func (w *EventProjectorWorker) projectTopAppsByDeployments(ctx context.Context, orgID uuid.UUID) error {
	query := `
		SELECT a.id, a.name, COUNT(p.id) as deployment_count
		FROM applications a
		LEFT JOIN pipelines p ON a.id = p.app_id
		WHERE a.org_id = $1
		GROUP BY a.id, a.name
		ORDER BY deployment_count DESC
		LIMIT 10
	`

	rows, err := w.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var topApps []map[string]interface{}
	for rows.Next() {
		var appID uuid.UUID
		var appName string
		var deploymentCount int

		err := rows.Scan(&appID, &appName, &deploymentCount)
		if err != nil {
			return err
		}

		app := map[string]interface{}{
			"id":               appID.String(),
			"name":             appName,
			"deployment_count": deploymentCount,
		}
		topApps = append(topApps, app)
	}

	// Store in read model
	project := &domain.ReadModelProject{
		ID:    uuid.New(),
		OrgID: orgID,
		Key:   "top_apps_by_deployments",
		Value: map[string]interface{}{
			"apps":       topApps,
			"updated_at": time.Now(),
		},
	}

	_, err = w.readModelRepo.CreateReadModelProject(ctx, project)
	return err
}

// projectClusterHealthSummary projects cluster health summary
func (w *EventProjectorWorker) projectClusterHealthSummary(ctx context.Context, orgID uuid.UUID) error {
	query := `
		SELECT 
			c.id,
			c.name,
			c.status,
			COUNT(a.id) as app_count,
			c.created_at
		FROM clusters c
		LEFT JOIN applications a ON c.id = a.cluster_id
		WHERE c.org_id = $1
		GROUP BY c.id, c.name, c.status, c.created_at
		ORDER BY c.created_at DESC
	`

	rows, err := w.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var clusters []map[string]interface{}
	for rows.Next() {
		var clusterID uuid.UUID
		var clusterName, status string
		var appCount int
		var createdAt time.Time

		err := rows.Scan(&clusterID, &clusterName, &status, &appCount, &createdAt)
		if err != nil {
			return err
		}

		cluster := map[string]interface{}{
			"id":         clusterID.String(),
			"name":       clusterName,
			"status":     status,
			"app_count":  appCount,
			"created_at": createdAt,
		}
		clusters = append(clusters, cluster)
	}

	// Store in read model
	project := &domain.ReadModelProject{
		ID:    uuid.New(),
		OrgID: orgID,
		Key:   "cluster_health_summary",
		Value: map[string]interface{}{
			"clusters":   clusters,
			"count":      len(clusters),
			"updated_at": time.Now(),
		},
	}

	_, err = w.readModelRepo.CreateReadModelProject(ctx, project)
	return err
}
