package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

type RepositoryRepository interface {
	CreateRepository(ctx context.Context, repo *domain.Repository) (*domain.Repository, error)
	GetRepositoryByID(ctx context.Context, id uuid.UUID) (*domain.Repository, error)
	GetRepositoriesByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.RepositorySummary, error)
	GetRepositoryByURL(ctx context.Context, orgID uuid.UUID, url string) (*domain.Repository, error)
	UpdateRepositoryConfig(ctx context.Context, id uuid.UUID, config []byte) (*domain.Repository, error)
	DeleteRepository(ctx context.Context, id uuid.UUID) error
}

type repositoryRepository struct {
	db *sql.DB
}

func NewRepositoryRepository(db *sql.DB) RepositoryRepository {
	return &repositoryRepository{db: db}
}

func (r *repositoryRepository) CreateRepository(ctx context.Context, repo *domain.Repository) (*domain.Repository, error) {
	query := `
		INSERT INTO repositories (org_id, type, url, default_branch, config)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, org_id, type, url, default_branch, config, created_at, updated_at
	`

	var createdRepo domain.Repository
	err := r.db.QueryRowContext(ctx, query,
		repo.OrgID,
		repo.Type,
		repo.URL,
		repo.DefaultBranch,
		repo.Config,
	).Scan(
		&createdRepo.ID,
		&createdRepo.OrgID,
		&createdRepo.Type,
		&createdRepo.URL,
		&createdRepo.DefaultBranch,
		&createdRepo.Config,
		&createdRepo.CreatedAt,
		&createdRepo.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdRepo, nil
}

func (r *repositoryRepository) GetRepositoryByID(ctx context.Context, id uuid.UUID) (*domain.Repository, error) {
	query := `
		SELECT id, org_id, type, url, default_branch, config, created_at, updated_at
		FROM repositories
		WHERE id = $1
	`

	var repo domain.Repository
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&repo.ID,
		&repo.OrgID,
		&repo.Type,
		&repo.URL,
		&repo.DefaultBranch,
		&repo.Config,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &repo, nil
}

func (r *repositoryRepository) GetRepositoriesByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.RepositorySummary, error) {
	query := `
		SELECT id, org_id, type, url, default_branch, config, created_at, updated_at
		FROM repositories
		WHERE org_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []domain.RepositorySummary
	for rows.Next() {
		var repo domain.RepositorySummary
		var config []byte

		err := rows.Scan(
			&repo.ID,
			&repo.Type,
			&repo.URL,
			&repo.DefaultBranch,
			&config,
			&repo.CreatedAt,
			&repo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		repos = append(repos, repo)
	}

	return repos, nil
}

func (r *repositoryRepository) GetRepositoryByURL(ctx context.Context, orgID uuid.UUID, url string) (*domain.Repository, error) {
	query := `
		SELECT id, org_id, type, url, default_branch, config, created_at, updated_at
		FROM repositories
		WHERE org_id = $1 AND url = $2
	`

	var repo domain.Repository
	err := r.db.QueryRowContext(ctx, query, orgID, url).Scan(
		&repo.ID,
		&repo.OrgID,
		&repo.Type,
		&repo.URL,
		&repo.DefaultBranch,
		&repo.Config,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &repo, nil
}

func (r *repositoryRepository) UpdateRepositoryConfig(ctx context.Context, id uuid.UUID, config []byte) (*domain.Repository, error) {
	query := `
		UPDATE repositories
		SET config = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, type, url, default_branch, config, created_at, updated_at
	`

	var repo domain.Repository
	err := r.db.QueryRowContext(ctx, query, id, config).Scan(
		&repo.ID,
		&repo.OrgID,
		&repo.Type,
		&repo.URL,
		&repo.DefaultBranch,
		&repo.Config,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &repo, nil
}

func (r *repositoryRepository) DeleteRepository(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM repositories WHERE id = $1`

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
