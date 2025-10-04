package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
)

// MockInfrastructureService is a mock implementation of InfrastructureService
type MockInfrastructureService struct {
	mock.Mock
}

func (m *MockInfrastructureService) ProvisionServices(ctx context.Context, userID, appID uuid.UUID, infraConfigYAML string) (*domain.ProvisionServiceResponse, error) {
	args := m.Called(ctx, userID, appID, infraConfigYAML)
	return args.Get(0).(*domain.ProvisionServiceResponse), args.Error(1)
}

func (m *MockInfrastructureService) GetServicesByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.ServiceDetail, error) {
	args := m.Called(ctx, userID, appID)
	return args.Get(0).([]domain.ServiceDetail), args.Error(1)
}

func (m *MockInfrastructureService) GetServiceConfig(ctx context.Context, userID, configID uuid.UUID) (*domain.ServiceConfigRevealResponse, error) {
	args := m.Called(ctx, userID, configID)
	return args.Get(0).(*domain.ServiceConfigRevealResponse), args.Error(1)
}

func (m *MockInfrastructureService) UnprovisionService(ctx context.Context, userID, serviceID uuid.UUID) error {
	args := m.Called(ctx, userID, serviceID)
	return args.Error(0)
}

func TestInfrastructureHandler_ProvisionServices(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		appID          string
		requestBody    interface{}
		mockSetup      func(*MockInfrastructureService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "successful provisioning",
			appID: uuid.New().String(),
			requestBody: domain.ProvisionServiceRequest{
				InfraConfig: `
services:
  db:
    chart: bitnami/postgresql
    env:
      POSTGRES_DB: webshop
      POSTGRES_PASSWORD: SECRET::db-password
app:
  env:
    DATABASE_URL: "postgres://shop:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop"
`,
			},
			mockSetup: func(m *MockInfrastructureService) {
				response := &domain.ProvisionServiceResponse{
					Services: []domain.ServiceSummary{
						{
							ID:        uuid.New(),
							Name:      "db",
							Chart:     "bitnami/postgresql",
							Status:    domain.ServiceStatusPending,
							Namespace: "test-app",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					},
					Message: "Provisioning initiated for 1 services",
				}
				m.On("ProvisionServices", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string")).Return(response, nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:  "invalid app ID",
			appID: "invalid-uuid",
			requestBody: domain.ProvisionServiceRequest{
				InfraConfig: "services:\n  db:\n    chart: bitnami/postgresql",
			},
			mockSetup:      func(m *MockInfrastructureService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid application ID format",
		},
		{
			name:  "missing infra config",
			appID: uuid.New().String(),
			requestBody: domain.ProvisionServiceRequest{
				InfraConfig: "",
			},
			mockSetup:      func(m *MockInfrastructureService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "infra_config is required",
		},
		{
			name:  "service error",
			appID: uuid.New().String(),
			requestBody: domain.ProvisionServiceRequest{
				InfraConfig: "services:\n  db:\n    chart: bitnami/postgresql",
			},
			mockSetup: func(m *MockInfrastructureService) {
				m.On("ProvisionServices", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("string")).Return((*domain.ProvisionServiceResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to provision services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInfrastructureService)
			tt.mockSetup(mockService)

			handler := NewInfrastructureHandler(mockService, zap.NewNop())

			// Create request
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/apps/"+tt.appID+"/infra/provision", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "appId", Value: tt.appID}}
			c.Set("user_id", uuid.New().String())

			// Call handler
			handler.ProvisionServices(c)

			// Assert response
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

func TestInfrastructureHandler_GetServicesByApp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		appID          string
		mockSetup      func(*MockInfrastructureService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "successful retrieval",
			appID: uuid.New().String(),
			mockSetup: func(m *MockInfrastructureService) {
				services := []domain.ServiceDetail{
					{
						Service: domain.Service{
							ID:        uuid.New(),
							AppID:     uuid.New(),
							Name:      "db",
							Chart:     "bitnami/postgresql",
							Status:    domain.ServiceStatusRunning,
							Namespace: "test-app",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Configs: []domain.ServiceConfigSummary{
							{
								ID:       uuid.New(),
								Key:      "POSTGRES_DB",
								Value:    "webshop",
								IsSecret: false,
							},
						},
					},
				}
				m.On("GetServicesByApp", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(services, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "invalid app ID",
			appID: "invalid-uuid",
			mockSetup: func(m *MockInfrastructureService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid application ID format",
		},
		{
			name:  "service error",
			appID: uuid.New().String(),
			mockSetup: func(m *MockInfrastructureService) {
				m.On("GetServicesByApp", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return([]domain.ServiceDetail(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to retrieve services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInfrastructureService)
			tt.mockSetup(mockService)

			handler := NewInfrastructureHandler(mockService, zap.NewNop())

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/apps/"+tt.appID+"/infra/services", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "appId", Value: tt.appID}}
			c.Set("user_id", uuid.New().String())

			// Call handler
			handler.GetServicesByApp(c)

			// Assert response
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

func TestInfrastructureHandler_GetServiceConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		configID       string
		mockSetup      func(*MockInfrastructureService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful retrieval",
			configID: uuid.New().String(),
			mockSetup: func(m *MockInfrastructureService) {
				response := &domain.ServiceConfigRevealResponse{
					Config: domain.ServiceConfigResponse{
						ID:       uuid.New(),
						Key:      "POSTGRES_PASSWORD",
						Value:    "***MASKED***",
						IsSecret: true,
					},
				}
				m.On("GetServiceConfig", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid config ID",
			configID: "invalid-uuid",
			mockSetup: func(m *MockInfrastructureService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid service config ID format",
		},
		{
			name:     "service error",
			configID: uuid.New().String(),
			mockSetup: func(m *MockInfrastructureService) {
				m.On("GetServiceConfig", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return((*domain.ServiceConfigRevealResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to retrieve service configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInfrastructureService)
			tt.mockSetup(mockService)

			handler := NewInfrastructureHandler(mockService, zap.NewNop())

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/services/"+tt.configID+"/config", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "configId", Value: tt.configID}}
			c.Set("user_id", uuid.New().String())

			// Call handler
			handler.GetServiceConfig(c)

			// Assert response
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

func TestInfrastructureHandler_UnprovisionService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceID      string
		mockSetup      func(*MockInfrastructureService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful unprovisioning",
			serviceID: uuid.New().String(),
			mockSetup: func(m *MockInfrastructureService) {
				m.On("UnprovisionService", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:      "invalid service ID",
			serviceID: "invalid-uuid",
			mockSetup: func(m *MockInfrastructureService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid service ID format",
		},
		{
			name:      "service error",
			serviceID: uuid.New().String(),
			mockSetup: func(m *MockInfrastructureService) {
				m.On("UnprovisionService", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to unprovision service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockInfrastructureService)
			tt.mockSetup(mockService)

			handler := NewInfrastructureHandler(mockService, zap.NewNop())

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/services/"+tt.serviceID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "serviceId", Value: tt.serviceID}}
			c.Set("user_id", uuid.New().String())

			// Call handler
			handler.UnprovisionService(c)

			// Assert response
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
