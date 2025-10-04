package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type ClusterService interface {
	CreateCluster(ctx context.Context, userID, orgID uuid.UUID, req *domain.CreateClusterRequest) (*domain.ClusterResponse, error)
	ImportCluster(ctx context.Context, userID, orgID uuid.UUID, name string, kubeconfigData []byte) (*domain.ClusterResponse, error)
	GetClustersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.ClusterSummary, error)
	GetCluster(ctx context.Context, userID, clusterID uuid.UUID) (*domain.ClusterDetailResponse, error)
	GetClusterHealth(ctx context.Context, userID, clusterID uuid.UUID) (*domain.ClusterHealth, error)
	DeleteCluster(ctx context.Context, userID, clusterID uuid.UUID) error
}

type clusterService struct {
	clusterRepo repo.ClusterRepository
	orgRepo     repo.OrganizationRepository
	crypto      *crypto.Crypto
}

func NewClusterService(clusterRepo repo.ClusterRepository, orgRepo repo.OrganizationRepository, crypto *crypto.Crypto) ClusterService {
	return &clusterService{
		clusterRepo: clusterRepo,
		orgRepo:     orgRepo,
		crypto:      crypto,
	}
}

func (s *clusterService) CreateCluster(ctx context.Context, userID, orgID uuid.UUID, req *domain.CreateClusterRequest) (*domain.ClusterResponse, error) {
	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Validate cluster status
	if !domain.IsValidClusterStatus(domain.StatusProvisioning) {
		return nil, errors.New("invalid cluster status")
	}

	cluster := &domain.Cluster{
		OrgID:    orgID,
		Name:     req.Name,
		Provider: req.Provider,
		Region:   req.Region,
		Status:   domain.StatusProvisioning,
	}

	// If kubeconfig is provided, validate and encrypt it
	if req.Kubeconfig != nil && *req.Kubeconfig != "" {
		// Decode base64 kubeconfig
		kubeconfigData, err := base64.StdEncoding.DecodeString(*req.Kubeconfig)
		if err != nil {
			return nil, errors.New("invalid base64 kubeconfig")
		}

		// Validate kubeconfig by trying to create a client
		if err := s.validateKubeconfig(kubeconfigData); err != nil {
			return nil, fmt.Errorf("invalid kubeconfig: %w", err)
		}

		// Encrypt kubeconfig
		encryptedKubeconfig, err := s.crypto.Encrypt(kubeconfigData)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt kubeconfig: %w", err)
		}

		cluster.KubeconfigEncrypted = encryptedKubeconfig
		cluster.Status = domain.StatusActive
	}

	createdCluster, err := s.clusterRepo.CreateCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}

	response := createdCluster.ToResponse()
	return &response, nil
}

func (s *clusterService) ImportCluster(ctx context.Context, userID, orgID uuid.UUID, name string, kubeconfigData []byte) (*domain.ClusterResponse, error) {
	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Validate kubeconfig
	if err := s.validateKubeconfig(kubeconfigData); err != nil {
		return nil, fmt.Errorf("invalid kubeconfig: %w", err)
	}

	// Encrypt kubeconfig
	encryptedKubeconfig, err := s.crypto.Encrypt(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt kubeconfig: %w", err)
	}

	cluster := &domain.Cluster{
		OrgID:               orgID,
		Name:                name,
		Provider:            "imported", // Default provider for imported clusters
		Region:              "unknown",  // Default region for imported clusters
		KubeconfigEncrypted: encryptedKubeconfig,
		Status:              domain.StatusActive,
	}

	createdCluster, err := s.clusterRepo.CreateCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}

	response := createdCluster.ToResponse()
	return &response, nil
}

func (s *clusterService) GetClustersByOrg(ctx context.Context, userID, orgID uuid.UUID) ([]domain.ClusterSummary, error) {
	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	clusters, err := s.clusterRepo.GetClustersByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

func (s *clusterService) GetCluster(ctx context.Context, userID, clusterID uuid.UUID) (*domain.ClusterDetailResponse, error) {
	// Get cluster
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	response := cluster.ToDetailResponse()
	return &response, nil
}

func (s *clusterService) GetClusterHealth(ctx context.Context, userID, clusterID uuid.UUID) (*domain.ClusterHealth, error) {
	// Get cluster
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		return nil, err
	}
	if role == "" {
		return nil, errors.New("user does not have access to this organization")
	}

	// Check if cluster has kubeconfig
	if len(cluster.KubeconfigEncrypted) == 0 {
		return nil, errors.New("cluster does not have kubeconfig")
	}

	// Decrypt kubeconfig
	kubeconfigData, err := s.crypto.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt kubeconfig: %w", err)
	}

	// Get cluster health
	health, err := s.getClusterHealthFromKubeconfig(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster health: %w", err)
	}

	// Update cluster health info in database
	if health.KubeVersion != "" {
		_, err = s.clusterRepo.UpdateClusterHealth(ctx, clusterID, health.KubeVersion)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to update cluster health: %v\n", err)
		}
	}

	return health, nil
}

func (s *clusterService) DeleteCluster(ctx context.Context, userID, clusterID uuid.UUID) error {
	// Get cluster
	cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterID)
	if err != nil {
		return err
	}
	if cluster == nil {
		return errors.New("cluster not found")
	}

	// Check if user has access to the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, cluster.OrgID)
	if err != nil {
		return err
	}
	if role == "" {
		return errors.New("user does not have access to this organization")
	}

	// Only allow owners and admins to delete clusters
	if role != domain.RoleOwner && role != domain.RoleAdmin {
		return errors.New("insufficient permissions to delete cluster")
	}

	err = s.clusterRepo.DeleteCluster(ctx, clusterID)
	if err != nil {
		return err
	}

	return nil
}

// validateKubeconfig validates a kubeconfig by trying to create a Kubernetes client
func (s *clusterService) validateKubeconfig(kubeconfigData []byte) error {
	// Create REST config from kubeconfig
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return fmt.Errorf("failed to create REST config: %w", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Test connection by getting server version
	_, err = clientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	return nil
}

// getClusterHealthFromKubeconfig gets cluster health information using the kubeconfig
func (s *clusterService) getClusterHealthFromKubeconfig(kubeconfigData []byte) (*domain.ClusterHealth, error) {
	// Create REST config from kubeconfig
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config: %w", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get server version
	version, err := clientset.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Convert nodes to domain format
	nodeInfos := make([]domain.NodeInfo, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		nodeInfo := domain.NodeInfo{
			Name:   node.Name,
			Status: string(node.Status.Phase),
		}

		// Get CPU and memory from node status
		if node.Status.Capacity != nil {
			if cpu, exists := node.Status.Capacity["cpu"]; exists {
				nodeInfo.CPU = cpu.String()
			}
			if memory, exists := node.Status.Capacity["memory"]; exists {
				nodeInfo.Memory = memory.String()
			}
		}

		nodeInfos = append(nodeInfos, nodeInfo)
	}

	health := &domain.ClusterHealth{
		Status:      domain.StatusActive,
		KubeVersion: version.GitVersion,
		Nodes:       nodeInfos,
		LastCheck:   time.Now(),
	}

	return health, nil
}
