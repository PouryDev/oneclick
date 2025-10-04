package domain

import (
	"time"

	"github.com/google/uuid"
)

// PipelineStatus represents the status of a pipeline
type PipelineStatus string

const (
	PipelineStatusPending   PipelineStatus = "pending"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusSuccess   PipelineStatus = "success"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCancelled PipelineStatus = "cancelled"
)

// PipelineStepStatus represents the status of a pipeline step
type PipelineStepStatus string

const (
	PipelineStepStatusPending   PipelineStepStatus = "pending"
	PipelineStepStatusRunning   PipelineStepStatus = "running"
	PipelineStepStatusSuccess   PipelineStepStatus = "success"
	PipelineStepStatusFailed    PipelineStepStatus = "failed"
	PipelineStepStatusSkipped   PipelineStepStatus = "skipped"
	PipelineStepStatusCancelled PipelineStepStatus = "cancelled"
)

// PipelineTriggerType represents how a pipeline was triggered
type PipelineTriggerType string

const (
	PipelineTriggerTypeManual   PipelineTriggerType = "manual"
	PipelineTriggerTypeWebhook  PipelineTriggerType = "webhook"
	PipelineTriggerTypeSchedule PipelineTriggerType = "schedule"
)

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID          uuid.UUID              `json:"id"`
	AppID       uuid.UUID              `json:"app_id"`
	RepoID      uuid.UUID              `json:"repo_id"`
	CommitSHA   string                 `json:"commit_sha"`
	Status      PipelineStatus         `json:"status"`
	TriggeredBy uuid.UUID              `json:"triggered_by"`
	LogsURL     string                 `json:"logs_url"`
	StartedAt   *time.Time             `json:"started_at"`
	FinishedAt  *time.Time             `json:"finished_at"`
	Meta        map[string]interface{} `json:"meta"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PipelineStep represents a step within a pipeline
type PipelineStep struct {
	ID         uuid.UUID          `json:"id"`
	PipelineID uuid.UUID          `json:"pipeline_id"`
	Name       string             `json:"name"`
	Status     PipelineStepStatus `json:"status"`
	StartedAt  *time.Time         `json:"started_at"`
	FinishedAt *time.Time         `json:"finished_at"`
	Logs       string             `json:"logs"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// CreatePipelineRequest represents the request to create a new pipeline
type CreatePipelineRequest struct {
	Ref       string `json:"ref" binding:"required"`        // Branch or tag reference
	CommitSHA string `json:"commit_sha" binding:"required"` // Git commit SHA
}

// PipelineResponse represents the response for pipeline details
type PipelineResponse struct {
	ID          uuid.UUID              `json:"id"`
	AppID       uuid.UUID              `json:"app_id"`
	RepoID      uuid.UUID              `json:"repo_id"`
	CommitSHA   string                 `json:"commit_sha"`
	Status      PipelineStatus         `json:"status"`
	TriggeredBy uuid.UUID              `json:"triggered_by"`
	LogsURL     string                 `json:"logs_url"`
	StartedAt   *time.Time             `json:"started_at"`
	FinishedAt  *time.Time             `json:"finished_at"`
	Meta        map[string]interface{} `json:"meta"`
	Steps       []PipelineStep         `json:"steps"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PipelineSummary represents a summary of pipeline information
type PipelineSummary struct {
	ID          uuid.UUID      `json:"id"`
	AppID       uuid.UUID      `json:"app_id"`
	RepoID      uuid.UUID      `json:"repo_id"`
	CommitSHA   string         `json:"commit_sha"`
	Status      PipelineStatus `json:"status"`
	TriggeredBy uuid.UUID      `json:"triggered_by"`
	LogsURL     string         `json:"logs_url"`
	StartedAt   *time.Time     `json:"started_at"`
	FinishedAt  *time.Time     `json:"finished_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PipelineLogsResponse represents the response for pipeline logs
type PipelineLogsResponse struct {
	PipelineID uuid.UUID      `json:"pipeline_id"`
	Logs       string         `json:"logs"`
	Steps      []PipelineStep `json:"steps"`
}

// PipelineJobPayload represents the payload for pipeline jobs
type PipelineJobPayload struct {
	PipelineID  uuid.UUID              `json:"pipeline_id"`
	AppID       uuid.UUID              `json:"app_id"`
	RepoID      uuid.UUID              `json:"repo_id"`
	CommitSHA   string                 `json:"commit_sha"`
	Ref         string                 `json:"ref"`
	TriggeredBy uuid.UUID              `json:"triggered_by"`
	Config      map[string]interface{} `json:"config"`
}

// IsValidPipelineStatus checks if the pipeline status is valid
func IsValidPipelineStatus(s string) bool {
	switch s {
	case string(PipelineStatusPending), string(PipelineStatusRunning),
		string(PipelineStatusSuccess), string(PipelineStatusFailed),
		string(PipelineStatusCancelled):
		return true
	default:
		return false
	}
}

// IsValidPipelineStepStatus checks if the pipeline step status is valid
func IsValidPipelineStepStatus(s string) bool {
	switch s {
	case string(PipelineStepStatusPending), string(PipelineStepStatusRunning),
		string(PipelineStepStatusSuccess), string(PipelineStepStatusFailed),
		string(PipelineStepStatusSkipped), string(PipelineStepStatusCancelled):
		return true
	default:
		return false
	}
}

// IsValidPipelineTriggerType checks if the pipeline trigger type is valid
func IsValidPipelineTriggerType(t string) bool {
	switch t {
	case string(PipelineTriggerTypeManual), string(PipelineTriggerTypeWebhook),
		string(PipelineTriggerTypeSchedule):
		return true
	default:
		return false
	}
}

// ToResponse converts a Pipeline to PipelineResponse
func (p *Pipeline) ToResponse(steps []PipelineStep) PipelineResponse {
	return PipelineResponse{
		ID:          p.ID,
		AppID:       p.AppID,
		RepoID:      p.RepoID,
		CommitSHA:   p.CommitSHA,
		Status:      p.Status,
		TriggeredBy: p.TriggeredBy,
		LogsURL:     p.LogsURL,
		StartedAt:   p.StartedAt,
		FinishedAt:  p.FinishedAt,
		Meta:        p.Meta,
		Steps:       steps,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToSummary converts a Pipeline to PipelineSummary
func (p *Pipeline) ToSummary() PipelineSummary {
	return PipelineSummary{
		ID:          p.ID,
		AppID:       p.AppID,
		RepoID:      p.RepoID,
		CommitSHA:   p.CommitSHA,
		Status:      p.Status,
		TriggeredBy: p.TriggeredBy,
		LogsURL:     p.LogsURL,
		StartedAt:   p.StartedAt,
		FinishedAt:  p.FinishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// GetDuration returns the duration of the pipeline if it has finished
func (p *Pipeline) GetDuration() *time.Duration {
	if p.StartedAt != nil && p.FinishedAt != nil {
		duration := p.FinishedAt.Sub(*p.StartedAt)
		return &duration
	}
	return nil
}

// IsFinished returns true if the pipeline has finished (success, failed, or cancelled)
func (p *Pipeline) IsFinished() bool {
	return p.Status == PipelineStatusSuccess ||
		p.Status == PipelineStatusFailed ||
		p.Status == PipelineStatusCancelled
}

// IsRunning returns true if the pipeline is currently running
func (p *Pipeline) IsRunning() bool {
	return p.Status == PipelineStatusRunning
}

// IsPending returns true if the pipeline is pending
func (p *Pipeline) IsPending() bool {
	return p.Status == PipelineStatusPending
}
