package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/PouryDev/oneclick/internal/app/crypto"
	"github.com/PouryDev/oneclick/internal/app/deployment"
	"github.com/PouryDev/oneclick/internal/domain"
	"github.com/PouryDev/oneclick/internal/repo"
)

// DeploymentWorker handles background deployment jobs
type DeploymentWorker struct {
	appRepo     repo.ApplicationRepository
	releaseRepo repo.ReleaseRepository
	clusterRepo repo.ClusterRepository
	crypto      *crypto.Crypto
	logger      *zap.Logger
	deployer    *deployment.DeploymentGenerator
}

// NewDeploymentWorker creates a new deployment worker
func NewDeploymentWorker(
	appRepo repo.ApplicationRepository,
	releaseRepo repo.ReleaseRepository,
	clusterRepo repo.ClusterRepository,
	crypto *crypto.Crypto,
	logger *zap.Logger,
) *DeploymentWorker {
	return &DeploymentWorker{
		appRepo:     appRepo,
		releaseRepo: releaseRepo,
		clusterRepo: clusterRepo,
		crypto:      crypto,
		logger:      logger,
		deployer:    deployment.NewDeploymentGenerator(),
	}
}

// DeploymentJob represents a deployment job
type DeploymentJob struct {
	ReleaseID uuid.UUID `json:"release_id"`
	AppID     uuid.UUID `json:"app_id"`
	Image     string    `json:"image"`
	Tag       string    `json:"tag"`
	CreatedAt time.Time `json:"created_at"`
}

// ProcessDeployment processes a deployment job
func (w *DeploymentWorker) ProcessDeployment(ctx context.Context, job *DeploymentJob) error {
	w.logger.Info("Processing deployment job",
		zap.String("release_id", job.ReleaseID.String()),
		zap.String("app_id", job.AppID.String()),
		zap.String("image", job.Image),
		zap.String("tag", job.Tag),
	)

	// Get application details
	app, err := w.appRepo.GetApplicationByID(ctx, job.AppID)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return fmt.Errorf("application not found")
	}

	// Get cluster details
	cluster, err := w.clusterRepo.GetClusterByID(ctx, app.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}
	if cluster == nil {
		return fmt.Errorf("cluster not found")
	}

	// Get release details
	release, err := w.releaseRepo.GetReleaseByID(ctx, job.ReleaseID)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}
	if release == nil {
		return fmt.Errorf("release not found")
	}

	// Update release status to running
	now := time.Now()
	_, err = w.releaseRepo.UpdateReleaseStatus(ctx, job.ReleaseID, domain.ReleaseStatusRunning, &now, nil)
	if err != nil {
		return fmt.Errorf("failed to update release status to running: %w", err)
	}

	// Decrypt kubeconfig
	kubeconfigBytes, err := w.crypto.Decrypt(cluster.KubeconfigEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt kubeconfig: %w", err)
	}

	// Create Kubernetes client
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	if err != nil {
		return fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Get release metadata
	meta, err := release.GetMeta()
	if err != nil {
		w.logger.Warn("Failed to parse release metadata", zap.Error(err))
		meta = &domain.ReleaseMeta{}
	}

	// Generate deployment configuration
	deployConfig := w.deployer.GenerateFromApplication(app, release, meta)

	// Deploy to Kubernetes
	err = w.deployToKubernetes(ctx, clientset, dynamicClient, deployConfig)
	if err != nil {
		// Update release status to failed
		finishedAt := time.Now()
		w.releaseRepo.UpdateReleaseStatus(ctx, job.ReleaseID, domain.ReleaseStatusFailed, nil, &finishedAt)
		return fmt.Errorf("failed to deploy to kubernetes: %w", err)
	}

	// Update release status to succeeded
	finishedAt := time.Now()
	_, err = w.releaseRepo.UpdateReleaseStatus(ctx, job.ReleaseID, domain.ReleaseStatusSucceeded, nil, &finishedAt)
	if err != nil {
		return fmt.Errorf("failed to update release status to succeeded: %w", err)
	}

	w.logger.Info("Deployment completed successfully",
		zap.String("release_id", job.ReleaseID.String()),
		zap.String("app_name", app.Name),
	)

	return nil
}

// deployToKubernetes deploys the application to Kubernetes
func (w *DeploymentWorker) deployToKubernetes(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, config *deployment.DeploymentConfig) error {
	// Generate deployment manifests
	manifests, err := w.deployer.GenerateAllManifests(config, []string{})
	if err != nil {
		return fmt.Errorf("failed to generate manifests: %w", err)
	}

	// Create namespace if it doesn't exist
	err = w.ensureNamespace(ctx, clientset, config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Apply each manifest
	for filename, manifest := range manifests {
		err = w.applyManifest(ctx, dynamicClient, manifest)
		if err != nil {
			return fmt.Errorf("failed to apply %s: %w", filename, err)
		}
		w.logger.Info("Applied manifest", zap.String("filename", filename))
	}

	// Wait for deployment to be ready
	err = w.waitForDeployment(ctx, clientset, config.Namespace, config.AppName)
	if err != nil {
		return fmt.Errorf("failed to wait for deployment: %w", err)
	}

	return nil
}

// ensureNamespace creates a namespace if it doesn't exist
func (w *DeploymentWorker) ensureNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil // Namespace already exists
	}

	// Create namespace
	_, err = clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"name": namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	w.logger.Info("Created namespace", zap.String("namespace", namespace))
	return nil
}

// applyManifest applies a Kubernetes manifest
func (w *DeploymentWorker) applyManifest(ctx context.Context, dynamicClient dynamic.Interface, manifest string) error {
	// Parse YAML
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(manifest), &obj)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Get GVR
	gvr := schema.GroupVersionResource{
		Group:    obj.GroupVersionKind().Group,
		Version:  obj.GroupVersionKind().Version,
		Resource: w.getResourceFromKind(obj.GetKind()),
	}

	// Apply the manifest
	_, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(ctx, &obj, metav1.CreateOptions{})
	if err != nil {
		// Try to update if create fails
		_, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Update(ctx, &obj, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create/update resource: %w", err)
		}
	}

	return nil
}

// getResourceFromKind converts a Kubernetes kind to a resource name
func (w *DeploymentWorker) getResourceFromKind(kind string) string {
	switch kind {
	case "Deployment":
		return "deployments"
	case "Service":
		return "services"
	case "ConfigMap":
		return "configmaps"
	case "Ingress":
		return "ingresses"
	default:
		return fmt.Sprintf("%ss", kind) // Simple pluralization
	}
}

// waitForDeployment waits for a deployment to be ready
func (w *DeploymentWorker) waitForDeployment(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for deployment to be ready")
		case <-ticker.C:
			deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				w.logger.Warn("Failed to get deployment status", zap.Error(err))
				continue
			}

			if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				w.logger.Info("Deployment is ready",
					zap.String("namespace", namespace),
					zap.String("name", name),
					zap.Int32("ready_replicas", deployment.Status.ReadyReplicas),
					zap.Int32("desired_replicas", *deployment.Spec.Replicas),
				)
				return nil
			}

			w.logger.Info("Waiting for deployment to be ready",
				zap.String("namespace", namespace),
				zap.String("name", name),
				zap.Int32("ready_replicas", deployment.Status.ReadyReplicas),
				zap.Int32("desired_replicas", *deployment.Spec.Replicas),
			)
		}
	}
}

// Start starts the deployment worker
func (w *DeploymentWorker) Start(ctx context.Context) {
	w.logger.Info("Starting deployment worker")

	// In a real implementation, this would:
	// 1. Connect to a message queue (Redis, RabbitMQ, etc.)
	// 2. Listen for deployment jobs
	// 3. Process jobs concurrently
	// 4. Handle retries and error recovery

	// For now, this is a stub that demonstrates the interface
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Deployment worker stopped")
			return
		case <-ticker.C:
			w.logger.Info("Deployment worker is running")
		}
	}
}
