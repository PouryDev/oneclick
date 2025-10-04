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

// MockRepositoryService is a mock implementation of RepositoryService
type MockRepositoryService struct {
	mock.Mock
}

func (m *MockRepositoryService) CreateRepository(ctx context.Context, userID, orgID uuid.UUID, req *domain.CreateRepositoryRequest) (*domain.RepositoryResponse, error) {
	args := m.Called(ctx, userID, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RepositoryResponse), args.Error(1)
}

func (m *MockRepositoryService) GetRepositoriesByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.RepositorySummary, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RepositorySummary), args.Error(1)
}

func (m *MockRepositoryService) GetRepository(ctx context.Context, userID, repoID uuid.UUID) (*domain.RepositoryResponse, error) {
	args := m.Called(ctx, userID, repoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RepositoryResponse), args.Error(1)
}

func (m *MockRepositoryService) DeleteRepository(ctx context.Context, userID, repoID uuid.UUID) error {
	args := m.Called(ctx, userID, repoID)
	return args.Error(0)
}

func (m *MockRepositoryService) ProcessWebhook(ctx context.Context, provider string, payload []byte, signature string) error {
	args := m.Called(ctx, provider, payload, signature)
	return args.Error(0)
}

func TestRepositoryHandler_CreateRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		orgID          string
		requestBody    interface{}
		mockSetup      func(*MockRepositoryService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful repository creation",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateRepositoryRequest{
				Type:          "github",
				URL:           "https://github.com/user/repo.git",
				DefaultBranch: "main",
			},
			mockSetup: func(m *MockRepositoryService) {
				repoResponse := &domain.RepositoryResponse{
					ID:            uuid.New(),
					Type:          "github",
					URL:           "https://github.com/user/repo.git",
					DefaultBranch: "main",
					Config:        json.RawMessage("{}"),
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				m.On("CreateRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateRepositoryRequest")).Return(repoResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "repository creation with token",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateRepositoryRequest{
				Type:          "gitlab",
				URL:           "https://gitlab.com/user/repo.git",
				DefaultBranch: "main",
				Token:         "secret-token",
			},
			mockSetup: func(m *MockRepositoryService) {
				repoResponse := &domain.RepositoryResponse{
					ID:            uuid.New(),
					Type:          "gitlab",
					URL:           "https://gitlab.com/user/repo.git",
					DefaultBranch: "main",
					Config:        json.RawMessage(`{"token":"encrypted-token"}`),
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				m.On("CreateRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateRepositoryRequest")).Return(repoResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "access denied",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateRepositoryRequest{
				Type:          "github",
				URL:           "https://github.com/user/repo.git",
				DefaultBranch: "main",
			},
			mockSetup: func(m *MockRepositoryService) {
				m.On("CreateRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateRepositoryRequest")).Return(nil, errors.New("user does not have access to this organization"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
		{
			name:   "repository already exists",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.CreateRepositoryRequest{
				Type:          "github",
				URL:           "https://github.com/user/repo.git",
				DefaultBranch: "main",
			},
			mockSetup: func(m *MockRepositoryService) {
				m.On("CreateRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateRepositoryRequest")).Return(nil, errors.New("repository already exists for this organization"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "repository already exists",
		},
		{
			name:           "invalid organization ID",
			userID:         uuid.New().String(),
			orgID:          "invalid-uuid",
			requestBody:    domain.CreateRepositoryRequest{Type: "github", URL: "https://github.com/user/repo.git", DefaultBranch: "main"},
			mockSetup:      func(m *MockRepositoryService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid organization ID",
		},
		{
			name:           "no user ID in context",
			userID:         "",
			orgID:          uuid.New().String(),
			requestBody:    domain.CreateRepositoryRequest{Type: "github", URL: "https://github.com/user/repo.git", DefaultBranch: "main"},
			mockSetup:      func(m *MockRepositoryService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRepositoryService)
			tt.mockSetup(mockService)

			handler := NewRepositoryHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.POST("/orgs/:orgId/repos", handler.CreateRepository)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/orgs/"+tt.orgID+"/repos", bytes.NewBuffer(jsonBody))
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

func TestRepositoryHandler_GetRepositoriesByOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		orgID          string
		mockSetup      func(*MockRepositoryService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful get repositories",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			mockSetup: func(m *MockRepositoryService) {
				repos := []domain.RepositorySummary{
					{
						ID:            uuid.New(),
						Type:          "github",
						URL:           "https://github.com/user/repo1.git",
						DefaultBranch: "main",
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
					{
						ID:            uuid.New(),
						Type:          "gitlab",
						URL:           "https://gitlab.com/user/repo2.git",
						DefaultBranch: "main",
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				m.On("GetRepositoriesByOrg", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(repos, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "access denied",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			mockSetup: func(m *MockRepositoryService) {
				m.On("GetRepositoriesByOrg", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("user does not have access to this organization"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRepositoryService)
			tt.mockSetup(mockService)

			handler := NewRepositoryHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.GET("/orgs/:orgId/repos", handler.GetRepositoriesByOrg)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/orgs/"+tt.orgID+"/repos", nil)

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

func TestRepositoryHandler_DeleteRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		repoID         string
		mockSetup      func(*MockRepositoryService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful repository deletion",
			userID: uuid.New().String(),
			repoID: uuid.New().String(),
			mockSetup: func(m *MockRepositoryService) {
				m.On("DeleteRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "repository not found",
			userID: uuid.New().String(),
			repoID: uuid.New().String(),
			mockSetup: func(m *MockRepositoryService) {
				m.On("DeleteRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("repository not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Repository not found",
		},
		{
			name:   "insufficient permissions",
			userID: uuid.New().String(),
			repoID: uuid.New().String(),
			mockSetup: func(m *MockRepositoryService) {
				m.On("DeleteRepository", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("insufficient permissions to delete repository"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions to delete repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRepositoryService)
			tt.mockSetup(mockService)

			handler := NewRepositoryHandler(mockService)

			// Create gin router for testing
			router := gin.New()

			// Add middleware to set user_id in context
			router.Use(func(c *gin.Context) {
				if userID := c.GetHeader("X-User-ID"); userID != "" {
					c.Set("user_id", userID)
				}
				c.Next()
			})

			router.DELETE("/repos/:repoId", handler.DeleteRepository)

			// Create request
			req := httptest.NewRequest("DELETE", "/repos/"+tt.repoID, nil)

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
