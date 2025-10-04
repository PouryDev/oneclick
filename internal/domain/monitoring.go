package domain

import (
	"time"

	"github.com/google/uuid"
)

// MetricType represents the type of metric being queried
type MetricType string

const (
	MetricTypeCPU    MetricType = "cpu"
	MetricTypeMemory MetricType = "memory"
	MetricTypeDisk   MetricType = "disk"
	MetricTypeNetwork MetricType = "network"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// TimeRange represents the time range for metrics queries
type TimeRange string

const (
	TimeRange5m  TimeRange = "5m"
	TimeRange15m TimeRange = "15m"
	TimeRange1h  TimeRange = "1h"
	TimeRange6h  TimeRange = "6h"
	TimeRange24h TimeRange = "24h"
)

// MetricDataPoint represents a single data point in a time series
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricSeries represents a time series of metric data
type MetricSeries struct {
	MetricName string             `json:"metric_name"`
	Labels     map[string]string  `json:"labels"`
	DataPoints []MetricDataPoint  `json:"data_points"`
}

// ClusterMetrics represents aggregated cluster-level metrics
type ClusterMetrics struct {
	ClusterID     uuid.UUID      `json:"cluster_id"`
	TimeRange     TimeRange      `json:"time_range"`
	CPUUsage      MetricSeries   `json:"cpu_usage"`
	MemoryUsage   MetricSeries   `json:"memory_usage"`
	NodeCount     int            `json:"node_count"`
	HealthyNodes  int            `json:"healthy_nodes"`
	UnhealthyNodes int           `json:"unhealthy_nodes"`
	Timestamp     time.Time      `json:"timestamp"`
}

// ApplicationMetrics represents application-level metrics
type ApplicationMetrics struct {
	AppID         uuid.UUID      `json:"app_id"`
	ClusterID     uuid.UUID      `json:"cluster_id"`
	TimeRange     TimeRange      `json:"time_range"`
	CPUUsage      MetricSeries   `json:"cpu_usage"`
	MemoryUsage   MetricSeries   `json:"memory_usage"`
	PodCount      int            `json:"pod_count"`
	RunningPods   int            `json:"running_pods"`
	PendingPods   int            `json:"pending_pods"`
	FailedPods    int            `json:"failed_pods"`
	TopAlerts     []Alert        `json:"top_alerts"`
	Timestamp     time.Time      `json:"timestamp"`
}

// PodMetrics represents pod-level metrics
type PodMetrics struct {
	PodID       uuid.UUID      `json:"pod_id"`
	PodName     string         `json:"pod_name"`
	Namespace   string         `json:"namespace"`
	ClusterID   uuid.UUID      `json:"cluster_id"`
	CPUUsage    MetricSeries   `json:"cpu_usage"`
	MemoryUsage MetricSeries   `json:"memory_usage"`
	Status      string         `json:"status"`
	Restarts    int32          `json:"restarts"`
	Timestamp   time.Time      `json:"timestamp"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Severity    AlertSeverity `json:"severity"`
	Status      string        `json:"status"` // "firing", "resolved"
	Labels      map[string]string `json:"labels"`
	StartsAt    time.Time     `json:"starts_at"`
	EndsAt      *time.Time    `json:"ends_at,omitempty"`
}

// PrometheusQuery represents a PromQL query
type PrometheusQuery struct {
	Query     string    `json:"query"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Step      string    `json:"step"`
}

// PrometheusResponse represents the response from Prometheus API
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}  `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// MonitoringRequest represents a request for monitoring data
type MonitoringRequest struct {
	TimeRange TimeRange `form:"range" binding:"omitempty,oneof=5m 15m 1h 6h 24h"`
	StartTime *time.Time `form:"start"`
	EndTime   *time.Time `form:"end"`
	Step      string    `form:"step"`
}

// MonitoringCacheEntry represents a cached monitoring result
type MonitoringCacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// IsValidMetricType checks if the metric type is valid
func IsValidMetricType(t string) bool {
	switch t {
	case string(MetricTypeCPU), string(MetricTypeMemory), string(MetricTypeDisk), string(MetricTypeNetwork):
		return true
	default:
		return false
	}
}

// IsValidAlertSeverity checks if the alert severity is valid
func IsValidAlertSeverity(s string) bool {
	switch s {
	case string(AlertSeverityInfo), string(AlertSeverityWarning), string(AlertSeverityCritical):
		return true
	default:
		return false
	}
}

// IsValidTimeRange checks if the time range is valid
func IsValidTimeRange(r string) bool {
	switch r {
	case string(TimeRange5m), string(TimeRange15m), string(TimeRange1h), string(TimeRange6h), string(TimeRange24h):
		return true
	default:
		return false
	}
}

// GetTimeRangeDuration returns the duration for a time range
func GetTimeRangeDuration(tr TimeRange) time.Duration {
	switch tr {
	case TimeRange5m:
		return 5 * time.Minute
	case TimeRange15m:
		return 15 * time.Minute
	case TimeRange1h:
		return 1 * time.Hour
	case TimeRange6h:
		return 6 * time.Hour
	case TimeRange24h:
		return 24 * time.Hour
	default:
		return 5 * time.Minute
	}
}