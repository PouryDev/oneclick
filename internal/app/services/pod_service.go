package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/app/kubeclient"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

type PodService interface {
	GetPodsByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.Pod, error)
	GetPodDetail(ctx context.Context, userID uuid.UUID, podName, namespace string) (*domain.PodDetail, error)
	GetPodLogs(ctx context.Context, userID uuid.UUID, podName, namespace string, req domain.PodLogsRequest) (*domain.PodLogsResponse, error)
	GetPodDescribe(ctx context.Context, userID uuid.UUID, podName, namespace string) (*domain.PodDescribeResponse, error)
	ExecInPod(ctx context.Context, userID uuid.UUID, podName, namespace string, req domain.PodExecRequest, conn *websocket.Conn) error
	CreateAuditLog(ctx context.Context, userID uuid.UUID, action, resource, namespace string, clusterID, appID uuid.UUID, ipAddress, userAgent string) error
}

type podService struct {
	appRepo       repo.ApplicationRepository
	clusterRepo   repo.ClusterRepository
	orgRepo       repo.OrganizationRepository
	cryptoService crypto.CryptoService
	kubeClient    kubeclient.KubernetesClientInterface
	logger        *zap.Logger
}

func NewPodService(
	appRepo repo.ApplicationRepository,
	clusterRepo repo.ClusterRepository,
	orgRepo repo.OrganizationRepository,
	cryptoService crypto.CryptoService,
	kubeClient kubeclient.KubernetesClientInterface,
	logger *zap.Logger,
) PodService {
	return &podService{
		appRepo:       appRepo,
		clusterRepo:   clusterRepo,
		orgRepo:       orgRepo,
		cryptoService: cryptoService,
		kubeClient:    kubeClient,
		logger:        logger,
	}
}

func (s *podService) GetPodsByApp(ctx context.Context, userID, appID uuid.UUID) ([]domain.Pod, error) {
	// Get application details
	app, err := s.appRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		s.logger.Error("Failed to get application", zap.Error(err), zap.String("appID", appID.String()))
		return nil, fmt.Errorf("application not found or access denied: %w", err)
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

	// Get cluster details
	cluster, err := s.clusterRepo.GetClusterByID(ctx, app.ClusterID)
	if err != nil {
		s.logger.Error("Failed to get cluster", zap.Error(err), zap.String("clusterID", app.ClusterID.String()))
		return nil, fmt.Errorf("cluster not found: %w", err)
	}
	if cluster == nil {
		return nil, errors.New("cluster not found")
	}

	// Use the injected Kubernetes client or create one if nil
	var kubeClient kubeclient.KubernetesClientInterface
	if s.kubeClient != nil {
		kubeClient = s.kubeClient
	} else {
		// Create Kubernetes client from cluster's kubeconfig
		kubeconfigBytes, err := s.cryptoService.Decrypt(cluster.KubeconfigEncrypted)
		if err != nil {
			s.logger.Error("Failed to decrypt kubeconfig", zap.Error(err))
			return nil, errors.New("failed to decrypt cluster credentials")
		}

		kubeClient, err = kubeclient.NewKubernetesClient(kubeconfigBytes, s.logger)
		if err != nil {
			s.logger.Error("Failed to create Kubernetes client", zap.Error(err))
			return nil, errors.New("failed to create Kubernetes client")
		}
	}

	// Get pods from Kubernetes
	podInfos, err := kubeClient.GetPodsByApp(ctx, app.Name, app.Name) // Using app name as namespace
	if err != nil {
		s.logger.Error("Failed to get pods from Kubernetes", zap.Error(err), zap.String("appName", app.Name))
		return nil, errors.New("failed to retrieve pods from cluster")
	}

	// Convert to domain models
	pods := make([]domain.Pod, 0, len(podInfos))
	for _, podInfo := range podInfos {
		pod := domain.Pod{
			ID:        uuid.New(),
			AppID:     appID,
			Name:      podInfo.Name,
			Namespace: podInfo.Namespace,
			Status:    podInfo.Status,
			Restarts:  podInfo.Restarts,
			Ready:     podInfo.Ready,
			Age:       podInfo.Age,
			NodeName:  podInfo.NodeName,
			Labels:    podInfo.Labels,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

// findApplicationByNamespace finds the application that owns pods in the given namespace
// This helper function searches through user's organizations and their clusters to find the app
func (s *podService) findApplicationByNamespace(ctx context.Context, userID uuid.UUID, namespace string) (*domain.Application, *domain.Cluster, error) {
	// Get user's organizations first
	userOrgs, err := s.orgRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user organizations", zap.Error(err), zap.String("userID", userID.String()))
		return nil, nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	// For each organization, get clusters and then applications
	for _, userOrg := range userOrgs {
		// Get clusters in this organization
		clusters, err := s.clusterRepo.GetClustersByOrgID(ctx, userOrg.ID)
		if err != nil {
			s.logger.Error("Failed to get clusters for organization", zap.Error(err), zap.String("orgID", userOrg.ID.String()))
			continue // Continue to next organization
		}

		// For each cluster, get applications
		for _, clusterSummary := range clusters {
			// Get applications in this cluster
			apps, err := s.appRepo.GetApplicationsByClusterID(ctx, clusterSummary.ID)
			if err != nil {
				s.logger.Error("Failed to get applications for cluster", zap.Error(err), zap.String("clusterID", clusterSummary.ID.String()))
				continue // Continue to next cluster
			}

			// Look for an application with a name that matches the namespace
			for _, appSummary := range apps {
				if appSummary.Name == namespace {
					// Found the application! Now get the full application and cluster details
					app, err := s.appRepo.GetApplicationByID(ctx, appSummary.ID)
					if err != nil {
						s.logger.Error("Failed to get application details", zap.Error(err), zap.String("appID", appSummary.ID.String()))
						continue
					}

					cluster, err := s.clusterRepo.GetClusterByID(ctx, clusterSummary.ID)
					if err != nil {
						s.logger.Error("Failed to get cluster details", zap.Error(err), zap.String("clusterID", clusterSummary.ID.String()))
						continue
					}

					return app, cluster, nil
				}
			}
		}
	}

	return nil, nil, errors.New("application not found for the given namespace")
}

func (s *podService) GetPodDetail(ctx context.Context, userID uuid.UUID, podName, namespace string) (*domain.PodDetail, error) {
	// Find the application that owns this pod
	app, cluster, err := s.findApplicationByNamespace(ctx, userID, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to find application: %w", err)
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

	// Use the injected Kubernetes client or create one if nil
	var kubeClient kubeclient.KubernetesClientInterface
	if s.kubeClient != nil {
		kubeClient = s.kubeClient
	} else {
		// Create Kubernetes client from cluster's kubeconfig
		kubeconfigBytes, err := s.cryptoService.Decrypt(cluster.KubeconfigEncrypted)
		if err != nil {
			s.logger.Error("Failed to decrypt kubeconfig", zap.Error(err))
			return nil, errors.New("failed to decrypt cluster credentials")
		}

		kubeClient, err = kubeclient.NewKubernetesClient(kubeconfigBytes, s.logger)
		if err != nil {
			s.logger.Error("Failed to create Kubernetes client", zap.Error(err))
			return nil, errors.New("failed to create Kubernetes client")
		}
	}

	// Get pod detail from Kubernetes
	podDetail, err := kubeClient.GetPodDetail(ctx, podName, namespace)
	if err != nil {
		s.logger.Error("Failed to get pod detail from Kubernetes", zap.Error(err), zap.String("podName", podName), zap.String("namespace", namespace))
		return nil, errors.New("failed to retrieve pod detail from cluster")
	}

	return podDetail, nil
}

func (s *podService) GetPodLogs(ctx context.Context, userID uuid.UUID, podName, namespace string, req domain.PodLogsRequest) (*domain.PodLogsResponse, error) {
	// Find the application that owns this pod
	app, cluster, err := s.findApplicationByNamespace(ctx, userID, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to find application: %w", err)
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

	// Use the injected Kubernetes client or create one if nil
	var kubeClient kubeclient.KubernetesClientInterface
	if s.kubeClient != nil {
		kubeClient = s.kubeClient
	} else {
		// Create Kubernetes client from cluster's kubeconfig
		kubeconfigBytes, err := s.cryptoService.Decrypt(cluster.KubeconfigEncrypted)
		if err != nil {
			s.logger.Error("Failed to decrypt kubeconfig", zap.Error(err))
			return nil, errors.New("failed to decrypt cluster credentials")
		}

		kubeClient, err = kubeclient.NewKubernetesClient(kubeconfigBytes, s.logger)
		if err != nil {
			s.logger.Error("Failed to create Kubernetes client", zap.Error(err))
			return nil, errors.New("failed to create Kubernetes client")
		}
	}

	// Get pod logs from Kubernetes
	podLogs, err := kubeClient.GetPodLogs(ctx, podName, namespace, req)
	if err != nil {
		s.logger.Error("Failed to get pod logs from Kubernetes", zap.Error(err), zap.String("podName", podName), zap.String("namespace", namespace))
		return nil, errors.New("failed to retrieve pod logs from cluster")
	}

	return podLogs, nil
}

func (s *podService) GetPodDescribe(ctx context.Context, userID uuid.UUID, podName, namespace string) (*domain.PodDescribeResponse, error) {
	// Find the application that owns this pod
	app, cluster, err := s.findApplicationByNamespace(ctx, userID, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to find application: %w", err)
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

	// Use the injected Kubernetes client or create one if nil
	var kubeClient kubeclient.KubernetesClientInterface
	if s.kubeClient != nil {
		kubeClient = s.kubeClient
	} else {
		// Create Kubernetes client from cluster's kubeconfig
		kubeconfigBytes, err := s.cryptoService.Decrypt(cluster.KubeconfigEncrypted)
		if err != nil {
			s.logger.Error("Failed to decrypt kubeconfig", zap.Error(err))
			return nil, errors.New("failed to decrypt cluster credentials")
		}

		kubeClient, err = kubeclient.NewKubernetesClient(kubeconfigBytes, s.logger)
		if err != nil {
			s.logger.Error("Failed to create Kubernetes client", zap.Error(err))
			return nil, errors.New("failed to create Kubernetes client")
		}
	}

	// Get pod describe from Kubernetes
	podDescribe, err := kubeClient.GetPodDescribe(ctx, podName, namespace)
	if err != nil {
		s.logger.Error("Failed to get pod describe from Kubernetes", zap.Error(err), zap.String("podName", podName), zap.String("namespace", namespace))
		return nil, errors.New("failed to retrieve pod describe from cluster")
	}

	return podDescribe, nil
}

func (s *podService) ExecInPod(ctx context.Context, userID uuid.UUID, podName, namespace string, req domain.PodExecRequest, conn *websocket.Conn) error {
	// Find the application that owns this pod
	app, cluster, err := s.findApplicationByNamespace(ctx, userID, namespace)
	if err != nil {
		return fmt.Errorf("failed to find application: %w", err)
	}

	// Check if user is member of the organization
	role, err := s.orgRepo.GetUserRoleInOrganization(ctx, userID, app.OrgID)
	if err != nil {
		s.logger.Error("Failed to check organization role", zap.Error(err), zap.String("orgID", app.OrgID.String()), zap.String("userID", userID.String()))
		return fmt.Errorf("failed to check organization role: %w", err)
	}
	if role == "" {
		return errors.New("unauthorized: user is not a member of the organization")
	}

	// Use the injected Kubernetes client or create one if nil
	var kubeClient kubeclient.KubernetesClientInterface
	if s.kubeClient != nil {
		kubeClient = s.kubeClient
	} else {
		// Create Kubernetes client from cluster's kubeconfig
		kubeconfigBytes, err := s.cryptoService.Decrypt(cluster.KubeconfigEncrypted)
		if err != nil {
			s.logger.Error("Failed to decrypt kubeconfig", zap.Error(err))
			return errors.New("failed to decrypt cluster credentials")
		}

		kubeClient, err = kubeclient.NewKubernetesClient(kubeconfigBytes, s.logger)
		if err != nil {
			s.logger.Error("Failed to create Kubernetes client", zap.Error(err))
			return errors.New("failed to create Kubernetes client")
		}
	}

	// Execute command in pod via WebSocket
	err = kubeClient.ExecInPod(ctx, podName, namespace, req, conn)
	if err != nil {
		s.logger.Error("Failed to execute command in pod", zap.Error(err), zap.String("podName", podName), zap.String("namespace", namespace))
		return fmt.Errorf("failed to execute command in pod: %w", err)
	}

	return nil
}

func (s *podService) CreateAuditLog(ctx context.Context, userID uuid.UUID, action, resource, namespace string, clusterID, appID uuid.UUID, ipAddress, userAgent string) error {
	// This would typically be implemented in a repository
	// For now, just log the audit event
	s.logger.Info("Audit log entry",
		zap.String("userID", userID.String()),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("namespace", namespace),
		zap.String("clusterID", clusterID.String()),
		zap.String("appID", appID.String()),
		zap.String("ipAddress", ipAddress),
		zap.String("userAgent", userAgent),
	)
	return nil
}
