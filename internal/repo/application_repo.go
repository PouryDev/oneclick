package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, app *domain.Application) (*domain.Application, error)
	GetApplicationByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	GetApplicationsByClusterID(ctx context.Context, clusterID uuid.UUID) ([]domain.ApplicationSummary, error)
	GetApplicationByNameInCluster(ctx context.Context, clusterID uuid.UUID, name string) (*domain.Application, error)
	DeleteApplication(ctx context.Context, id uuid.UUID) error
}

type ReleaseRepository interface {
	CreateRelease(ctx context.Context, release *domain.Release) (*domain.Release, error)
	GetReleaseByID(ctx context.Context, id uuid.UUID) (*domain.Release, error)
	GetReleasesByAppID(ctx context.Context, appID uuid.UUID) ([]domain.ReleaseSummary, error)
	GetLatestReleaseByAppID(ctx context.Context, appID uuid.UUID) (*domain.Release, error)
	UpdateReleaseStatus(ctx context.Context, id uuid.UUID, status domain.ReleaseStatus, startedAt, finishedAt *time.Time) (*domain.Release, error)
	UpdateReleaseMeta(ctx context.Context, id uuid.UUID, meta []byte) (*domain.Release, error)
	DeleteRelease(ctx context.Context, id uuid.UUID) error
}

type applicationRepository struct {
	db *sql.DB
}

type releaseRepository struct {
	db *sql.DB
}

func NewApplicationRepository(db *sql.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func NewReleaseRepository(db *sql.DB) ReleaseRepository {
	return &releaseRepository{db: db}
}

// Application Repository Implementation

func (r *applicationRepository) CreateApplication(ctx context.Context, app *domain.Application) (*domain.Application, error) {
	query := `
		INSERT INTO applications (org_id, cluster_id, name, repo_id, path, default_branch)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, org_id, cluster_id, name, repo_id, path, default_branch, created_at, updated_at
	`

	var createdApp domain.Application
	err := r.db.QueryRowContext(ctx, query,
		app.OrgID,
		app.ClusterID,
		app.Name,
		app.RepoID,
		app.Path,
		app.DefaultBranch,
	).Scan(
		&createdApp.ID,
		&createdApp.OrgID,
		&createdApp.ClusterID,
		&createdApp.Name,
		&createdApp.RepoID,
		&createdApp.Path,
		&createdApp.DefaultBranch,
		&createdApp.CreatedAt,
		&createdApp.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdApp, nil
}

func (r *applicationRepository) GetApplicationByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	query := `
		SELECT id, org_id, cluster_id, name, repo_id, path, default_branch, created_at, updated_at
		FROM applications
		WHERE id = $1
	`

	var app domain.Application
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&app.ID,
		&app.OrgID,
		&app.ClusterID,
		&app.Name,
		&app.RepoID,
		&app.Path,
		&app.DefaultBranch,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &app, nil
}

func (r *applicationRepository) GetApplicationsByClusterID(ctx context.Context, clusterID uuid.UUID) ([]domain.ApplicationSummary, error) {
	query := `
		SELECT id, org_id, cluster_id, name, repo_id, path, default_branch, created_at, updated_at
		FROM applications
		WHERE cluster_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []domain.ApplicationSummary
	for rows.Next() {
		var app domain.ApplicationSummary
		var orgID, clusterID uuid.UUID

		err := rows.Scan(
			&app.ID,
			&orgID,
			&clusterID,
			&app.Name,
			&app.RepoID,
			&app.Path,
			&app.DefaultBranch,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, nil
}

func (r *applicationRepository) GetApplicationByNameInCluster(ctx context.Context, clusterID uuid.UUID, name string) (*domain.Application, error) {
	query := `
		SELECT id, org_id, cluster_id, name, repo_id, path, default_branch, created_at, updated_at
		FROM applications
		WHERE cluster_id = $1 AND name = $2
	`

	var app domain.Application
	err := r.db.QueryRowContext(ctx, query, clusterID, name).Scan(
		&app.ID,
		&app.OrgID,
		&app.ClusterID,
		&app.Name,
		&app.RepoID,
		&app.Path,
		&app.DefaultBranch,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &app, nil
}

func (r *applicationRepository) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM applications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Release Repository Implementation

func (r *releaseRepository) CreateRelease(ctx context.Context, release *domain.Release) (*domain.Release, error) {
	query := `
		INSERT INTO releases (app_id, image, tag, created_by, status, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, app_id, image, tag, created_by, status, started_at, finished_at, meta, created_at, updated_at
	`

	var createdRelease domain.Release
	err := r.db.QueryRowContext(ctx, query,
		release.AppID,
		release.Image,
		release.Tag,
		release.CreatedBy,
		release.Status,
		release.Meta,
	).Scan(
		&createdRelease.ID,
		&createdRelease.AppID,
		&createdRelease.Image,
		&createdRelease.Tag,
		&createdRelease.CreatedBy,
		&createdRelease.Status,
		&createdRelease.StartedAt,
		&createdRelease.FinishedAt,
		&createdRelease.Meta,
		&createdRelease.CreatedAt,
		&createdRelease.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdRelease, nil
}

func (r *releaseRepository) GetReleaseByID(ctx context.Context, id uuid.UUID) (*domain.Release, error) {
	query := `
		SELECT id, app_id, image, tag, created_by, status, started_at, finished_at, meta, created_at, updated_at
		FROM releases
		WHERE id = $1
	`

	var release domain.Release
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&release.ID,
		&release.AppID,
		&release.Image,
		&release.Tag,
		&release.CreatedBy,
		&release.Status,
		&release.StartedAt,
		&release.FinishedAt,
		&release.Meta,
		&release.CreatedAt,
		&release.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &release, nil
}

func (r *releaseRepository) GetReleasesByAppID(ctx context.Context, appID uuid.UUID) ([]domain.ReleaseSummary, error) {
	query := `
		SELECT id, app_id, image, tag, created_by, status, started_at, finished_at, meta, created_at, updated_at
		FROM releases
		WHERE app_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var releases []domain.ReleaseSummary
	for rows.Next() {
		var release domain.ReleaseSummary
		var appID uuid.UUID
		var meta []byte

		err := rows.Scan(
			&release.ID,
			&appID,
			&release.Image,
			&release.Tag,
			&release.CreatedBy,
			&release.Status,
			&release.StartedAt,
			&release.FinishedAt,
			&meta,
			&release.CreatedAt,
			&release.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		releases = append(releases, release)
	}

	return releases, nil
}

func (r *releaseRepository) GetLatestReleaseByAppID(ctx context.Context, appID uuid.UUID) (*domain.Release, error) {
	query := `
		SELECT id, app_id, image, tag, created_by, status, started_at, finished_at, meta, created_at, updated_at
		FROM releases
		WHERE app_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var release domain.Release
	err := r.db.QueryRowContext(ctx, query, appID).Scan(
		&release.ID,
		&release.AppID,
		&release.Image,
		&release.Tag,
		&release.CreatedBy,
		&release.Status,
		&release.StartedAt,
		&release.FinishedAt,
		&release.Meta,
		&release.CreatedAt,
		&release.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &release, nil
}

func (r *releaseRepository) UpdateReleaseStatus(ctx context.Context, id uuid.UUID, status domain.ReleaseStatus, startedAt, finishedAt *time.Time) (*domain.Release, error) {
	query := `
		UPDATE releases
		SET status = $2, started_at = $3, finished_at = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, image, tag, created_by, status, started_at, finished_at, meta, created_at, updated_at
	`

	var release domain.Release
	err := r.db.QueryRowContext(ctx, query, id, status, startedAt, finishedAt).Scan(
		&release.ID,
		&release.AppID,
		&release.Image,
		&release.Tag,
		&release.CreatedBy,
		&release.Status,
		&release.StartedAt,
		&release.FinishedAt,
		&release.Meta,
		&release.CreatedAt,
		&release.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &release, nil
}

func (r *releaseRepository) UpdateReleaseMeta(ctx context.Context, id uuid.UUID, meta []byte) (*domain.Release, error) {
	query := `
		UPDATE releases
		SET meta = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, app_id, image, tag, created_by, status, started_at, finished_at, meta, created_at, updated_at
	`

	var release domain.Release
	err := r.db.QueryRowContext(ctx, query, id, meta).Scan(
		&release.ID,
		&release.AppID,
		&release.Image,
		&release.Tag,
		&release.CreatedBy,
		&release.Status,
		&release.StartedAt,
		&release.FinishedAt,
		&release.Meta,
		&release.CreatedAt,
		&release.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &release, nil
}

func (r *releaseRepository) DeleteRelease(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM releases WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
