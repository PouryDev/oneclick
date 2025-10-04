package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/domain"
)

// MockApplicationRepository is a mock implementation of ApplicationRepository
type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) CreateApplication(ctx context.Context, app *domain.Application) (*domain.Application, error) {
	args := m.Called(ctx, app)
	return args.Get(0).(*domain.Application), args.Error(1)
}

func (m *MockApplicationRepository) GetApplicationByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Application), args.Error(1)
}

func (m *MockApplicationRepository) GetApplicationsByClusterID(ctx context.Context, clusterID uuid.UUID) ([]domain.ApplicationSummary, error) {
	args := m.Called(ctx, clusterID)
	return args.Get(0).([]domain.ApplicationSummary), args.Error(1)
}

func (m *MockApplicationRepository) GetApplicationByNameInCluster(ctx context.Context, clusterID uuid.UUID, name string) (*domain.Application, error) {
	args := m.Called(ctx, clusterID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Application), args.Error(1)
}

func (m *MockApplicationRepository) DeleteApplication(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockClusterRepository is a mock implementation of ClusterRepository
type MockClusterRepository struct {
	mock.Mock
}

func (m *MockClusterRepository) CreateCluster(ctx context.Context, cluster *domain.Cluster) (*domain.Cluster, error) {
	args := m.Called(ctx, cluster)
	return args.Get(0).(*domain.Cluster), args.Error(1)
}

func (m *MockClusterRepository) GetClusterByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cluster), args.Error(1)
}

func (m *MockClusterRepository) GetClustersByOrgID(ctx context.Context, orgID uuid.UUID) ([]domain.ClusterSummary, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.ClusterSummary), args.Error(1)
}

func (m *MockClusterRepository) UpdateClusterStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Cluster, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(*domain.Cluster), args.Error(1)
}

func (m *MockClusterRepository) UpdateClusterKubeconfig(ctx context.Context, id uuid.UUID, kubeconfigEncrypted []byte, status string) (*domain.Cluster, error) {
	args := m.Called(ctx, id, kubeconfigEncrypted, status)
	return args.Get(0).(*domain.Cluster), args.Error(1)
}

func (m *MockClusterRepository) UpdateClusterHealth(ctx context.Context, id uuid.UUID, kubeVersion string) (*domain.Cluster, error) {
	args := m.Called(ctx, id, kubeVersion)
	return args.Get(0).(*domain.Cluster), args.Error(1)
}

func (m *MockClusterRepository) UpdateClusterNodeCount(ctx context.Context, id uuid.UUID, nodeCount int) (*domain.Cluster, error) {
	args := m.Called(ctx, id, nodeCount)
	return args.Get(0).(*domain.Cluster), args.Error(1)
}

func (m *MockClusterRepository) DeleteCluster(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockOrganizationRepository is a mock implementation of OrganizationRepository
type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) CreateOrganization(ctx context.Context, name string) (*domain.Organization, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) UpdateOrganization(ctx context.Context, id uuid.UUID, name string) (*domain.Organization, error) {
	args := m.Called(ctx, id, name)
	return args.Get(0).(*domain.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationRepository) AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) (*domain.UserOrganization, error) {
	args := m.Called(ctx, userID, orgID, role)
	return args.Get(0).(*domain.UserOrganization), args.Error(1)
}

func (m *MockOrganizationRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]domain.UserOrganizationSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.UserOrganizationSummary), args.Error(1)
}

func (m *MockOrganizationRepository) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]domain.OrganizationMember), args.Error(1)
}

func (m *MockOrganizationRepository) GetUserRoleInOrganization(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID, orgID)
	return args.String(0), args.Error(1)
}

func (m *MockOrganizationRepository) UpdateUserRoleInOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) (*domain.UserOrganization, error) {
	args := m.Called(ctx, userID, orgID, role)
	return args.Get(0).(*domain.UserOrganization), args.Error(1)
}

func (m *MockOrganizationRepository) RemoveUserFromOrganization(ctx context.Context, userID, orgID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func (m *MockOrganizationRepository) IsUserOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// MockCryptoService is a mock implementation of CryptoService
type MockCryptoService struct {
	mock.Mock
}

func (m *MockCryptoService) Decrypt(encryptedData []byte) ([]byte, error) {
	args := m.Called(encryptedData)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) DecryptString(encryptedData string) (string, error) {
	args := m.Called(encryptedData)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoService) Encrypt(plaintext []byte) ([]byte, error) {
	args := m.Called(plaintext)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCryptoService) EncryptString(plaintext string) (string, error) {
	args := m.Called(plaintext)
	return args.String(0), args.Error(1)
}

// MockKubernetesClient is a mock implementation of KubernetesClientInterface
type MockKubernetesClient struct {
	mock.Mock
}

func (m *MockKubernetesClient) GetPodsByApp(ctx context.Context, appName, namespace string) ([]domain.Pod, error) {
	args := m.Called(ctx, appName, namespace)
	return args.Get(0).([]domain.Pod), args.Error(1)
}

func (m *MockKubernetesClient) GetPodDetail(ctx context.Context, podName, namespace string) (*domain.PodDetail, error) {
	args := m.Called(ctx, podName, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PodDetail), args.Error(1)
}

func (m *MockKubernetesClient) GetPodLogs(ctx context.Context, podName, namespace string, req domain.PodLogsRequest) (*domain.PodLogsResponse, error) {
	args := m.Called(ctx, podName, namespace, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PodLogsResponse), args.Error(1)
}

func (m *MockKubernetesClient) GetPodDescribe(ctx context.Context, podName, namespace string) (*domain.PodDescribeResponse, error) {
	args := m.Called(ctx, podName, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PodDescribeResponse), args.Error(1)
}

func (m *MockKubernetesClient) ExecInPod(ctx context.Context, podName, namespace string, req domain.PodExecRequest, conn *websocket.Conn) error {
	args := m.Called(ctx, podName, namespace, req, conn)
	return args.Error(0)
}

func TestPodService_GetPodsByApp(t *testing.T) {
	logger := zap.NewNop()

	// Create mocks
	mockAppRepo := &MockApplicationRepository{}
	mockClusterRepo := &MockClusterRepository{}
	mockOrgRepo := &MockOrganizationRepository{}
	mockCryptoService := &MockCryptoService{}
	mockKubeClient := &MockKubernetesClient{}

	// Create pod service
	podService := &podService{
		appRepo:       mockAppRepo,
		clusterRepo:   mockClusterRepo,
		orgRepo:       mockOrgRepo,
		cryptoService: mockCryptoService,
		kubeClient:    mockKubeClient,
		logger:        logger,
	}

	// Test data
	userID := uuid.New()
	appID := uuid.New()
	orgID := uuid.New()
	clusterID := uuid.New()

	app := &domain.Application{
		ID:        appID,
		OrgID:     orgID,
		ClusterID: clusterID,
		Name:      "test-app",
	}

	cluster := &domain.Cluster{
		ID:                  clusterID,
		OrgID:               orgID,
		KubeconfigEncrypted: []byte("encrypted-kubeconfig"),
	}

	// Setup mocks
	mockAppRepo.On("GetApplicationByID", mock.Anything, appID).Return(app, nil)
	mockOrgRepo.On("GetUserRoleInOrganization", mock.Anything, userID, orgID).Return("member", nil)
	mockClusterRepo.On("GetClusterByID", mock.Anything, clusterID).Return(cluster, nil)
	mockKubeClient.On("GetPodsByApp", mock.Anything, "test-app", "test-app").Return([]domain.Pod{}, nil)

	// Test
	pods, err := podService.GetPodsByApp(context.Background(), userID, appID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, pods)
	assert.Len(t, pods, 0) // Empty slice is expected from mock

	// Verify all mocks were called
	mockAppRepo.AssertExpectations(t)
	mockOrgRepo.AssertExpectations(t)
	mockClusterRepo.AssertExpectations(t)
	mockKubeClient.AssertExpectations(t)
}

func TestPodService_GetPodsByApp_ApplicationNotFound(t *testing.T) {
	logger := zap.NewNop()

	// Create mocks
	mockAppRepo := &MockApplicationRepository{}
	mockClusterRepo := &MockClusterRepository{}
	mockOrgRepo := &MockOrganizationRepository{}
	mockCryptoService := &MockCryptoService{}
	mockKubeClient := &MockKubernetesClient{}

	// Create pod service
	podService := &podService{
		appRepo:       mockAppRepo,
		clusterRepo:   mockClusterRepo,
		orgRepo:       mockOrgRepo,
		cryptoService: mockCryptoService,
		kubeClient:    mockKubeClient,
		logger:        logger,
	}

	// Test data
	userID := uuid.New()
	appID := uuid.New()

	// Setup mocks - application not found
	mockAppRepo.On("GetApplicationByID", mock.Anything, appID).Return(nil, errors.New("application not found"))

	// Test
	pods, err := podService.GetPodsByApp(context.Background(), userID, appID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, pods)
	assert.Contains(t, err.Error(), "application not found")

	// Verify mocks were called
	mockAppRepo.AssertExpectations(t)
}

func TestPodService_GetPodsByApp_Unauthorized(t *testing.T) {
	logger := zap.NewNop()

	// Create mocks
	mockAppRepo := &MockApplicationRepository{}
	mockClusterRepo := &MockClusterRepository{}
	mockOrgRepo := &MockOrganizationRepository{}
	mockCryptoService := &MockCryptoService{}
	mockKubeClient := &MockKubernetesClient{}

	// Create pod service
	podService := &podService{
		appRepo:       mockAppRepo,
		clusterRepo:   mockClusterRepo,
		orgRepo:       mockOrgRepo,
		cryptoService: mockCryptoService,
		kubeClient:    mockKubeClient,
		logger:        logger,
	}

	// Test data
	userID := uuid.New()
	appID := uuid.New()
	orgID := uuid.New()
	clusterID := uuid.New()

	app := &domain.Application{
		ID:        appID,
		OrgID:     orgID,
		ClusterID: clusterID,
		Name:      "test-app",
	}

	// Setup mocks - user not a member
	mockAppRepo.On("GetApplicationByID", mock.Anything, appID).Return(app, nil)
	mockOrgRepo.On("GetUserRoleInOrganization", mock.Anything, userID, orgID).Return("", nil)

	// Test
	pods, err := podService.GetPodsByApp(context.Background(), userID, appID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, pods)
	assert.Contains(t, err.Error(), "unauthorized")

	// Verify mocks were called
	mockAppRepo.AssertExpectations(t)
	mockOrgRepo.AssertExpectations(t)
}

func TestPodService_CreateAuditLog(t *testing.T) {
	logger := zap.NewNop()

	// Create mocks
	mockAppRepo := &MockApplicationRepository{}
	mockClusterRepo := &MockClusterRepository{}
	mockOrgRepo := &MockOrganizationRepository{}
	mockCryptoService := &MockCryptoService{}
	mockKubeClient := &MockKubernetesClient{}

	// Create pod service
	podService := &podService{
		appRepo:       mockAppRepo,
		clusterRepo:   mockClusterRepo,
		orgRepo:       mockOrgRepo,
		cryptoService: mockCryptoService,
		kubeClient:    mockKubeClient,
		logger:        logger,
	}

	// Test data
	userID := uuid.New()
	clusterID := uuid.New()
	appID := uuid.New()

	// Test
	err := podService.CreateAuditLog(context.Background(), userID, "pod_list", "test-pod", "test-namespace", clusterID, appID, "127.0.0.1", "test-agent")

	// Assertions
	assert.NoError(t, err)

	// Verify mocks were called (none expected for audit log)
	mockAppRepo.AssertExpectations(t)
	mockOrgRepo.AssertExpectations(t)
	mockClusterRepo.AssertExpectations(t)
	mockCryptoService.AssertExpectations(t)
}
