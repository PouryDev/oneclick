package domain

import (
	"time"

	"github.com/google/uuid"
)

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStatusPending      ServiceStatus = "pending"
	ServiceStatusProvisioning ServiceStatus = "provisioning"
	ServiceStatusRunning      ServiceStatus = "running"
	ServiceStatusFailed       ServiceStatus = "failed"
	ServiceStatusStopped      ServiceStatus = "stopped"
)

// Service represents a provisioned service
type Service struct {
	ID        uuid.UUID     `json:"id"`
	AppID     uuid.UUID     `json:"app_id"`
	Name      string        `json:"name"`
	Chart     string        `json:"chart"`
	Status    ServiceStatus `json:"status"`
	Namespace string        `json:"namespace"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ServiceConfig represents a configuration key-value pair for a service
type ServiceConfig struct {
	ID        uuid.UUID `json:"id"`
	ServiceID uuid.UUID `json:"service_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	IsSecret  bool      `json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServiceSummary represents a service in list views
type ServiceSummary struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Chart     string        `json:"chart"`
	Status    ServiceStatus `json:"status"`
	Namespace string        `json:"namespace"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ServiceDetail represents detailed service information
type ServiceDetail struct {
	Service
	Configs []ServiceConfigSummary `json:"configs"`
}

// ServiceConfigSummary represents a service config in list views
type ServiceConfigSummary struct {
	ID       uuid.UUID `json:"id"`
	Key      string    `json:"key"`
	Value    string    `json:"value"`
	IsSecret bool      `json:"is_secret"`
}

// ServiceConfigResponse represents a service config response (masks secrets)
type ServiceConfigResponse struct {
	ID       uuid.UUID `json:"id"`
	Key      string    `json:"key"`
	Value    string    `json:"value"`
	IsSecret bool      `json:"is_secret"`
}

// Infrastructure Configuration Models

// InfraConfig represents the parsed infra-config.yml structure
type InfraConfig struct {
	Services map[string]ServiceDefinition `yaml:"services"`
	App      AppDefinition                `yaml:"app"`
}

// ServiceDefinition represents a service definition in infra-config.yml
type ServiceDefinition struct {
	Chart string            `yaml:"chart"`
	Env   map[string]string `yaml:"env"`
}

// AppDefinition represents the app configuration in infra-config.yml
type AppDefinition struct {
	Env map[string]string `yaml:"env"`
}

// SecretReference represents a secret reference in configuration
type SecretReference struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ProvisionServiceRequest represents a request to provision services
type ProvisionServiceRequest struct {
	InfraConfig string `json:"infra_config"` // YAML content
}

// ProvisionServiceResponse represents a response to service provisioning
type ProvisionServiceResponse struct {
	Services []ServiceSummary `json:"services"`
	Message  string           `json:"message"`
}

// ServiceConfigRevealRequest represents a request to reveal service config
type ServiceConfigRevealRequest struct {
	ConfigID uuid.UUID `json:"config_id"`
}

// ServiceConfigRevealResponse represents a response with revealed config
type ServiceConfigRevealResponse struct {
	Config ServiceConfigResponse `json:"config"`
}

// ValidServiceStatuses returns a list of valid service statuses
func ValidServiceStatuses() []ServiceStatus {
	return []ServiceStatus{
		ServiceStatusPending,
		ServiceStatusProvisioning,
		ServiceStatusRunning,
		ServiceStatusFailed,
		ServiceStatusStopped,
	}
}

// IsValidServiceStatus checks if a status is valid
func IsValidServiceStatus(status ServiceStatus) bool {
	for _, validStatus := range ValidServiceStatuses() {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ToSummary converts a Service to ServiceSummary
func (s *Service) ToSummary() ServiceSummary {
	return ServiceSummary{
		ID:        s.ID,
		Name:      s.Name,
		Chart:     s.Chart,
		Status:    s.Status,
		Namespace: s.Namespace,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// ToSummary converts a ServiceConfig to ServiceConfigSummary
func (sc *ServiceConfig) ToSummary() ServiceConfigSummary {
	return ServiceConfigSummary{
		ID:       sc.ID,
		Key:      sc.Key,
		Value:    sc.Value,
		IsSecret: sc.IsSecret,
	}
}

// ToResponse converts a ServiceConfig to ServiceConfigResponse (masks secrets)
func (sc *ServiceConfig) ToResponse() ServiceConfigResponse {
	value := sc.Value
	if sc.IsSecret {
		value = "***MASKED***"
	}

	return ServiceConfigResponse{
		ID:       sc.ID,
		Key:      sc.Key,
		Value:    value,
		IsSecret: sc.IsSecret,
	}
}

// ToDetail converts a Service to ServiceDetail
func (s *Service) ToDetail(configs []ServiceConfig) ServiceDetail {
	var configSummaries []ServiceConfigSummary
	for _, config := range configs {
		configSummaries = append(configSummaries, config.ToSummary())
	}

	return ServiceDetail{
		Service: *s,
		Configs: configSummaries,
	}
}

// IsRunning returns true if the service is currently running
func (s *Service) IsRunning() bool {
	return s.Status == ServiceStatusRunning
}

// IsProvisioning returns true if the service is currently provisioning
func (s *Service) IsProvisioning() bool {
	return s.Status == ServiceStatusProvisioning
}

// IsFailed returns true if the service has failed
func (s *Service) IsFailed() bool {
	return s.Status == ServiceStatusFailed
}

// IsStopped returns true if the service is stopped
func (s *Service) IsStopped() bool {
	return s.Status == ServiceStatusStopped
}
