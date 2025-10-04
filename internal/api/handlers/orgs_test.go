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

// MockOrganizationService is a mock implementation of OrganizationService
type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) CreateOrganization(ctx context.Context, userID uuid.UUID, req *domain.CreateOrganizationRequest) (*domain.OrganizationResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationResponse), args.Error(1)
}

func (m *MockOrganizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domain.UserOrganizationSummary, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserOrganizationSummary), args.Error(1)
}

func (m *MockOrganizationService) GetOrganization(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationWithMembers, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationWithMembers), args.Error(1)
}

func (m *MockOrganizationService) AddMember(ctx context.Context, userID, orgID uuid.UUID, req *domain.AddMemberRequest) (*domain.OrganizationMember, error) {
	args := m.Called(ctx, userID, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationMember), args.Error(1)
}

func (m *MockOrganizationService) UpdateMemberRole(ctx context.Context, userID, orgID, memberID uuid.UUID, req *domain.UpdateMemberRoleRequest) (*domain.OrganizationMember, error) {
	args := m.Called(ctx, userID, orgID, memberID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrganizationMember), args.Error(1)
}

func (m *MockOrganizationService) RemoveMember(ctx context.Context, userID, orgID, memberID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID, memberID)
	return args.Error(0)
}

func (m *MockOrganizationService) DeleteOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func TestOrganizationHandler_CreateOrganization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockOrganizationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful organization creation",
			userID: uuid.New().String(),
			requestBody: domain.CreateOrganizationRequest{
				Name: "Test Organization",
			},
			mockSetup: func(m *MockOrganizationService) {
				orgResponse := &domain.OrganizationResponse{
					ID:        uuid.New(),
					Name:      "Test Organization",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("CreateOrganization", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.CreateOrganizationRequest")).Return(orgResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "empty organization name",
			userID: uuid.New().String(),
			requestBody: domain.CreateOrganizationRequest{
				Name: "",
			},
			mockSetup:      func(m *MockOrganizationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Field validation",
		},
		{
			name:           "invalid request body",
			userID:         uuid.New().String(),
			requestBody:    map[string]string{"invalid": "data"},
			mockSetup:      func(m *MockOrganizationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Field validation",
		},
		{
			name:           "no user ID in context",
			userID:         "",
			requestBody:    domain.CreateOrganizationRequest{Name: "Test Org"},
			mockSetup:      func(m *MockOrganizationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrganizationService)
			tt.mockSetup(mockService)

			handler := NewOrganizationHandler(mockService)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/orgs", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Set user ID in context if provided
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.CreateOrganization(c)

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

func TestOrganizationHandler_GetUserOrganizations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockOrganizationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful get user organizations",
			userID: uuid.New().String(),
			mockSetup: func(m *MockOrganizationService) {
				orgs := []domain.UserOrganizationSummary{
					{
						ID:            uuid.New(),
						Name:          "Test Organization",
						Role:          domain.RoleOwner,
						ClustersCount: 0,
						AppsCount:     0,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				}
				m.On("GetUserOrganizations", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(orgs, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no user ID in context",
			userID:         "",
			mockSetup:      func(m *MockOrganizationService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrganizationService)
			tt.mockSetup(mockService)

			handler := NewOrganizationHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/orgs", nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Set user ID in context if provided
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.GetUserOrganizations(c)

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

func TestOrganizationHandler_AddMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		orgID          string
		requestBody    interface{}
		mockSetup      func(*MockOrganizationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful add member",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.AddMemberRequest{
				Email: "newmember@example.com",
				Role:  domain.RoleMember,
			},
			mockSetup: func(m *MockOrganizationService) {
				member := &domain.OrganizationMember{
					ID:    uuid.New(),
					Name:  "New Member",
					Email: "newmember@example.com",
					Role:  domain.RoleMember,
				}
				m.On("AddMember", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.AddMemberRequest")).Return(member, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "user not found",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.AddMemberRequest{
				Email: "nonexistent@example.com",
				Role:  domain.RoleMember,
			},
			mockSetup: func(m *MockOrganizationService) {
				m.On("AddMember", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.AddMemberRequest")).Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "User not found",
		},
		{
			name:   "insufficient permissions",
			userID: uuid.New().String(),
			orgID:  uuid.New().String(),
			requestBody: domain.AddMemberRequest{
				Email: "member@example.com",
				Role:  domain.RoleMember,
			},
			mockSetup: func(m *MockOrganizationService) {
				m.On("AddMember", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*domain.AddMemberRequest")).Return(nil, errors.New("insufficient permissions"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
		{
			name:           "invalid organization ID",
			userID:         uuid.New().String(),
			orgID:          "invalid-uuid",
			requestBody:    domain.AddMemberRequest{Email: "test@example.com", Role: domain.RoleMember},
			mockSetup:      func(m *MockOrganizationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid organization ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrganizationService)
			tt.mockSetup(mockService)

			handler := NewOrganizationHandler(mockService)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/orgs/"+tt.orgID+"/members", bytes.NewBuffer(jsonBody))
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
			handler.AddMember(c)

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

func TestOrganizationHandler_RemoveMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		orgID          string
		memberID       string
		mockSetup      func(*MockOrganizationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful remove member",
			userID:   uuid.New().String(),
			orgID:    uuid.New().String(),
			memberID: uuid.New().String(),
			mockSetup: func(m *MockOrganizationService) {
				m.On("RemoveMember", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "member not found",
			userID:   uuid.New().String(),
			orgID:    uuid.New().String(),
			memberID: uuid.New().String(),
			mockSetup: func(m *MockOrganizationService) {
				m.On("RemoveMember", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("member not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Member not found",
		},
		{
			name:     "insufficient permissions",
			userID:   uuid.New().String(),
			orgID:    uuid.New().String(),
			memberID: uuid.New().String(),
			mockSetup: func(m *MockOrganizationService) {
				m.On("RemoveMember", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).Return(errors.New("insufficient permissions"))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrganizationService)
			tt.mockSetup(mockService)

			handler := NewOrganizationHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/orgs/"+tt.orgID+"/members/"+tt.memberID, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "orgId", Value: tt.orgID},
				{Key: "userId", Value: tt.memberID},
			}

			// Set user ID in context if provided
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			// Call handler
			handler.RemoveMember(c)

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
