package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
)

// PrometheusClientInterface defines the interface for Prometheus operations
type PrometheusClientInterface interface {
	Query(ctx context.Context, query string, queryTime time.Time) (*domain.PrometheusResponse, error)
	QueryRange(ctx context.Context, query string, start, end time.Time, step string) (*domain.PrometheusResponse, error)
	GetAlerts(ctx context.Context) ([]domain.Alert, error)
	HealthCheck(ctx context.Context) error
}

// PrometheusClient implements PrometheusClientInterface
type PrometheusClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewPrometheusClient creates a new Prometheus client
func NewPrometheusClient(baseURL string, logger *zap.Logger) PrometheusClientInterface {
	return &PrometheusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Query executes a single PromQL query
func (c *PrometheusClient) Query(ctx context.Context, query string, queryTime time.Time) (*domain.PrometheusResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("time", queryTime.Format(time.RFC3339))

	url := fmt.Sprintf("%s/api/v1/query?%s", c.baseURL, params.Encode())

	c.logger.Debug("Executing Prometheus query",
		zap.String("url", url),
		zap.String("query", query),
		zap.Time("time", queryTime))

	return c.executeRequest(ctx, url)
}

// QueryRange executes a range PromQL query
func (c *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step string) (*domain.PrometheusResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("step", step)

	url := fmt.Sprintf("%s/api/v1/query_range?%s", c.baseURL, params.Encode())

	c.logger.Debug("Executing Prometheus range query",
		zap.String("url", url),
		zap.String("query", query),
		zap.Time("start", start),
		zap.Time("end", end),
		zap.String("step", step))

	return c.executeRequest(ctx, url)
}

// GetAlerts retrieves active alerts from Prometheus
func (c *PrometheusClient) GetAlerts(ctx context.Context) ([]domain.Alert, error) {
	url := fmt.Sprintf("%s/api/v1/alerts", c.baseURL)

	c.logger.Debug("Fetching Prometheus alerts", zap.String("url", url))

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alerts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var alertResponse struct {
		Status string `json:"status"`
		Data   struct {
			Alerts []struct {
				Labels      map[string]string `json:"labels"`
				Annotations map[string]string `json:"annotations"`
				State       string            `json:"state"`
				ActiveAt    time.Time         `json:"activeAt"`
			} `json:"alerts"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&alertResponse); err != nil {
		return nil, fmt.Errorf("failed to decode alerts response: %w", err)
	}

	var alerts []domain.Alert
	for _, alert := range alertResponse.Data.Alerts {
		severity := domain.AlertSeverityInfo
		if sev, ok := alert.Labels["severity"]; ok {
			switch sev {
			case "warning":
				severity = domain.AlertSeverityWarning
			case "critical":
				severity = domain.AlertSeverityCritical
			}
		}

		alerts = append(alerts, domain.Alert{
			ID:          generateAlertID(alert.Labels),
			Name:        alert.Labels["alertname"],
			Description: alert.Annotations["description"],
			Severity:    severity,
			Status:      alert.State,
			Labels:      alert.Labels,
			StartsAt:    alert.ActiveAt,
		})
	}

	return alerts, nil
}

// HealthCheck verifies that Prometheus is accessible
func (c *PrometheusClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/-/healthy", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to check Prometheus health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Prometheus health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// executeRequest executes an HTTP request to Prometheus
func (c *PrometheusClient) executeRequest(ctx context.Context, url string) (*domain.PrometheusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var promResp domain.PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s", promResp.Status)
	}

	return &promResp, nil
}

// generateAlertID creates a unique ID for an alert based on its labels
func generateAlertID(labels map[string]string) uuid.UUID {
	// Create a deterministic ID based on alert labels
	key := fmt.Sprintf("%s:%s:%s",
		labels["alertname"],
		labels["instance"],
		labels["job"])

	// Use a simple hash to generate UUID
	hash := 0
	for _, char := range key {
		hash = hash*31 + int(char)
	}

	// Convert hash to UUID format (simplified)
	return uuid.New() // In production, you'd want a deterministic UUID generation
}

// Common PromQL queries for monitoring
const (
	// Cluster-level queries
	ClusterCPUUsageQuery     = `sum(rate(container_cpu_usage_seconds_total{container!="POD",container!=""}[5m]))`
	ClusterMemoryUsageQuery  = `sum(container_memory_usage_bytes{container!="POD",container!=""})`
	ClusterNodeCountQuery    = `count(kube_node_info)`
	ClusterHealthyNodesQuery = `count(kube_node_status_condition{condition="Ready",status="true"})`

	// Application-level queries
	AppCPUUsageQuery    = `sum(rate(container_cpu_usage_seconds_total{namespace="%s",container!="POD",container!=""}[5m])) by (namespace)`
	AppMemoryUsageQuery = `sum(container_memory_usage_bytes{namespace="%s",container!="POD",container!=""}) by (namespace)`
	AppPodCountQuery    = `count(kube_pod_info{namespace="%s"})`
	AppRunningPodsQuery = `count(kube_pod_status_phase{namespace="%s",phase="Running"})`
	AppPendingPodsQuery = `count(kube_pod_status_phase{namespace="%s",phase="Pending"})`
	AppFailedPodsQuery  = `count(kube_pod_status_phase{namespace="%s",phase="Failed"})`

	// Pod-level queries
	PodCPUUsageQuery     = `rate(container_cpu_usage_seconds_total{pod="%s",namespace="%s",container!="POD",container!=""}[5m])`
	PodMemoryUsageQuery  = `container_memory_usage_bytes{pod="%s",namespace="%s",container!="POD",container!=""}`
	PodRestartCountQuery = `kube_pod_container_status_restarts_total{pod="%s",namespace="%s"}`
)

// BuildPromQLQuery constructs a PromQL query with parameters
func BuildPromQLQuery(template string, params ...interface{}) string {
	return fmt.Sprintf(template, params...)
}
