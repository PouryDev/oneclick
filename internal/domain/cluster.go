package domain

import (
	"time"

	"github.com/google/uuid"
)

// Cluster represents a Kubernetes cluster
type Cluster struct {
	ID                   uuid.UUID  `json:"id"`
	OrgID                uuid.UUID  `json:"org_id"`
	Name                 string     `json:"name"`
	Provider             string     `json:"provider"`
	Region               string     `json:"region"`
	KubeconfigEncrypted  []byte     `json:"-"` // Never expose in JSON
	NodeCount            int        `json:"node_count"`
	Status               string     `json:"status"`
	KubeVersion          *string    `json:"kube_version,omitempty"`
	LastHealthCheck      *time.Time `json:"last_health_check,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ClusterSummary represents a cluster in list views
type ClusterSummary struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Provider    string     `json:"provider"`
	Region      string     `json:"region"`
	NodeCount   int        `json:"node_count"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ClusterDetail represents detailed cluster information
type ClusterDetail struct {
	Cluster
	HasKubeconfig bool `json:"has_kubeconfig"`
}

// ClusterHealth represents cluster health status
type ClusterHealth struct {
	Status      string      `json:"status"`
	KubeVersion string      `json:"kube_version"`
	Nodes       []NodeInfo  `json:"nodes"`
	LastCheck   time.Time   `json:"last_check"`
}

// NodeInfo represents a Kubernetes node
type NodeInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// Request/Response DTOs
type CreateClusterRequest struct {
	Name      string `json:"name" validate:"required,min=1"`
	Provider  string `json:"provider" validate:"required,min=1"`
	Region    string `json:"region" validate:"required,min=1"`
	Kubeconfig *string `json:"kubeconfig,omitempty"` // Base64 encoded kubeconfig
}

type ImportClusterRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

type ClusterResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Provider    string     `json:"provider"`
	Region      string     `json:"region"`
	NodeCount   int        `json:"node_count"`
	Status      string     `json:"status"`
	KubeVersion *string    `json:"kube_version,omitempty"`
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ClusterDetailResponse struct {
	ClusterResponse
	HasKubeconfig bool `json:"has_kubeconfig"`
}

// Valid cluster statuses
const (
	StatusProvisioning = "provisioning"
	StatusActive       = "active"
	StatusError        = "error"
	StatusDeleting     = "deleting"
)

// ValidStatuses returns a list of valid cluster statuses
func ValidClusterStatuses() []string {
	return []string{StatusProvisioning, StatusActive, StatusError, StatusDeleting}
}

// IsValidClusterStatus checks if a status is valid
func IsValidClusterStatus(status string) bool {
	for _, validStatus := range ValidClusterStatuses() {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ToResponse converts a Cluster to ClusterResponse
func (c *Cluster) ToResponse() ClusterResponse {
	return ClusterResponse{
		ID:              c.ID,
		Name:            c.Name,
		Provider:        c.Provider,
		Region:          c.Region,
		NodeCount:       c.NodeCount,
		Status:          c.Status,
		KubeVersion:     c.KubeVersion,
		LastHealthCheck: c.LastHealthCheck,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

// ToSummary converts a Cluster to ClusterSummary
func (c *Cluster) ToSummary() ClusterSummary {
	return ClusterSummary{
		ID:        c.ID,
		Name:      c.Name,
		Provider:  c.Provider,
		Region:    c.Region,
		NodeCount: c.NodeCount,
		Status:    c.Status,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ToDetailResponse converts a Cluster to ClusterDetailResponse
func (c *Cluster) ToDetailResponse() ClusterDetailResponse {
	return ClusterDetailResponse{
		ClusterResponse: c.ToResponse(),
		HasKubeconfig:   len(c.KubeconfigEncrypted) > 0,
	}
}
