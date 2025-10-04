package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/PouryDev/oneclick/internal/domain"
)

// GitServerRepository defines the interface for managing git servers
type GitServerRepository interface {
	CreateGitServer(ctx context.Context, gitServer *domain.GitServer) (*domain.GitServer, error)
	GetGitServerByID(ctx context.Context, id uuid.UUID) (*domain.GitServer, error)
	GetGitServersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.GitServer, error)
	GetGitServerByDomainInOrg(ctx context.Context, orgID uuid.UUID, domainName string) (*domain.GitServer, error)
	UpdateGitServerStatus(ctx context.Context, id uuid.UUID, status domain.GitServerStatus) (*domain.GitServer, error)
	UpdateGitServerConfig(ctx context.Context, id uuid.UUID, config domain.GitServerConfig) (*domain.GitServer, error)
	DeleteGitServer(ctx context.Context, id uuid.UUID) error
}

// RunnerRepository defines the interface for managing CI runners
type RunnerRepository interface {
	CreateRunner(ctx context.Context, runner *domain.Runner) (*domain.Runner, error)
	GetRunnerByID(ctx context.Context, id uuid.UUID) (*domain.Runner, error)
	GetRunnersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.Runner, error)
	GetRunnerByNameInOrg(ctx context.Context, orgID uuid.UUID, name string) (*domain.Runner, error)
	UpdateRunnerStatus(ctx context.Context, id uuid.UUID, status domain.RunnerStatus) (*domain.Runner, error)
	UpdateRunnerConfig(ctx context.Context, id uuid.UUID, config domain.RunnerConfig) (*domain.Runner, error)
	DeleteRunner(ctx context.Context, id uuid.UUID) error
}

// JobRepository defines the interface for managing background jobs
type JobRepository interface {
	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
	GetJobByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	GetJobsByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.Job, error)
	GetPendingJobs(ctx context.Context) ([]domain.Job, error)
	UpdateJobStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) (*domain.Job, error)
	StartJob(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	CompleteJob(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	FailJob(ctx context.Context, id uuid.UUID, errorMessage string) (*domain.Job, error)
	DeleteJob(ctx context.Context, id uuid.UUID) error
}

type gitServerRepo struct {
	db *sql.DB
}

type runnerRepo struct {
	db *sql.DB
}

type jobRepo struct {
	db *sql.DB
}

func NewGitServerRepository(db *sql.DB) GitServerRepository {
	return &gitServerRepo{db: db}
}

func NewRunnerRepository(db *sql.DB) RunnerRepository {
	return &runnerRepo{db: db}
}

func NewJobRepository(db *sql.DB) JobRepository {
	return &jobRepo{db: db}
}

// GitServer repository implementation
func (r *gitServerRepo) CreateGitServer(ctx context.Context, gitServer *domain.GitServer) (*domain.GitServer, error) {
	configBytes, err := json.Marshal(gitServer.Config)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO git_servers (org_id, type, domain, storage, status, config)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	var id uuid.UUID
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query,
		gitServer.OrgID,
		gitServer.Type,
		gitServer.Domain,
		gitServer.Storage,
		gitServer.Status,
		configBytes,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	gitServer.ID = id
	if createdAt.Valid {
		gitServer.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		gitServer.UpdatedAt = updatedAt.Time
	}

	return gitServer, nil
}

func (r *gitServerRepo) GetGitServerByID(ctx context.Context, id uuid.UUID) (*domain.GitServer, error) {
	query := `
		SELECT id, org_id, type, domain, storage, status, config, created_at, updated_at
		FROM git_servers
		WHERE id = $1`

	var gitServer domain.GitServer
	var configBytes []byte
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&gitServer.ID,
		&gitServer.OrgID,
		&gitServer.Type,
		&gitServer.Domain,
		&gitServer.Storage,
		&gitServer.Status,
		&configBytes,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &gitServer.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		gitServer.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		gitServer.UpdatedAt = updatedAt.Time
	}

	return &gitServer, nil
}

func (r *gitServerRepo) GetGitServersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.GitServer, error) {
	query := `
		SELECT id, org_id, type, domain, storage, status, config, created_at, updated_at
		FROM git_servers
		WHERE org_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gitServers []domain.GitServer
	for rows.Next() {
		var gitServer domain.GitServer
		var configBytes []byte
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&gitServer.ID,
			&gitServer.OrgID,
			&gitServer.Type,
			&gitServer.Domain,
			&gitServer.Storage,
			&gitServer.Status,
			&configBytes,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configBytes, &gitServer.Config); err != nil {
			return nil, err
		}

		if createdAt.Valid {
			gitServer.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			gitServer.UpdatedAt = updatedAt.Time
		}

		gitServers = append(gitServers, gitServer)
	}

	return gitServers, nil
}

func (r *gitServerRepo) GetGitServerByDomainInOrg(ctx context.Context, orgID uuid.UUID, domainName string) (*domain.GitServer, error) {
	query := `
		SELECT id, org_id, type, domain, storage, status, config, created_at, updated_at
		FROM git_servers
		WHERE org_id = $1 AND domain = $2`

	var gitServer domain.GitServer
	var configBytes []byte
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orgID, domainName).Scan(
		&gitServer.ID,
		&gitServer.OrgID,
		&gitServer.Type,
		&gitServer.Domain,
		&gitServer.Storage,
		&gitServer.Status,
		&configBytes,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &gitServer.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		gitServer.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		gitServer.UpdatedAt = updatedAt.Time
	}

	return &gitServer, nil
}

func (r *gitServerRepo) UpdateGitServerStatus(ctx context.Context, id uuid.UUID, status domain.GitServerStatus) (*domain.GitServer, error) {
	query := `
		UPDATE git_servers
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, type, domain, storage, status, config, created_at, updated_at`

	var gitServer domain.GitServer
	var configBytes []byte
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&gitServer.ID,
		&gitServer.OrgID,
		&gitServer.Type,
		&gitServer.Domain,
		&gitServer.Storage,
		&gitServer.Status,
		&configBytes,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &gitServer.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		gitServer.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		gitServer.UpdatedAt = updatedAt.Time
	}

	return &gitServer, nil
}

func (r *gitServerRepo) UpdateGitServerConfig(ctx context.Context, id uuid.UUID, config domain.GitServerConfig) (*domain.GitServer, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE git_servers
		SET config = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, type, domain, storage, status, config, created_at, updated_at`

	var gitServer domain.GitServer
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, id, configBytes).Scan(
		&gitServer.ID,
		&gitServer.OrgID,
		&gitServer.Type,
		&gitServer.Domain,
		&gitServer.Storage,
		&gitServer.Status,
		&configBytes,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &gitServer.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		gitServer.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		gitServer.UpdatedAt = updatedAt.Time
	}

	return &gitServer, nil
}

func (r *gitServerRepo) DeleteGitServer(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM git_servers WHERE id = $1`
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

// Runner repository implementation
func (r *runnerRepo) CreateRunner(ctx context.Context, runner *domain.Runner) (*domain.Runner, error) {
	configBytes, err := json.Marshal(runner.Config)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO runners (org_id, name, type, config, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	var id uuid.UUID
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query,
		runner.OrgID,
		runner.Name,
		runner.Type,
		configBytes,
		runner.Status,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	runner.ID = id
	if createdAt.Valid {
		runner.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		runner.UpdatedAt = updatedAt.Time
	}

	return runner, nil
}

func (r *runnerRepo) GetRunnerByID(ctx context.Context, id uuid.UUID) (*domain.Runner, error) {
	query := `
		SELECT id, org_id, name, type, config, status, created_at, updated_at
		FROM runners
		WHERE id = $1`

	var runner domain.Runner
	var configBytes []byte
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&runner.ID,
		&runner.OrgID,
		&runner.Name,
		&runner.Type,
		&configBytes,
		&runner.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &runner.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		runner.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		runner.UpdatedAt = updatedAt.Time
	}

	return &runner, nil
}

func (r *runnerRepo) GetRunnersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.Runner, error) {
	query := `
		SELECT id, org_id, name, type, config, status, created_at, updated_at
		FROM runners
		WHERE org_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runners []domain.Runner
	for rows.Next() {
		var runner domain.Runner
		var configBytes []byte
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&runner.ID,
			&runner.OrgID,
			&runner.Name,
			&runner.Type,
			&configBytes,
			&runner.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configBytes, &runner.Config); err != nil {
			return nil, err
		}

		if createdAt.Valid {
			runner.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			runner.UpdatedAt = updatedAt.Time
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

func (r *runnerRepo) GetRunnerByNameInOrg(ctx context.Context, orgID uuid.UUID, name string) (*domain.Runner, error) {
	query := `
		SELECT id, org_id, name, type, config, status, created_at, updated_at
		FROM runners
		WHERE org_id = $1 AND name = $2`

	var runner domain.Runner
	var configBytes []byte
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orgID, name).Scan(
		&runner.ID,
		&runner.OrgID,
		&runner.Name,
		&runner.Type,
		&configBytes,
		&runner.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &runner.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		runner.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		runner.UpdatedAt = updatedAt.Time
	}

	return &runner, nil
}

func (r *runnerRepo) UpdateRunnerStatus(ctx context.Context, id uuid.UUID, status domain.RunnerStatus) (*domain.Runner, error) {
	query := `
		UPDATE runners
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, name, type, config, status, created_at, updated_at`

	var runner domain.Runner
	var configBytes []byte
	var createdAt, updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&runner.ID,
		&runner.OrgID,
		&runner.Name,
		&runner.Type,
		&configBytes,
		&runner.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &runner.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		runner.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		runner.UpdatedAt = updatedAt.Time
	}

	return &runner, nil
}

func (r *runnerRepo) UpdateRunnerConfig(ctx context.Context, id uuid.UUID, config domain.RunnerConfig) (*domain.Runner, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE runners
		SET config = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, name, type, config, status, created_at, updated_at`

	var runner domain.Runner
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, id, configBytes).Scan(
		&runner.ID,
		&runner.OrgID,
		&runner.Name,
		&runner.Type,
		&configBytes,
		&runner.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &runner.Config); err != nil {
		return nil, err
	}

	if createdAt.Valid {
		runner.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		runner.UpdatedAt = updatedAt.Time
	}

	return &runner, nil
}

func (r *runnerRepo) DeleteRunner(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM runners WHERE id = $1`
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

// Job repository implementation
func (r *jobRepo) CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO job_queue (org_id, type, status, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	var id uuid.UUID
	var createdAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query,
		job.OrgID,
		job.Type,
		job.Status,
		payloadBytes,
	).Scan(&id, &createdAt)

	if err != nil {
		return nil, err
	}

	job.ID = id
	if createdAt.Valid {
		job.CreatedAt = createdAt.Time
	}

	return job, nil
}

func (r *jobRepo) GetJobByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	query := `
		SELECT id, org_id, type, status, payload, error_message, created_at, started_at, completed_at
		FROM job_queue
		WHERE id = $1`

	var job domain.Job
	var payloadBytes []byte
	var errorMessage sql.NullString
	var createdAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.OrgID,
		&job.Type,
		&job.Status,
		&payloadBytes,
		&errorMessage,
		&createdAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
		return nil, err
	}

	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}
	if createdAt.Valid {
		job.CreatedAt = createdAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

func (r *jobRepo) GetJobsByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.Job, error) {
	query := `
		SELECT id, org_id, type, status, payload, error_message, created_at, started_at, completed_at
		FROM job_queue
		WHERE org_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var payloadBytes []byte
		var errorMessage sql.NullString
		var createdAt, startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.OrgID,
			&job.Type,
			&job.Status,
			&payloadBytes,
			&errorMessage,
			&createdAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
			return nil, err
		}

		if errorMessage.Valid {
			job.ErrorMessage = errorMessage.String
		}
		if createdAt.Valid {
			job.CreatedAt = createdAt.Time
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *jobRepo) GetPendingJobs(ctx context.Context) ([]domain.Job, error) {
	query := `
		SELECT id, org_id, type, status, payload, error_message, created_at, started_at, completed_at
		FROM job_queue
		WHERE status = 'pending'
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		var payloadBytes []byte
		var errorMessage sql.NullString
		var createdAt, startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&job.ID,
			&job.OrgID,
			&job.Type,
			&job.Status,
			&payloadBytes,
			&errorMessage,
			&createdAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
			return nil, err
		}

		if errorMessage.Valid {
			job.ErrorMessage = errorMessage.String
		}
		if createdAt.Valid {
			job.CreatedAt = createdAt.Time
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *jobRepo) UpdateJobStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) (*domain.Job, error) {
	query := `
		UPDATE job_queue
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, type, status, payload, error_message, created_at, started_at, completed_at`

	var job domain.Job
	var payloadBytes []byte
	var errorMessage sql.NullString
	var createdAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, status).Scan(
		&job.ID,
		&job.OrgID,
		&job.Type,
		&job.Status,
		&payloadBytes,
		&errorMessage,
		&createdAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
		return nil, err
	}

	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}
	if createdAt.Valid {
		job.CreatedAt = createdAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

func (r *jobRepo) StartJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	query := `
		UPDATE job_queue
		SET status = 'processing', started_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id, org_id, type, status, payload, error_message, created_at, started_at, completed_at`

	var job domain.Job
	var payloadBytes []byte
	var errorMessage sql.NullString
	var createdAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.OrgID,
		&job.Type,
		&job.Status,
		&payloadBytes,
		&errorMessage,
		&createdAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
		return nil, err
	}

	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}
	if createdAt.Valid {
		job.CreatedAt = createdAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

func (r *jobRepo) CompleteJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	query := `
		UPDATE job_queue
		SET status = 'completed', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, type, status, payload, error_message, created_at, started_at, completed_at`

	var job domain.Job
	var payloadBytes []byte
	var errorMessage sql.NullString
	var createdAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.OrgID,
		&job.Type,
		&job.Status,
		&payloadBytes,
		&errorMessage,
		&createdAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
		return nil, err
	}

	if errorMessage.Valid {
		job.ErrorMessage = errorMessage.String
	}
	if createdAt.Valid {
		job.CreatedAt = createdAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

func (r *jobRepo) FailJob(ctx context.Context, id uuid.UUID, errorMessage string) (*domain.Job, error) {
	query := `
		UPDATE job_queue
		SET status = 'failed', error_message = $2, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, org_id, type, status, payload, error_message, created_at, started_at, completed_at`

	var job domain.Job
	var payloadBytes []byte
	var errorMsg sql.NullString
	var createdAt, startedAt, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, errorMessage).Scan(
		&job.ID,
		&job.OrgID,
		&job.Type,
		&job.Status,
		&payloadBytes,
		&errorMsg,
		&createdAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadBytes, &job.Payload); err != nil {
		return nil, err
	}

	if errorMsg.Valid {
		job.ErrorMessage = errorMsg.String
	}
	if createdAt.Valid {
		job.CreatedAt = createdAt.Time
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

func (r *jobRepo) DeleteJob(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM job_queue WHERE id = $1`
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
