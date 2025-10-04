package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// MockApplicationService is a mock implementation of ApplicationService
type MockApplicationService struct {
	mock.Mock
}

func (m *MockApplicationService) CreateApplication(ctx context.Context, userID, clusterID uuid.UUID, req *domain.CreateApplicationRequest) (*domain.ApplicationResponse, error) {
	args := m.Called(ctx, userID, clusterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ApplicationResponse), args.Error(1)
}

func (m *MockApplicationService) GetApplicationsByCluster(ctx context.Context, userID, clusterID uuid.UUID) ([]domain.ApplicationSummary, error) {
	args := m.Called(ctx, userID, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ApplicationSummary), args.Error(1)
}

func (m *MockApplicationService) GetApplication(ctx context.Context, userID, appID uuid.UUID) (*domain.ApplicationDetail, error) {
	args := m.Called(ctx, userID, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ApplicationDetail), args.Error(1)
}

func (m *MockApplicationService) DeleteApplication(ctx context.Context, userID, appID uuid.UUID) error {
	args := m.Called(ctx, userID, appID)
	return args.Error(0)
}

func (m *MockApplicationService) DeployApplication(ctx context.Context, userID, appID uuid.UUID, req *domain.DeployApplicationRequest) (*domain.DeployApplicationResponse, error) {
	args := m.Called(ctx, userID, appID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeployApplicationResponse), args.Error(1)
}

func (m *MockApplicationService) RollbackApplication(ctx context.Context, userID, appID, releaseID uuid.UUID) (*domain.DeployApplicationResponse, error) {
	args := m.Called(ctx, userID, appID, releaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeployApplicationResponse), args.Error(1)
}

func (m *MockApplicationService) GetReleasesByApplication(ctx context.Context, userID, appID uuid.UUID) ([]domain.ReleaseSummary, error) {
	args := m.Called(ctx, userID, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ReleaseSummary), args.Error(1)
}

func TestApplicationHandler_CreateApplication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		clusterID      string
		requestBody    interface{}
		mockSetup      func(*MockApplicationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful application creation",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			requestBody: domain.CreateApplicationRequest{
				Name:          "test-app",
				RepoID:        uuid.New().String(),
				DefaultBranch: "main",
			},
			mockSetup: func(m *MockApplicationService) {
				appResponse := &domain.ApplicationResponse{
					ID:            uuid.New(),
					OrgID:         uuid.New(),
					ClusterID:     uuid.New(),
					Name:          "test-app",
					RepoID:        uuid.New(),
					DefaultBranch: "main",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				m.On("CreateApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateApplicationRequest")).Return(appResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "access denied",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			requestBody: domain.CreateApplicationRequest{
				Name:          "test-app",
				RepoID:        uuid.New().String(),
				DefaultBranch: "main",
			},
			mockSetup: func(m *MockApplicationService) {
				m.On("CreateApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateApplicationRequest")).Return(nil, errors.New("user does not have access to this organization"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
		{
			name:      "application name already exists",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			requestBody: domain.CreateApplicationRequest{
				Name:          "test-app",
				RepoID:        uuid.New().String(),
				DefaultBranch: "main",
			},
			mockSetup: func(m *MockApplicationService) {
				m.On("CreateApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateApplicationRequest")).Return(nil, errors.New("application name already exists in this cluster"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "application name already exists",
		},
		{
			name:           "invalid cluster ID",
			userID:         uuid.New().String(),
			clusterID:      "invalid-uuid",
			requestBody:    domain.CreateApplicationRequest{Name: "test-app", RepoID: uuid.New().String(), DefaultBranch: "main"},
			mockSetup:      func(m *MockApplicationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid cluster ID",
		},
		{
			name:           "no user ID in context",
			userID:         "",
			clusterID:      uuid.New().String(),
			requestBody:    domain.CreateApplicationRequest{Name: "test-app", RepoID: uuid.New().String(), DefaultBranch: "main"},
			mockSetup:      func(m *MockApplicationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			tt.mockSetup(mockService)

			handler := NewApplicationHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.POST("/clusters/:clusterId/apps", handler.CreateApplication)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/clusters/"+tt.clusterID+"/apps", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

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

func TestApplicationHandler_DeployApplication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		appID          string
		requestBody    interface{}
		mockSetup      func(*MockApplicationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful deployment",
			userID: uuid.New().String(),
			appID:  uuid.New().String(),
			requestBody: domain.DeployApplicationRequest{
				Image: "myapp:latest",
				Tag:   "latest",
			},
			mockSetup: func(m *MockApplicationService) {
				deployResponse := &domain.DeployApplicationResponse{
					ReleaseID: uuid.New(),
					Status:    "pending",
					Message:   "Deployment initiated",
				}
				m.On("DeployApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.DeployApplicationRequest")).Return(deployResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "application not found",
			userID: uuid.New().String(),
			appID:  uuid.New().String(),
			requestBody: domain.DeployApplicationRequest{
				Image: "myapp:latest",
				Tag:   "latest",
			},
			mockSetup: func(m *MockApplicationService) {
				m.On("DeployApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.DeployApplicationRequest")).Return(nil, errors.New("application not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Application not found",
		},
		{
			name:   "missing image",
			userID: uuid.New().String(),
			appID:  uuid.New().String(),
			requestBody: domain.DeployApplicationRequest{
				Tag: "latest",
			},
			mockSetup: func(m *MockApplicationService) {
				m.On("DeployApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.DeployApplicationRequest")).Return(nil, errors.New("image is required"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "image is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			tt.mockSetup(mockService)

			handler := NewApplicationHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.POST("/apps/:appId/deploy", handler.DeployApplication)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/apps/"+tt.appID+"/deploy", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

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

func TestApplicationHandler_GetApplicationsByCluster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		clusterID      string
		mockSetup      func(*MockApplicationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful get applications",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockApplicationService) {
				apps := []domain.ApplicationSummary{
					{
						ID:            uuid.New(),
						Name:          "app1",
						RepoID:        uuid.New(),
						DefaultBranch: "main",
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
					{
						ID:            uuid.New(),
						Name:          "app2",
						RepoID:        uuid.New(),
						DefaultBranch: "main",
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				m.On("GetApplicationsByCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(apps, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "access denied",
			userID:    uuid.New().String(),
			clusterID: uuid.New().String(),
			mockSetup: func(m *MockApplicationService) {
				m.On("GetApplicationsByCluster", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("user does not have access to this organization"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			tt.mockSetup(mockService)

			handler := NewApplicationHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.GET("/clusters/:clusterId/apps", handler.GetApplicationsByCluster)

			// Create request
			req := httptest.NewRequest("GET", "/clusters/"+tt.clusterID+"/apps", nil)

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

func TestApplicationHandler_DeleteApplication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		appID          string
		mockSetup      func(*MockApplicationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful application deletion",
			userID: uuid.New().String(),
			appID:  uuid.New().String(),
			mockSetup: func(m *MockApplicationService) {
				m.On("DeleteApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "application not found",
			userID: uuid.New().String(),
			appID:  uuid.New().String(),
			mockSetup: func(m *MockApplicationService) {
				m.On("DeleteApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("application not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Application not found",
		},
		{
			name:   "insufficient permissions",
			userID: uuid.New().String(),
			appID:  uuid.New().String(),
			mockSetup: func(m *MockApplicationService) {
				m.On("DeleteApplication", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("insufficient permissions to delete application"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions to delete application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			tt.mockSetup(mockService)

			handler := NewApplicationHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.DELETE("/apps/:appId", handler.DeleteApplication)

			// Create request
			req := httptest.NewRequest("DELETE", "/apps/"+tt.appID, nil)

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
