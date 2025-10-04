package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/prometheus"
	"github.com/PouryDev/oneclick/internal/domain"
)

// MockPrometheusClient is a mock implementation of PrometheusClientInterface
type MockPrometheusClient struct {
	mock.Mock
}

func (m *MockPrometheusClient) Query(ctx context.Context, query string, queryTime time.Time) (*domain.PrometheusResponse, error) {
	args := m.Called(ctx, query, queryTime)
	return args.Get(0).(*domain.PrometheusResponse), args.Error(1)
}

func (m *MockPrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step string) (*domain.PrometheusResponse, error) {
	args := m.Called(ctx, query, start, end, step)
	return args.Get(0).(*domain.PrometheusResponse), args.Error(1)
}

func (m *MockPrometheusClient) GetAlerts(ctx context.Context) ([]domain.Alert, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockPrometheusClient) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestMonitoringService_GetClusterMetrics(t *testing.T) {
	logger := zap.NewNop()
	mockAppRepo := &MockApplicationRepository{}
	mockClusterRepo := &MockClusterRepository{}
	mockOrgRepo := &MockOrganizationRepository{}
	mockCryptoService := &MockCryptoService{}
	mockPrometheusClient := &MockPrometheusClient{}

	service := NewMonitoringService(mockAppRepo, mockClusterRepo, mockOrgRepo, mockCryptoService, mockPrometheusClient, logger)

	userID := uuid.New()
	clusterID := uuid.New()
	orgID := uuid.New()

	// Mock cluster data
	cluster := &domain.Cluster{
		ID:     clusterID,
		OrgID:  orgID,
		Name:   "test-cluster",
		Status: "active",
	}

	// Mock Prometheus responses
	cpuResponse := &domain.PrometheusResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		}{
			ResultType: "matrix",
			Result: []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			}{
				{
					Metric: map[string]string{"__name__": "cpu_usage"},
					Values: [][]interface{}{
						{float64(1640995200), "0.5"},
						{float64(1640995260), "0.6"},
					},
				},
			},
		},
	}

	memResponse := &domain.PrometheusResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		}{
			ResultType: "matrix",
			Result: []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			}{
				{
					Metric: map[string]string{"__name__": "memory_usage"},
					Values: [][]interface{}{
						{float64(1640995200), "1073741824"},
						{float64(1640995260), "1073741824"},
					},
				},
			},
		},
	}

	nodeCountResponse := &domain.PrometheusResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		}{
			ResultType: "vector",
			Result: []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			}{
				{
					Metric: map[string]string{"__name__": "node_count"},
					Values: [][]interface{}{
						{float64(1640995200), "3"},
					},
				},
			},
		},
	}

	healthyNodesResponse := &domain.PrometheusResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		}{
			ResultType: "vector",
			Result: []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			}{
				{
					Metric: map[string]string{"__name__": "healthy_nodes"},
					Values: [][]interface{}{
						{float64(1640995200), "3"},
					},
				},
			},
		},
	}

	// Set up mocks
	mockClusterRepo.On("GetClusterByID", mock.Anything, clusterID).Return(cluster, nil)
	mockOrgRepo.On("GetUserRoleInOrganization", mock.Anything, userID, orgID).Return("admin", nil)
	mockPrometheusClient.On("QueryRange", mock.Anything, prometheus.ClusterCPUUsageQuery, mock.Anything, mock.Anything, "1m").Return(cpuResponse, nil)
	mockPrometheusClient.On("QueryRange", mock.Anything, prometheus.ClusterMemoryUsageQuery, mock.Anything, mock.Anything, "1m").Return(memResponse, nil)
	mockPrometheusClient.On("Query", mock.Anything, prometheus.ClusterNodeCountQuery, mock.Anything).Return(nodeCountResponse, nil)
	mockPrometheusClient.On("Query", mock.Anything, prometheus.ClusterHealthyNodesQuery, mock.Anything).Return(healthyNodesResponse, nil)

	// Test
	req := domain.MonitoringRequest{
		TimeRange: domain.TimeRange5m,
	}

	metrics, err := service.GetClusterMetrics(context.Background(), userID, clusterID, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, clusterID, metrics.ClusterID)
	assert.Equal(t, domain.TimeRange5m, metrics.TimeRange)
	assert.Equal(t, 3, metrics.NodeCount)
	assert.Equal(t, 3, metrics.HealthyNodes)
	assert.Equal(t, 0, metrics.UnhealthyNodes)
	assert.Len(t, metrics.CPUUsage.DataPoints, 2)
	assert.Len(t, metrics.MemoryUsage.DataPoints, 2)

	// Verify all mocks were called
	mockClusterRepo.AssertExpectations(t)
	mockOrgRepo.AssertExpectations(t)
	mockPrometheusClient.AssertExpectations(t)
}

func TestMonitoringService_RateLimit(t *testing.T) {
	// Test the rate limiter directly
	rateLimiter := newRateLimiter(5, time.Minute) // 5 requests per minute for testing

	userID := "test-user"

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		assert.True(t, rateLimiter.Allow(userID), "Request %d should be allowed", i+1)
	}

	// 6th request should be rate limited
	assert.False(t, rateLimiter.Allow(userID), "6th request should be rate limited")

	// Test with different user - should be allowed
	assert.True(t, rateLimiter.Allow("different-user"), "Different user should be allowed")
}
