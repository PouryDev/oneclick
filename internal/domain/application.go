package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Application represents a deployed application
type Application struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	ClusterID     uuid.UUID `json:"cluster_id"`
	Name          string    `json:"name"`
	RepoID        uuid.UUID `json:"repo_id"`
	Path          *string   `json:"path"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ApplicationSummary represents an application in list views
type ApplicationSummary struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	RepoID        uuid.UUID `json:"repo_id"`
	Path          *string   `json:"path"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ApplicationDetail represents detailed application information
type ApplicationDetail struct {
	Application
	CurrentRelease *ReleaseSummary `json:"current_release,omitempty"`
	ReleaseCount   int             `json:"release_count"`
	Status         string          `json:"status"`
}

// ReleaseStatus represents the status of a release
type ReleaseStatus string

const (
	ReleaseStatusPending   ReleaseStatus = "pending"
	ReleaseStatusRunning   ReleaseStatus = "running"
	ReleaseStatusSucceeded ReleaseStatus = "succeeded"
	ReleaseStatusFailed    ReleaseStatus = "failed"
)

// Release represents a deployment release
type Release struct {
	ID         uuid.UUID       `json:"id"`
	AppID      uuid.UUID       `json:"app_id"`
	Image      string          `json:"image"`
	Tag        string          `json:"tag"`
	CreatedBy  uuid.UUID       `json:"created_by"`
	Status     ReleaseStatus   `json:"status"`
	StartedAt  *time.Time      `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at"`
	Meta       json.RawMessage `json:"meta"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ReleaseSummary represents a release in list views
type ReleaseSummary struct {
	ID         uuid.UUID     `json:"id"`
	Image      string        `json:"image"`
	Tag        string        `json:"tag"`
	CreatedBy  uuid.UUID     `json:"created_by"`
	Status     ReleaseStatus `json:"status"`
	StartedAt  *time.Time    `json:"started_at"`
	FinishedAt *time.Time    `json:"finished_at"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// ReleaseMeta represents metadata for a release
type ReleaseMeta struct {
	CommitSHA     string            `json:"commit_sha,omitempty"`
	CommitMessage string            `json:"commit_message,omitempty"`
	Branch        string            `json:"branch,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	Config        map[string]string `json:"config,omitempty"`
}

// Request/Response DTOs
type CreateApplicationRequest struct {
	Name          string `json:"name" validate:"required"`
	RepoID        string `json:"repo_id" validate:"required,uuid"`
	Path          string `json:"path,omitempty"`
	DefaultBranch string `json:"default_branch" validate:"required"`
}

type ApplicationResponse struct {
	ID            uuid.UUID `json:"id"`
	OrgID         uuid.UUID `json:"org_id"`
	ClusterID     uuid.UUID `json:"cluster_id"`
	Name          string    `json:"name"`
	RepoID        uuid.UUID `json:"repo_id"`
	Path          *string   `json:"path"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeployApplicationRequest struct {
	Image string `json:"image,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type DeployApplicationResponse struct {
	ReleaseID uuid.UUID `json:"release_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

type ReleaseResponse struct {
	ID         uuid.UUID       `json:"id"`
	AppID      uuid.UUID       `json:"app_id"`
	Image      string          `json:"image"`
	Tag        string          `json:"tag"`
	CreatedBy  uuid.UUID       `json:"created_by"`
	Status     ReleaseStatus   `json:"status"`
	StartedAt  *time.Time      `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at"`
	Meta       json.RawMessage `json:"meta"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ValidReleaseStatuses returns a list of valid release statuses
func ValidReleaseStatuses() []ReleaseStatus {
	return []ReleaseStatus{
		ReleaseStatusPending,
		ReleaseStatusRunning,
		ReleaseStatusSucceeded,
		ReleaseStatusFailed,
	}
}

// IsValidReleaseStatus checks if a status is valid
func IsValidReleaseStatus(status ReleaseStatus) bool {
	for _, validStatus := range ValidReleaseStatuses() {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ToResponse converts an Application to ApplicationResponse
func (a *Application) ToResponse() ApplicationResponse {
	return ApplicationResponse{
		ID:            a.ID,
		OrgID:         a.OrgID,
		ClusterID:     a.ClusterID,
		Name:          a.Name,
		RepoID:        a.RepoID,
		Path:          a.Path,
		DefaultBranch: a.DefaultBranch,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// ToSummary converts an Application to ApplicationSummary
func (a *Application) ToSummary() ApplicationSummary {
	return ApplicationSummary{
		ID:            a.ID,
		Name:          a.Name,
		RepoID:        a.RepoID,
		Path:          a.Path,
		DefaultBranch: a.DefaultBranch,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// ToResponse converts a Release to ReleaseResponse
func (r *Release) ToResponse() ReleaseResponse {
	return ReleaseResponse{
		ID:         r.ID,
		AppID:      r.AppID,
		Image:      r.Image,
		Tag:        r.Tag,
		CreatedBy:  r.CreatedBy,
		Status:     r.Status,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
		Meta:       r.Meta,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// ToSummary converts a Release to ReleaseSummary
func (r *Release) ToSummary() ReleaseSummary {
	return ReleaseSummary{
		ID:         r.ID,
		Image:      r.Image,
		Tag:        r.Tag,
		CreatedBy:  r.CreatedBy,
		Status:     r.Status,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// GetMeta parses the release metadata JSON
func (r *Release) GetMeta() (*ReleaseMeta, error) {
	var meta ReleaseMeta
	if len(r.Meta) > 0 {
		err := json.Unmarshal(r.Meta, &meta)
		if err != nil {
			return nil, err
		}
	}
	return &meta, nil
}

// SetMeta sets the release metadata JSON
func (r *Release) SetMeta(meta *ReleaseMeta) error {
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	r.Meta = metaBytes
	return nil
}

// IsActive returns true if the release is currently active
func (r *Release) IsActive() bool {
	return r.Status == ReleaseStatusRunning || r.Status == ReleaseStatusSucceeded
}

// IsCompleted returns true if the release has finished (succeeded or failed)
func (r *Release) IsCompleted() bool {
	return r.Status == ReleaseStatusSucceeded || r.Status == ReleaseStatusFailed
}
