package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/PouryDev/oneclick/internal/domain"
)

// PipelineRepository defines the interface for pipeline operations
type PipelineRepository interface {
	CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) (*domain.Pipeline, error)
	GetPipelineByID(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error)
	GetPipelinesByAppID(ctx context.Context, appID uuid.UUID, limit, offset int) ([]domain.PipelineSummary, error)
	UpdatePipelineStatus(ctx context.Context, id uuid.UUID, status domain.PipelineStatus) (*domain.Pipeline, error)
	UpdatePipelineStarted(ctx context.Context, id uuid.UUID, status domain.PipelineStatus, startedAt *time.Time) (*domain.Pipeline, error)
	UpdatePipelineFinished(ctx context.Context, id uuid.UUID, status domain.PipelineStatus, finishedAt *time.Time) (*domain.Pipeline, error)
	UpdatePipelineLogsURL(ctx context.Context, id uuid.UUID, logsURL string) (*domain.Pipeline, error)
	DeletePipeline(ctx context.Context, id uuid.UUID) error
}

// PipelineStepRepository defines the interface for pipeline step operations
type PipelineStepRepository interface {
	CreatePipelineStep(ctx context.Context, step *domain.PipelineStep) (*domain.PipelineStep, error)
	GetPipelineStepsByPipelineID(ctx context.Context, pipelineID uuid.UUID) ([]domain.PipelineStep, error)
	UpdatePipelineStepStatus(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus) (*domain.PipelineStep, error)
	UpdatePipelineStepStarted(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus, startedAt *time.Time) (*domain.PipelineStep, error)
	UpdatePipelineStepFinished(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus, finishedAt *time.Time) (*domain.PipelineStep, error)
	UpdatePipelineStepLogs(ctx context.Context, id uuid.UUID, logs string) (*domain.PipelineStep, error)
	DeletePipelineStep(ctx context.Context, id uuid.UUID) error
}

// pipelineRepository implements PipelineRepository
type pipelineRepository struct {
	db *sqlx.DB
}

// pipelineStepRepository implements PipelineStepRepository
type pipelineStepRepository struct {
	db *sqlx.DB
}

// NewPipelineRepository creates a new pipeline repository
func NewPipelineRepository(db *sqlx.DB) PipelineRepository {
	return &pipelineRepository{db: db}
}

// NewPipelineStepRepository creates a new pipeline step repository
func NewPipelineStepRepository(db *sqlx.DB) PipelineStepRepository {
	return &pipelineStepRepository{db: db}
}

// CreatePipeline creates a new pipeline
func (r *pipelineRepository) CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) (*domain.Pipeline, error) {
	query := `
		INSERT INTO pipelines (app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta, created_at, updated_at
	`

	var metaJSON []byte
	var err error
	if pipeline.Meta != nil {
		metaJSON, err = json.Marshal(pipeline.Meta)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal meta: %w", err)
		}
	} else {
		metaJSON = []byte("{}")
	}

	var result domain.Pipeline
	err = r.db.QueryRowContext(ctx, query,
		pipeline.AppID,
		pipeline.RepoID,
		pipeline.CommitSHA,
		pipeline.Status,
		pipeline.TriggeredBy,
		pipeline.LogsURL,
		pipeline.StartedAt,
		pipeline.FinishedAt,
		metaJSON,
	).Scan(
		&result.ID,
		&result.AppID,
		&result.RepoID,
		&result.CommitSHA,
		&result.Status,
		&result.TriggeredBy,
		&result.LogsURL,
		&result.StartedAt,
		&result.FinishedAt,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Unmarshal meta
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &result.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &result, nil
}

// GetPipelineByID retrieves a pipeline by ID
func (r *pipelineRepository) GetPipelineByID(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error) {
	query := `
		SELECT id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta, created_at, updated_at
		FROM pipelines
		WHERE id = $1
	`

	var result domain.Pipeline
	var metaJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.AppID,
		&result.RepoID,
		&result.CommitSHA,
		&result.Status,
		&result.TriggeredBy,
		&result.LogsURL,
		&result.StartedAt,
		&result.FinishedAt,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Unmarshal meta
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &result.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &result, nil
}

// GetPipelinesByAppID retrieves pipelines for an application
func (r *pipelineRepository) GetPipelinesByAppID(ctx context.Context, appID uuid.UUID, limit, offset int) ([]domain.PipelineSummary, error) {
	query := `
		SELECT id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, created_at, updated_at
		FROM pipelines
		WHERE app_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, appID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipelines: %w", err)
	}
	defer rows.Close()

	var pipelines []domain.PipelineSummary
	for rows.Next() {
		var pipeline domain.PipelineSummary
		err := rows.Scan(
			&pipeline.ID,
			&pipeline.AppID,
			&pipeline.RepoID,
			&pipeline.CommitSHA,
			&pipeline.Status,
			&pipeline.TriggeredBy,
			&pipeline.LogsURL,
			&pipeline.StartedAt,
			&pipeline.FinishedAt,
			&pipeline.CreatedAt,
			&pipeline.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pipeline: %w", err)
		}
		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

// UpdatePipelineStatus updates the status of a pipeline
func (r *pipelineRepository) UpdatePipelineStatus(ctx context.Context, id uuid.UUID, status domain.PipelineStatus) (*domain.Pipeline, error) {
	query := `
		UPDATE pipelines
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta, created_at, updated_at
	`

	var result domain.Pipeline
	var metaJSON []byte
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&result.ID,
		&result.AppID,
		&result.RepoID,
		&result.CommitSHA,
		&result.Status,
		&result.TriggeredBy,
		&result.LogsURL,
		&result.StartedAt,
		&result.FinishedAt,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline status: %w", err)
	}

	// Unmarshal meta
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &result.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &result, nil
}

// UpdatePipelineStarted updates the pipeline when it starts
func (r *pipelineRepository) UpdatePipelineStarted(ctx context.Context, id uuid.UUID, status domain.PipelineStatus, startedAt *time.Time) (*domain.Pipeline, error) {
	query := `
		UPDATE pipelines
		SET status = $2, started_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta, created_at, updated_at
	`

	var result domain.Pipeline
	var metaJSON []byte
	err := r.db.QueryRowContext(ctx, query, id, status, startedAt).Scan(
		&result.ID,
		&result.AppID,
		&result.RepoID,
		&result.CommitSHA,
		&result.Status,
		&result.TriggeredBy,
		&result.LogsURL,
		&result.StartedAt,
		&result.FinishedAt,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline started: %w", err)
	}

	// Unmarshal meta
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &result.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &result, nil
}

// UpdatePipelineFinished updates the pipeline when it finishes
func (r *pipelineRepository) UpdatePipelineFinished(ctx context.Context, id uuid.UUID, status domain.PipelineStatus, finishedAt *time.Time) (*domain.Pipeline, error) {
	query := `
		UPDATE pipelines
		SET status = $2, finished_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta, created_at, updated_at
	`

	var result domain.Pipeline
	var metaJSON []byte
	err := r.db.QueryRowContext(ctx, query, id, status, finishedAt).Scan(
		&result.ID,
		&result.AppID,
		&result.RepoID,
		&result.CommitSHA,
		&result.Status,
		&result.TriggeredBy,
		&result.LogsURL,
		&result.StartedAt,
		&result.FinishedAt,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline finished: %w", err)
	}

	// Unmarshal meta
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &result.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &result, nil
}

// UpdatePipelineLogsURL updates the logs URL of a pipeline
func (r *pipelineRepository) UpdatePipelineLogsURL(ctx context.Context, id uuid.UUID, logsURL string) (*domain.Pipeline, error) {
	query := `
		UPDATE pipelines
		SET logs_url = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, repo_id, commit_sha, status, triggered_by, logs_url, started_at, finished_at, meta, created_at, updated_at
	`

	var result domain.Pipeline
	var metaJSON []byte
	err := r.db.QueryRowContext(ctx, query, id, logsURL).Scan(
		&result.ID,
		&result.AppID,
		&result.RepoID,
		&result.CommitSHA,
		&result.Status,
		&result.TriggeredBy,
		&result.LogsURL,
		&result.StartedAt,
		&result.FinishedAt,
		&metaJSON,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline logs URL: %w", err)
	}

	// Unmarshal meta
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &result.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &result, nil
}

// DeletePipeline deletes a pipeline
func (r *pipelineRepository) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pipelines WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete pipeline: %w", err)
	}
	return nil
}

// CreatePipelineStep creates a new pipeline step
func (r *pipelineStepRepository) CreatePipelineStep(ctx context.Context, step *domain.PipelineStep) (*domain.PipelineStep, error) {
	query := `
		INSERT INTO pipeline_steps (pipeline_id, name, status, started_at, finished_at, logs)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, pipeline_id, name, status, started_at, finished_at, logs, created_at, updated_at
	`

	var result domain.PipelineStep
	err := r.db.QueryRowContext(ctx, query,
		step.PipelineID,
		step.Name,
		step.Status,
		step.StartedAt,
		step.FinishedAt,
		step.Logs,
	).Scan(
		&result.ID,
		&result.PipelineID,
		&result.Name,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.Logs,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline step: %w", err)
	}

	return &result, nil
}

// GetPipelineStepsByPipelineID retrieves all steps for a pipeline
func (r *pipelineStepRepository) GetPipelineStepsByPipelineID(ctx context.Context, pipelineID uuid.UUID) ([]domain.PipelineStep, error) {
	query := `
		SELECT id, pipeline_id, name, status, started_at, finished_at, logs, created_at, updated_at
		FROM pipeline_steps
		WHERE pipeline_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline steps: %w", err)
	}
	defer rows.Close()

	var steps []domain.PipelineStep
	for rows.Next() {
		var step domain.PipelineStep
		err := rows.Scan(
			&step.ID,
			&step.PipelineID,
			&step.Name,
			&step.Status,
			&step.StartedAt,
			&step.FinishedAt,
			&step.Logs,
			&step.CreatedAt,
			&step.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pipeline step: %w", err)
		}
		steps = append(steps, step)
	}

	return steps, nil
}

// UpdatePipelineStepStatus updates the status of a pipeline step
func (r *pipelineStepRepository) UpdatePipelineStepStatus(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus) (*domain.PipelineStep, error) {
	query := `
		UPDATE pipeline_steps
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, pipeline_id, name, status, started_at, finished_at, logs, created_at, updated_at
	`

	var result domain.PipelineStep
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&result.ID,
		&result.PipelineID,
		&result.Name,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.Logs,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline step status: %w", err)
	}

	return &result, nil
}

// UpdatePipelineStepStarted updates the pipeline step when it starts
func (r *pipelineStepRepository) UpdatePipelineStepStarted(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus, startedAt *time.Time) (*domain.PipelineStep, error) {
	query := `
		UPDATE pipeline_steps
		SET status = $2, started_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, pipeline_id, name, status, started_at, finished_at, logs, created_at, updated_at
	`

	var result domain.PipelineStep
	err := r.db.QueryRowContext(ctx, query, id, status, startedAt).Scan(
		&result.ID,
		&result.PipelineID,
		&result.Name,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.Logs,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline step started: %w", err)
	}

	return &result, nil
}

// UpdatePipelineStepFinished updates the pipeline step when it finishes
func (r *pipelineStepRepository) UpdatePipelineStepFinished(ctx context.Context, id uuid.UUID, status domain.PipelineStepStatus, finishedAt *time.Time) (*domain.PipelineStep, error) {
	query := `
		UPDATE pipeline_steps
		SET status = $2, finished_at = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, pipeline_id, name, status, started_at, finished_at, logs, created_at, updated_at
	`

	var result domain.PipelineStep
	err := r.db.QueryRowContext(ctx, query, id, status, finishedAt).Scan(
		&result.ID,
		&result.PipelineID,
		&result.Name,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.Logs,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline step finished: %w", err)
	}

	return &result, nil
}

// UpdatePipelineStepLogs updates the logs of a pipeline step
func (r *pipelineStepRepository) UpdatePipelineStepLogs(ctx context.Context, id uuid.UUID, logs string) (*domain.PipelineStep, error) {
	query := `
		UPDATE pipeline_steps
		SET logs = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, pipeline_id, name, status, started_at, finished_at, logs, created_at, updated_at
	`

	var result domain.PipelineStep
	err := r.db.QueryRowContext(ctx, query, id, logs).Scan(
		&result.ID,
		&result.PipelineID,
		&result.Name,
		&result.Status,
		&result.StartedAt,
		&result.FinishedAt,
		&result.Logs,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update pipeline step logs: %w", err)
	}

	return &result, nil
}

// DeletePipelineStep deletes a pipeline step
func (r *pipelineStepRepository) DeletePipelineStep(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pipeline_steps WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete pipeline step: %w", err)
	}
	return nil
}
