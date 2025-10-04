package domain

import (
	"time"

	"github.com/google/uuid"
)

// GitServerType defines the type of git server
type GitServerType string

const (
	GitServerTypeGitea GitServerType = "gitea"
)

// GitServerStatus defines the status of a git server
type GitServerStatus string

const (
	GitServerStatusPending      GitServerStatus = "pending"
	GitServerStatusProvisioning GitServerStatus = "provisioning"
	GitServerStatusRunning      GitServerStatus = "running"
	GitServerStatusFailed       GitServerStatus = "failed"
	GitServerStatusStopped      GitServerStatus = "stopped"
)

// RunnerType defines the type of CI runner
type RunnerType string

const (
	RunnerTypeGitHub RunnerType = "github"
	RunnerTypeGitLab RunnerType = "gitlab"
	RunnerTypeCustom RunnerType = "custom"
)

// RunnerStatus defines the status of a CI runner
type RunnerStatus string

const (
	RunnerStatusPending      RunnerStatus = "pending"
	RunnerStatusProvisioning RunnerStatus = "provisioning"
	RunnerStatusRunning      RunnerStatus = "running"
	RunnerStatusFailed       RunnerStatus = "failed"
	RunnerStatusStopped      RunnerStatus = "stopped"
)

// JobType defines the type of background job
type JobType string

const (
	JobTypeGitServerInstall JobType = "git_server_install"
	JobTypeRunnerDeploy     JobType = "runner_deploy"
	JobTypeGitServerStop    JobType = "git_server_stop"
	JobTypeRunnerStop       JobType = "runner_stop"
)

// JobStatus defines the status of a background job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// GitServer represents a self-hosted git server instance
type GitServer struct {
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
	Type      GitServerType   `json:"type"`
	Domain    string          `json:"domain"`
	Storage   string          `json:"storage"`
	Status    GitServerStatus `json:"status"`
	Config    GitServerConfig `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GitServerConfig contains configuration for a git server
type GitServerConfig struct {
	AdminUser     string            `json:"admin_user,omitempty"`
	AdminPassword string            `json:"admin_password,omitempty"`
	AdminEmail    string            `json:"admin_email,omitempty"`
	Repositories  []string          `json:"repositories,omitempty"`
	Settings      map[string]string `json:"settings,omitempty"`
}

// Runner represents a CI runner instance
type Runner struct {
	ID        uuid.UUID    `json:"id"`
	OrgID     uuid.UUID    `json:"org_id"`
	Name      string       `json:"name"`
	Type      RunnerType   `json:"type"`
	Config    RunnerConfig `json:"config"`
	Status    RunnerStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// RunnerConfig contains configuration for a CI runner
type RunnerConfig struct {
	Labels       []string          `json:"labels,omitempty"`
	NodeSelector map[string]string `json:"node_selector,omitempty"`
	Resources    RunnerResources   `json:"resources,omitempty"`
	Token        string            `json:"token,omitempty"` // Encrypted token
	URL          string            `json:"url,omitempty"`
	Settings     map[string]string `json:"settings,omitempty"`
}

// RunnerResources defines resource limits for runners
type RunnerResources struct {
	CPU     string `json:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty"`
	Storage string `json:"storage,omitempty"`
}

// Job represents a background job in the queue
type Job struct {
	ID           uuid.UUID  `json:"id"`
	OrgID        uuid.UUID  `json:"org_id"`
	Type         JobType    `json:"type"`
	Status       JobStatus  `json:"status"`
	Payload      JobPayload `json:"payload"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// JobPayload contains the data for a background job
type JobPayload struct {
	GitServerID *uuid.UUID             `json:"git_server_id,omitempty"`
	RunnerID    *uuid.UUID             `json:"runner_id,omitempty"`
	DomainID    *uuid.UUID             `json:"domain_id,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// Request/Response DTOs

// CreateGitServerRequest is the request body for creating a git server
type CreateGitServerRequest struct {
	Type    GitServerType `json:"type" validate:"required,oneof=gitea"`
	Domain  string        `json:"domain" validate:"required,min=3,max=255"`
	Storage string        `json:"storage" validate:"required,min=1,max=50"`
}

// CreateRunnerRequest is the request body for creating a runner
type CreateRunnerRequest struct {
	Name         string            `json:"name" validate:"required,min=3,max=100"`
	Type         RunnerType        `json:"type" validate:"required,oneof=github gitlab custom"`
	Labels       []string          `json:"labels,omitempty"`
	NodeSelector map[string]string `json:"node_selector,omitempty"`
	Resources    RunnerResources   `json:"resources,omitempty"`
	Token        string            `json:"token,omitempty"`
	URL          string            `json:"url,omitempty"`
}

// GitServerResponse is the response body for git server details
type GitServerResponse struct {
	ID        uuid.UUID       `json:"id"`
	OrgID     uuid.UUID       `json:"org_id"`
	Type      GitServerType   `json:"type"`
	Domain    string          `json:"domain"`
	Storage   string          `json:"storage"`
	Status    GitServerStatus `json:"status"`
	Config    GitServerConfig `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// RunnerResponse is the response body for runner details
type RunnerResponse struct {
	ID        uuid.UUID    `json:"id"`
	OrgID     uuid.UUID    `json:"org_id"`
	Name      string       `json:"name"`
	Type      RunnerType   `json:"type"`
	Config    RunnerConfig `json:"config"`
	Status    RunnerStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// JobResponse is the response body for job details
type JobResponse struct {
	ID           uuid.UUID  `json:"id"`
	OrgID        uuid.UUID  `json:"org_id"`
	Type         JobType    `json:"type"`
	Status       JobStatus  `json:"status"`
	Payload      JobPayload `json:"payload"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// Conversion methods

// ToResponse converts GitServer to GitServerResponse
func (gs *GitServer) ToResponse() GitServerResponse {
	// Mask sensitive information
	config := gs.Config
	if config.AdminPassword != "" {
		config.AdminPassword = "***MASKED***"
	}

	return GitServerResponse{
		ID:        gs.ID,
		OrgID:     gs.OrgID,
		Type:      gs.Type,
		Domain:    gs.Domain,
		Storage:   gs.Storage,
		Status:    gs.Status,
		Config:    config,
		CreatedAt: gs.CreatedAt,
		UpdatedAt: gs.UpdatedAt,
	}
}

// ToResponse converts Runner to RunnerResponse
func (r *Runner) ToResponse() RunnerResponse {
	// Mask sensitive information
	config := r.Config
	if config.Token != "" {
		config.Token = "***MASKED***"
	}

	return RunnerResponse{
		ID:        r.ID,
		OrgID:     r.OrgID,
		Name:      r.Name,
		Type:      r.Type,
		Config:    config,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// ToResponse converts Job to JobResponse
func (j *Job) ToResponse() JobResponse {
	return JobResponse{
		ID:           j.ID,
		OrgID:        j.OrgID,
		Type:         j.Type,
		Status:       j.Status,
		Payload:      j.Payload,
		ErrorMessage: j.ErrorMessage,
		CreatedAt:    j.CreatedAt,
		StartedAt:    j.StartedAt,
		CompletedAt:  j.CompletedAt,
	}
}

// Validation helpers

// IsValidGitServerType checks if the git server type is valid
func IsValidGitServerType(t string) bool {
	return t == string(GitServerTypeGitea)
}

// IsValidRunnerType checks if the runner type is valid
func IsValidRunnerType(t string) bool {
	switch t {
	case string(RunnerTypeGitHub), string(RunnerTypeGitLab), string(RunnerTypeCustom):
		return true
	default:
		return false
	}
}

// IsValidGitServerStatus checks if the git server status is valid
func IsValidGitServerStatus(s string) bool {
	switch s {
	case string(GitServerStatusPending), string(GitServerStatusProvisioning),
		string(GitServerStatusRunning), string(GitServerStatusFailed), string(GitServerStatusStopped):
		return true
	default:
		return false
	}
}

// IsValidRunnerStatus checks if the runner status is valid
func IsValidRunnerStatus(s string) bool {
	switch s {
	case string(RunnerStatusPending), string(RunnerStatusProvisioning),
		string(RunnerStatusRunning), string(RunnerStatusFailed), string(RunnerStatusStopped):
		return true
	default:
		return false
	}
}

// IsValidJobType checks if the job type is valid
func IsValidJobType(t string) bool {
	switch t {
	case string(JobTypeGitServerInstall), string(JobTypeRunnerDeploy),
		string(JobTypeGitServerStop), string(JobTypeRunnerStop):
		return true
	default:
		return false
	}
}

// IsValidJobStatus checks if the job status is valid
func IsValidJobStatus(s string) bool {
	switch s {
	case string(JobStatusPending), string(JobStatusProcessing),
		string(JobStatusCompleted), string(JobStatusFailed):
		return true
	default:
		return false
	}
}
