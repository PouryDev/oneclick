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

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResponse), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID string) (*domain.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserResponse), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			requestBody: domain.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				userResponse := &domain.UserResponse{
					ID:        uuid.New(),
					Name:      "John Doe",
					Email:     "john@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("Register", mock.Anything, mock.AnythingOfType("*domain.CreateUserRequest")).Return(userResponse, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "user already exists",
			requestBody: domain.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*domain.CreateUserRequest")).Return(nil, errors.New("user with this email already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "user with this email already exists",
		},
		{
			name: "invalid request body",
			requestBody: map[string]string{
				"name": "John Doe",
				// missing email and password
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Field validation",
		},
		{
			name: "password too short",
			requestBody: domain.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123", // too short
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Field validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockSetup(mockService)

			handler := NewAuthHandler(mockService)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.Register(c)

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

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: domain.LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				authResponse := &domain.AuthResponse{
					AccessToken: "jwt-token-here",
					User: domain.UserResponse{
						ID:        uuid.New(),
						Name:      "John Doe",
						Email:     "john@example.com",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				m.On("Login", mock.Anything, mock.AnythingOfType("*domain.LoginRequest")).Return(authResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			requestBody: domain.LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", mock.Anything, mock.AnythingOfType("*domain.LoginRequest")).Return(nil, errors.New("invalid email or password"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid email or password",
		},
		{
			name: "invalid request body",
			requestBody: map[string]string{
				"email": "john@example.com",
				// missing password
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Field validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockSetup(mockService)

			handler := NewAuthHandler(mockService)

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.Login(c)

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

func TestAuthHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "successful get user",
			userID: uuid.New().String(),
			mockSetup: func(m *MockAuthService) {
				userResponse := &domain.UserResponse{
					ID:        uuid.New(),
					Name:      "John Doe",
					Email:     "john@example.com",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.On("GetUserByID", mock.Anything, mock.AnythingOfType("string")).Return(userResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: uuid.New().String(),
			mockSetup: func(m *MockAuthService) {
				m.On("GetUserByID", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "User not found",
		},
		{
			name:           "no user ID in context",
			userID:         "",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuthService)
			tt.mockSetup(mockService)

			handler := NewAuthHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

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
			handler.Me(c)

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
