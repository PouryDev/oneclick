package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/PouryDev/oneclick/internal/domain"
)

// MockClusterService is a mock implementation of ClusterService
type MockClusterService struct {
	mock.Mock
}

func (m *MockClusterService) CreateCluster(ctx context.Context, userID, orgID uuid.UUID, req *domain.CreateClusterRequest) (*domain.ClusterResponse, error) {
	args := m.Called(ctx, userID, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClusterResponse), args.Error(1)
}

func (m *MockClusterService) ImportCluster(ctx context.Context, userID, orgID uuid.UUID, name string, kubeconfigData []byte) (*domain.ClusterResponse, error) {
	args := m.Called(ctx, userID, orgID, name, kubeconfigData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClusterResponse), args.Error(1)
}

func (m *MockClusterService) GetClustersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.ClusterSummary, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ClusterSummary), args.Error(1)
}

func (m *MockClusterService) GetCluster(ctx context.Context, userID, clusterID uuid.UUID) (*domain.ClusterDetailResponse, error) {
	args := m.Called(ctx, userID, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClusterDetailResponse), args.Error(1)
}

func (m *MockClusterService) GetClusterHealth(ctx context.Context, userID, clusterID uuid.UUID) (*domain.ClusterHealth, error) {
	args := m.Called(ctx, userID, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClusterHealth), args.Error(1)
}

func (m *MockClusterService) DeleteCluster(ctx context.Context, userID, clusterID uuid.UUID) error {
	args := m.Called(ctx, userID, clusterID)
	return args.Error(0)
}

func TestClusterHandler_CreateCluster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		orgID          string
		requestBody    interface{}
		mockSetup      func(*MockClusterService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful cluster creation",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateClusterRequest{
				Name:     "Test Cluster",
				Provider: "aws",
				Region:   "us-west-2",
			},
			mockSetup: func(m *MockClusterService) {
				clusterResponse := &domain.ClusterResponse{
					ID:        uuid.New(),
					Name:      "Test Cluster",
					Provider:  "aws",
					Region:    "us-west-2",
					NodeCount: 0,
					Status:    domain.StatusProvisioning,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("CreateCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateClusterRequest")).Return(clusterResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "cluster creation with kubeconfig",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateClusterRequest{
				Name:       "Test Cluster",
				Provider:   "aws",
				Region:     "us-west-2",
				Kubeconfig: stringPtr("dGVzdC1rdWJlY29uZmln"), // base64 encoded test kubeconfig
			},
			mockSetup: func(m *MockClusterService) {
				clusterResponse := &domain.ClusterResponse{
					ID:        uuid.New(),
					Name:      "Test Cluster",
					Provider:  "aws",
					Region:    "us-west-2",
					NodeCount: 0,
					Status:    domain.StatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("CreateCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateClusterRequest")).Return(clusterResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "access denied",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateClusterRequest{
				Name:     "Test Cluster",
				Provider: "aws",
				Region:   "us-west-2",
			},
			mockSetup: func(m *MockClusterService) {
				m.On("CreateCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateClusterRequest")).Return(nil, errors.New("user does not have access to this organization"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
		{
			name:           "invalid organization ID",
			userID:         uuid.New().String(),
			orgID:          "invalid-uuid",
			requestBody:    domain.CreateClusterRequest{Name: "Test Cluster", Provider: "aws", Region: "us-west-2"},
			mockSetup:      func(m *MockClusterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid organization ID",
		},
		{
			name:           "no user ID in context",
			userID:         "",
			orgID:          uuid.New().String(),
			requestBody:    domain.CreateClusterRequest{Name: "Test Cluster", Provider: "aws", Region: "us-west-2"},
			mockSetup:      func(m *MockClusterService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClusterService)
			tt.mockSetup(mockService)

			handler := NewClusterHandler(mockService)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/orgs/"+tt.orgID+"/clusters", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "orgId", Value: tt.orgID}}

			// Set user ID in context if provided
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.CreateCluster(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestClusterHandler_ImportCluster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		orgID          string
		clusterName    string
		kubeconfigData []byte
		mockSetup      func(*MockClusterService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful cluster import",
			userID:         uuid.New().String(),
			orgID:          uuid.New().String(),
			clusterName:    "Imported Cluster",
			kubeconfigData: []byte("test kubeconfig data"),
			mockSetup: func(m *MockClusterService) {
				clusterResponse := &domain.ClusterResponse{
					ID:        uuid.New(),
					Name:      "Imported Cluster",
					Provider:  "imported",
					Region:    "unknown",
					NodeCount: 0,
					Status:    domain.StatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("ImportCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), "Imported Cluster", []byte("test kubeconfig data")).Return(clusterResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid kubeconfig",
			userID:         uuid.New().String(),
			orgID:          uuid.New().String(),
			clusterName:    "Invalid Cluster",
			kubeconfigData: []byte("invalid kubeconfig"),
			mockSetup: func(m *MockClusterService) {
				m.On("ImportCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), "Invalid Cluster", []byte("invalid kubeconfig")).Return(nil, errors.New("invalid kubeconfig: failed to create REST config"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid kubeconfig",
		},
		{
			name:           "access denied",
			userID:         uuid.New().String(),
			orgID:          uuid.New().String(),
			clusterName:    "Test Cluster",
			kubeconfigData: []byte("test kubeconfig"),
			mockSetup: func(m *MockClusterService) {
				m.On("ImportCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), "Test Cluster", []byte("test kubeconfig")).Return(nil, errors.New("user does not have access to this organization"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClusterService)
			tt.mockSetup(mockService)

			handler := NewClusterHandler(mockService)

			// Create multipart form request
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add cluster name
			writer.WriteField("name", tt.clusterName)

			// Add kubeconfig file
			part, _ := writer.CreateFormFile("kubeconfig", "kubeconfig.yaml")
			part.Write(tt.kubeconfigData)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/orgs/"+tt.orgID+"/clusters/import", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "orgId", Value: tt.orgID}}

			// Set user ID in context if provided
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.ImportCluster(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestClusterHandler_GetClusterHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		clusterID      string
		mockSetup      func(*MockClusterService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful health check",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				health := &domain.ClusterHealth{
					Status:      domain.StatusActive,
					KubeVersion: "v1.25.0",
					Nodes: []domain.NodeInfo{
						{
							Name:   "node-1",
							Status: "Ready",
							CPU:    "2",
							Memory: "4Gi",
						},
					},
					LastCheck: time.Now(),
				}
				m.On("GetClusterHealth", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(health, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "cluster not found",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				m.On("GetClusterHealth", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("cluster not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Cluster not found",
		},
		{
			name:      "cluster does not have kubeconfig",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				m.On("GetClusterHealth", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("cluster does not have kubeconfig"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Cluster does not have kubeconfig",
		},
		{
			name:      "failed to connect to cluster",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				m.On("GetClusterHealth", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("failed to get cluster health: failed to connect to cluster"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Failed to connect to cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClusterService)
			tt.mockSetup(mockService)

			handler := NewClusterHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/clusters/"+tt.clusterID+"/status", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "clusterId", Value: tt.clusterID}}

			// Set user ID in context if provided
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.GetClusterHealth(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestClusterHandler_DeleteCluster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		clusterID      string
		mockSetup      func(*MockClusterService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful cluster deletion",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				m.On("DeleteCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:      "cluster not found",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				m.On("DeleteCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("cluster not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Cluster not found",
		},
		{
			name:      "insufficient permissions",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockClusterService) {
				m.On("DeleteCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("insufficient permissions to delete cluster"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions to delete cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockClusterService)
			tt.mockSetup(mockService)

			handler := NewClusterHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.DELETE("/clusters/:clusterId", handler.DeleteCluster)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/clusters/"+tt.clusterID, nil)

			// Set user ID in context if provided
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
