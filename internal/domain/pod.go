package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Pod represents a Kubernetes pod
type Pod struct {
	ID        uuid.UUID         `json:"id"`
	AppID     uuid.UUID         `json:"app_id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Restarts  int32             `json:"restarts"`
	Ready     string            `json:"ready"` // "1/2" format
	Age       string            `json:"age"`
	NodeName  string            `json:"node_name"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// PodDetail represents detailed pod information
type PodDetail struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Restarts   int32             `json:"restarts"`
	Ready      string            `json:"ready"`
	Age        string            `json:"age"`
	NodeName   string            `json:"node_name"`
	Labels     map[string]string `json:"labels"`
	Containers []ContainerInfo   `json:"containers"`
	Events     []EventInfo       `json:"events"`
	OwnerRefs  []OwnerReference  `json:"owner_refs"`
	CreatedAt  time.Time         `json:"created_at"`
	IP         string            `json:"ip"`
	HostIP     string            `json:"host_ip"`
	Phase      string            `json:"phase"`
	Conditions []PodCondition    `json:"conditions"`
}

// ContainerInfo represents container information
type ContainerInfo struct {
	Name         string     `json:"name"`
	Image        string     `json:"image"`
	Ready        bool       `json:"ready"`
	RestartCount int32      `json:"restart_count"`
	State        string     `json:"state"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
}

// EventInfo represents Kubernetes event information
type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// OwnerReference represents owner reference information
type OwnerReference struct {
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	APIVersion string `json:"api_version"`
	Controller bool   `json:"controller"`
}

// PodCondition represents pod condition information
type PodCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
}

// PodLogsRequest represents a request for pod logs
type PodLogsRequest struct {
	Container string `form:"container" binding:"omitempty"`
	Follow    bool   `form:"follow" binding:"omitempty"`
	TailLines int64  `form:"tailLines" binding:"omitempty,min=1,max=10000"`
}

// PodExecRequest represents a request for pod exec
type PodExecRequest struct {
	Container string   `json:"container" binding:"required"`
	Command   []string `json:"command" binding:"required"`
	TTY       bool     `json:"tty"`
	Stdin     bool     `json:"stdin"`
}

// PodLogsResponse represents the response for pod logs
type PodLogsResponse struct {
	PodName   string `json:"pod_name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	Logs      string `json:"logs"`
	Follow    bool   `json:"follow"`
}

// PodDescribeResponse represents the response for pod describe
type PodDescribeResponse struct {
	PodDetail PodDetail `json:"pod_detail"`
	RawYAML   string    `json:"raw_yaml,omitempty"`
}

// PodListResponse represents the response for pod list
type PodListResponse struct {
	Pods  []Pod `json:"pods"`
	Total int   `json:"total"`
}

// AuditLogEntry represents an audit log entry for pod access
type AuditLogEntry struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"`   // "pod_list", "pod_detail", "pod_logs", "pod_exec"
	Resource  string    `json:"resource"` // pod name
	Namespace string    `json:"namespace"`
	ClusterID uuid.UUID `json:"cluster_id"`
	AppID     uuid.UUID `json:"app_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

// PodService defines the interface for pod management
type PodService interface {
	GetPodsByApp(ctx context.Context, userID, appID uuid.UUID) ([]Pod, error)
	GetPodDetail(ctx context.Context, userID, podName, namespace string) (*PodDetail, error)
	GetPodLogs(ctx context.Context, userID, podName, namespace string, req PodLogsRequest) (*PodLogsResponse, error)
	GetPodDescribe(ctx context.Context, userID, podName, namespace string) (*PodDescribeResponse, error)
	ExecInPod(ctx context.Context, userID, podName, namespace string, req PodExecRequest) error
	CreateAuditLog(ctx context.Context, userID uuid.UUID, action, resource, namespace string, clusterID, appID uuid.UUID, ipAddress, userAgent string) error
}

// PodRepository defines the interface for pod data access
type PodRepository interface {
	CreateAuditLog(ctx context.Context, entry *AuditLogEntry) error
	GetAuditLogsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]AuditLogEntry, error)
	GetAuditLogsByResource(ctx context.Context, resource, namespace string, limit, offset int) ([]AuditLogEntry, error)
}
