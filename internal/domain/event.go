package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventAction represents the type of action performed
type EventAction string

const (
	EventActionAppCreated        EventAction = "app_created"
	EventActionAppUpdated        EventAction = "app_updated"
	EventActionAppDeleted        EventAction = "app_deleted"
	EventActionClusterImported   EventAction = "cluster_imported"
	EventActionClusterUpdated    EventAction = "cluster_updated"
	EventActionClusterDeleted    EventAction = "cluster_deleted"
	EventActionPipelineStarted   EventAction = "pipeline_started"
	EventActionPipelineCompleted EventAction = "pipeline_completed"
	EventActionPipelineFailed    EventAction = "pipeline_failed"
	EventActionReleaseCreated    EventAction = "release_created"
	EventActionReleaseUpdated    EventAction = "release_updated"
	EventActionReleaseDeleted    EventAction = "release_deleted"
	EventActionUserJoined        EventAction = "user_joined"
	EventActionUserLeft          EventAction = "user_left"
)

// ResourceType represents the type of resource affected by the event
type ResourceType string

const (
	ResourceTypeApp      ResourceType = "app"
	ResourceTypeCluster  ResourceType = "cluster"
	ResourceTypePipeline ResourceType = "pipeline"
	ResourceTypeRelease  ResourceType = "release"
	ResourceTypeUser     ResourceType = "user"
	ResourceTypeOrg      ResourceType = "organization"
)

// EventLog represents an audit event in the system
type EventLog struct {
	ID           uuid.UUID              `json:"id"`
	OrgID        uuid.UUID              `json:"org_id"`
	UserID       uuid.UUID              `json:"user_id"`
	Action       EventAction            `json:"action"`
	ResourceType ResourceType           `json:"resource_type"`
	ResourceID   uuid.UUID              `json:"resource_id"`
	Details      map[string]interface{} `json:"details"`
	CreatedAt    time.Time              `json:"created_at"`
}

// CreateEventRequest represents a request to create a new event
type CreateEventRequest struct {
	OrgID        uuid.UUID              `json:"org_id"`
	UserID       uuid.UUID              `json:"user_id"`
	Action       EventAction            `json:"action"`
	ResourceType ResourceType           `json:"resource_type"`
	ResourceID   uuid.UUID              `json:"resource_id"`
	Details      map[string]interface{} `json:"details"`
}

// EventResponse represents an event in API responses
type EventResponse struct {
	ID           uuid.UUID              `json:"id"`
	OrgID        uuid.UUID              `json:"org_id"`
	UserID       uuid.UUID              `json:"user_id"`
	Action       EventAction            `json:"action"`
	ResourceType ResourceType           `json:"resource_type"`
	ResourceID   uuid.UUID              `json:"resource_id"`
	Details      map[string]interface{} `json:"details"`
	CreatedAt    time.Time              `json:"created_at"`
}

// DashboardCounts represents the aggregated counts for dashboard
type DashboardCounts struct {
	OrgID            uuid.UUID `json:"org_id"`
	AppsCount        int       `json:"apps_count"`
	ClustersCount    int       `json:"clusters_count"`
	RunningPipelines int       `json:"running_pipelines"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ReadModelProject represents a generic read model projection
type ReadModelProject struct {
	ID        uuid.UUID              `json:"id"`
	OrgID     uuid.UUID              `json:"org_id"`
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// EventSummary represents a summary of events for the UI
type EventSummary struct {
	ID           uuid.UUID    `json:"id"`
	Action       EventAction  `json:"action"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   uuid.UUID    `json:"resource_id"`
	UserID       uuid.UUID    `json:"user_id"`
	Details      string       `json:"details"` // Short description
	CreatedAt    time.Time    `json:"created_at"`
}

// ToResponse converts EventLog to EventResponse
func (e *EventLog) ToResponse() *EventResponse {
	return &EventResponse{
		ID:           e.ID,
		OrgID:        e.OrgID,
		UserID:       e.UserID,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		Details:      e.Details,
		CreatedAt:    e.CreatedAt,
	}
}

// ToSummary converts EventLog to EventSummary
func (e *EventLog) ToSummary() *EventSummary {
	details := ""
	if e.Details != nil {
		if name, ok := e.Details["name"].(string); ok {
			details = name
		} else if message, ok := e.Details["message"].(string); ok {
			details = message
		}
	}

	return &EventSummary{
		ID:           e.ID,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		UserID:       e.UserID,
		Details:      details,
		CreatedAt:    e.CreatedAt,
	}
}
