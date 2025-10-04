package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/app/prometheus"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// MonitoringService defines the interface for monitoring operations
type MonitoringService interface {
	GetClusterMetrics(ctx context.Context, userID, clusterID uuid.UUID, req domain.MonitoringRequest) (*domain.ClusterMetrics, error)
	GetApplicationMetrics(ctx context.Context, userID, appID uuid.UUID, req domain.MonitoringRequest) (*domain.ApplicationMetrics, error)
	GetPodMetrics(ctx context.Context, userID, podID uuid.UUID, req domain.MonitoringRequest) (*domain.PodMetrics, error)
	GetAlerts(ctx context.Context, userID, clusterID uuid.UUID) ([]domain.Alert, error)
}

// monitoringService implements MonitoringService
type monitoringService struct {
	appRepo       repo.ApplicationRepository
	clusterRepo   repo.ClusterRepository
	orgRepo       repo.OrganizationRepository
	cryptoService crypto.CryptoService
	prometheusClient prometheus.PrometheusClientInterface
	logger        *zap.Logger
	cache         *monitoringCache
	rateLimiter   *rateLimiter
}

// monitoringCache provides in-memory caching for monitoring data
type monitoringCache struct {
	mu    sync.RWMutex
	items map[string]*domain.MonitoringCacheEntry
}

// rateLimiter provides rate limiting for monitoring requests
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(
	appRepo repo.ApplicationRepository,
	clusterRepo repo.ClusterRepository,
	orgRepo repo.OrganizationRepository,
	cryptoService crypto.CryptoService,
	prometheusClient prometheus.PrometheusClientInterface,
	logger *zap.Logger,
) MonitoringService {
	return &monitoringService{
		appRepo:          appRepo,
		clusterRepo:      clusterRepo,
		orgRepo:          orgRepo,
		cryptoService:    cryptoService,
		prometheusClient: prometheusClient,
		logger:           logger,
		cache:            newMonitoringCache(),
		rateLimiter:      newRateLimiter(100, time.Minute), // 100 requests per minute
	}
}

// GetClusterMetrics retrieves cluster-level metrics
func (s *monitoringService) GetClusterMetrics(ctx context.Context, userID, clusterID uuid.UUID, req domain.MonitoringRequest) (*domain.ClusterMetrics, error) {
	// Check rate limit
	if !s.rateLimiter.Allow(userID.String()) {
		return nil, errors.New("rate limit exceeded")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("cluster_metrics_%s_%s", clusterID.String(), req.TimeRange)
	if cached := s.cache.Get(cacheKey); cached != nil {
		if metrics, ok := cached.(*domain.ClusterMetrics); ok {
			s.logger.Debug("Returning cached cluster metrics", zap.String("clusterID", clusterID.String()))
			return metrics, nil
		}
	}

	// Get cluster details
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		s.logger.Error("Failed to get cluster", zap.Error(err), zap.String("clusterID", clusterID.String()))
		return nil, fmt.Errorf("cluster not found: %w", err)
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", cluster.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Set default time range if not provided
	if req.TimeRange == "" {
		req.TimeRange = domain.TimeRange5m
	}

	// Calculate time range
	endTime := time.Now()
	if req.EndTime != nil {
		endTime = *req.EndTime
	}
	startTime := endTime.Add(-domain.GetTimeRangeDuration(req.TimeRange))
	if req.StartTime != nil {
		startTime = *req.StartTime
	}

	// Query Prometheus for cluster metrics
	metrics := &domain.ClusterMetrics{
		ClusterID: clusterID,
		TimeRange: req.TimeRange,
		Timestamp: time.Now(),
	}

	// Get CPU usage
	cpuQuery := prometheus.ClusterCPUUsageQuery
	cpuResp, err := s.prometheusClient.QueryRange(ctx, cpuQuery, startTime, endTime, "1m")
	if err != nil {
		s.logger.Error("Failed to query CPU usage", zap.Error(err))
		metrics.CPUUsage = domain.MetricSeries{MetricName: "cpu_usage", DataPoints: []domain.MetricDataPoint{}}
	} else {
		metrics.CPUUsage = s.convertPrometheusResponse(cpuResp, "cpu_usage")
	}

	// Get memory usage
	memQuery := prometheus.ClusterMemoryUsageQuery
	memResp, err := s.prometheusClient.QueryRange(ctx, memQuery, startTime, endTime, "1m")
	if err != nil {
		s.logger.Error("Failed to query memory usage", zap.Error(err))
		metrics.MemoryUsage = domain.MetricSeries{MetricName: "memory_usage", DataPoints: []domain.MetricDataPoint{}}
	} else {
		metrics.MemoryUsage = s.convertPrometheusResponse(memResp, "memory_usage")
	}

	// Get node count
	nodeCountResp, err := s.prometheusClient.Query(ctx, prometheus.ClusterNodeCountQuery, endTime)
	if err != nil {
		s.logger.Error("Failed to query node count", zap.Error(err))
		metrics.NodeCount = 0
	} else {
		metrics.NodeCount = s.extractSingleValue(nodeCountResp)
	}

	// Get healthy nodes count
	healthyNodesResp, err := s.prometheusClient.Query(ctx, prometheus.ClusterHealthyNodesQuery, endTime)
	if err != nil {
		s.logger.Error("Failed to query healthy nodes", zap.Error(err))
		metrics.HealthyNodes = 0
	} else {
		metrics.HealthyNodes = s.extractSingleValue(healthyNodesResp)
	}

	metrics.UnhealthyNodes = metrics.NodeCount - metrics.HealthyNodes

	// Cache the result
	s.cache.Set(cacheKey, metrics, 30*time.Second)

	return metrics, nil
}

// GetApplicationMetrics retrieves application-level metrics
func (s *monitoringService) GetApplicationMetrics(ctx context.Context, userID, appID uuid.UUID, req domain.MonitoringRequest) (*domain.ApplicationMetrics, error) {
	// Check rate limit
	if !s.rateLimiter.Allow(userID.String()) {
		return nil, errors.New("rate limit exceeded")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("app_metrics_%s_%s", appID.String(), req.TimeRange)
	if cached := s.cache.Get(cacheKey); cached != nil {
		if metrics, ok := cached.(*domain.ApplicationMetrics); ok {
			s.logger.Debug("Returning cached application metrics", zap.String("appID", appID.String()))
			return metrics, nil
		}
	}

	// Get application details
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", appID.String()))
		return nil, fmt.Errorf("application not found: %w", err)
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Set default time range if not provided
	if req.TimeRange == "" {
		req.TimeRange = domain.TimeRange5m
	}

	// Calculate time range
	endTime := time.Now()
	if req.EndTime != nil {
		endTime = *req.EndTime
	}
	startTime := endTime.Add(-domain.GetTimeRangeDuration(req.TimeRange))
	if req.StartTime != nil {
		startTime = *req.StartTime
	}

	// Query Prometheus for application metrics
	metrics := &domain.ApplicationMetrics{
		AppID:     appID,
		ClusterID: app.ClusterID,
		TimeRange: req.TimeRange,
		Timestamp: time.Now(),
	}

	// Get CPU usage for the application namespace
	cpuQuery := prometheus.BuildPromQLQuery(prometheus.AppCPUUsageQuery, app.Name)
	cpuResp, err := s.prometheusClient.QueryRange(ctx, cpuQuery, startTime, endTime, "1m")
	if err != nil {
		s.logger.Error("Failed to query app CPU usage", zap.Error(err))
		metrics.CPUUsage = domain.MetricSeries{MetricName: "cpu_usage", DataPoints: []domain.MetricDataPoint{}}
	} else {
		metrics.CPUUsage = s.convertPrometheusResponse(cpuResp, "cpu_usage")
	}

	// Get memory usage for the application namespace
	memQuery := prometheus.BuildPromQLQuery(prometheus.AppMemoryUsageQuery, app.Name)
	memResp, err := s.prometheusClient.QueryRange(ctx, memQuery, startTime, endTime, "1m")
	if err != nil {
		s.logger.Error("Failed to query app memory usage", zap.Error(err))
		metrics.MemoryUsage = domain.MetricSeries{MetricName: "memory_usage", DataPoints: []domain.MetricDataPoint{}}
	} else {
		metrics.MemoryUsage = s.convertPrometheusResponse(memResp, "memory_usage")
	}

	// Get pod counts
	podCountQuery := prometheus.BuildPromQLQuery(prometheus.AppPodCountQuery, app.Name)
	podCountResp, err := s.prometheusClient.Query(ctx, podCountQuery, endTime)
	if err != nil {
		s.logger.Error("Failed to query pod count", zap.Error(err))
		metrics.PodCount = 0
	} else {
		metrics.PodCount = s.extractSingleValue(podCountResp)
	}

	runningPodsQuery := prometheus.BuildPromQLQuery(prometheus.AppRunningPodsQuery, app.Name)
	runningPodsResp, err := s.prometheusClient.Query(ctx, runningPodsQuery, endTime)
	if err != nil {
		s.logger.Error("Failed to query running pods", zap.Error(err))
		metrics.RunningPods = 0
	} else {
		metrics.RunningPods = s.extractSingleValue(runningPodsResp)
	}

	pendingPodsQuery := prometheus.BuildPromQLQuery(prometheus.AppPendingPodsQuery, app.Name)
	pendingPodsResp, err := s.prometheusClient.Query(ctx, pendingPodsQuery, endTime)
	if err != nil {
		s.logger.Error("Failed to query pending pods", zap.Error(err))
		metrics.PendingPods = 0
	} else {
		metrics.PendingPods = s.extractSingleValue(pendingPodsResp)
	}

	failedPodsQuery := prometheus.BuildPromQLQuery(prometheus.AppFailedPodsQuery, app.Name)
	failedPodsResp, err := s.prometheusClient.Query(ctx, failedPodsQuery, endTime)
	if err != nil {
		s.logger.Error("Failed to query failed pods", zap.Error(err))
		metrics.FailedPods = 0
	} else {
		metrics.FailedPods = s.extractSingleValue(failedPodsResp)
	}

	// Get top 5 alerts for the application
	alerts, err := s.GetAlerts(ctx, userID, app.ClusterID)
	if err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		metrics.TopAlerts = []domain.Alert{}
	} else {
		// Filter alerts for this application and limit to top 5
		var appAlerts []domain.Alert
		for _, alert := range alerts {
			if alert.Labels["namespace"] == app.Name {
				appAlerts = append(appAlerts, alert)
				if len(appAlerts) >= 5 {
					break
				}
			}
		}
		metrics.TopAlerts = appAlerts
	}

	// Cache the result
	s.cache.Set(cacheKey, metrics, 30*time.Second)

	return metrics, nil
}

// GetPodMetrics retrieves pod-level metrics
func (s *monitoringService) GetPodMetrics(ctx context.Context, userID, podID uuid.UUID, req domain.MonitoringRequest) (*domain.PodMetrics, error) {
	// Check rate limit
	if !s.rateLimiter.Allow(userID.String()) {
		return nil, errors.New("rate limit exceeded")
	}

	// For pod metrics, we need to find the pod first
	// This is a simplified implementation - in production you'd want a pod repository
	return nil, errors.New("pod metrics not fully implemented - requires pod repository")
}

// GetAlerts retrieves alerts for a cluster
func (s *monitoringService) GetAlerts(ctx context.Context, userID, clusterID uuid.UUID) ([]domain.Alert, error) {
	// Check rate limit
	if !s.rateLimiter.Allow(userID.String()) {
		return nil, errors.New("rate limit exceeded")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("alerts_%s", clusterID.String())
	if cached := s.cache.Get(cacheKey); cached != nil {
		if alerts, ok := cached.([]domain.Alert); ok {
			s.logger.Debug("Returning cached alerts", zap.String("clusterID", clusterID.String()))
			return alerts, nil
		}
	}

	// Get cluster details for authorization
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		s.logger.Error("Failed to get cluster", zap.Error(err), zap.String("clusterID", clusterID.String()))
		return nil, fmt.Errorf("cluster not found: %w", err)
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", cluster.OrgID.String()), zap.String("userID", userID.String()))
		return nil, fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return nil, errors.New("unauthorized: user is not a member of the organization")
	}

	// Get alerts from Prometheus
	alerts, err := s.prometheusClient.GetAlerts(ctx)
	if err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	// Cache the result
	s.cache.Set(cacheKey, alerts, 1*time.Minute)

	return alerts, nil
}

// convertPrometheusResponse converts Prometheus response to domain MetricSeries
func (s *monitoringService) convertPrometheusResponse(resp *domain.PrometheusResponse, metricName string) domain.MetricSeries {
	series := domain.MetricSeries{
		MetricName: metricName,
		Labels:     make(map[string]string),
		DataPoints: []domain.MetricDataPoint{},
	}

	if len(resp.Data.Result) == 0 {
		return series
	}

	// Use the first result
	result := resp.Data.Result[0]
	series.Labels = result.Metric

	// Convert values to data points
	for _, value := range result.Values {
		if len(value) >= 2 {
			if timestamp, ok := value[0].(float64); ok {
				if val, ok := value[1].(string); ok {
					if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
						series.DataPoints = append(series.DataPoints, domain.MetricDataPoint{
							Timestamp: time.Unix(int64(timestamp), 0),
							Value:     floatVal,
						})
					}
				}
			}
		}
	}

	return series
}

// extractSingleValue extracts a single numeric value from Prometheus response
func (s *monitoringService) extractSingleValue(resp *domain.PrometheusResponse) int {
	if len(resp.Data.Result) == 0 {
		return 0
	}

	result := resp.Data.Result[0]
	if len(result.Values) == 0 {
		return 0
	}

	value := result.Values[0]
	if len(value) >= 2 {
		if val, ok := value[1].(string); ok {
			if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				return int(floatVal)
			}
		}
	}

	return 0
}

// newMonitoringCache creates a new monitoring cache
func newMonitoringCache() *monitoringCache {
	return &monitoringCache{
		items: make(map[string]*domain.MonitoringCacheEntry),
	}
}

// Get retrieves an item from the cache
func (c *monitoringCache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil
	}

	if time.Since(entry.Timestamp) > entry.TTL {
		delete(c.items, key)
		return nil
	}

	return entry.Data
}

// Set stores an item in the cache
func (c *monitoringCache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &domain.MonitoringCacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed under the rate limit
func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Clean old requests
	if requests, exists := r.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		r.requests[key] = validRequests
	}

	// Check if under limit
	if len(r.requests[key]) < r.limit {
		r.requests[key] = append(r.requests[key], now)
		return true
	}

	return false
}