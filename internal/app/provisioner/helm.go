package provisioner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/PouryDev/oneclick/internal/domain"
)

// Provisioner interface for installing services
type Provisioner interface {
	Install(ctx context.Context, chart, namespace string, values map[string]interface{}) error
	Uninstall(ctx context.Context, releaseName, namespace string) error
	GetStatus(ctx context.Context, releaseName, namespace string) (string, error)
	Upgrade(ctx context.Context, releaseName, chart, namespace string, values map[string]interface{}) error
}

// HelmProvisioner implements Provisioner using Helm CLI
type HelmProvisioner struct {
	logger *zap.Logger
}

// NewHelmProvisioner creates a new Helm provisioner
func NewHelmProvisioner(logger *zap.Logger) *HelmProvisioner {
	return &HelmProvisioner{
		logger: logger,
	}
}

// Install installs a Helm chart
func (h *HelmProvisioner) Install(ctx context.Context, chart, namespace string, values map[string]interface{}) error {
	h.logger.Info("Installing Helm chart",
		zap.String("chart", chart),
		zap.String("namespace", namespace),
	)

	// Generate release name from chart
	releaseName := h.generateReleaseName(chart)

	// Build helm install command
	args := []string{
		"install",
		releaseName,
		chart,
		"--namespace", namespace,
		"--create-namespace",
		"--wait",
		"--timeout", "5m",
	}

	// Add values if provided
	if len(values) > 0 {
		valuesYAML, err := h.convertToYAML(values)
		if err != nil {
			return fmt.Errorf("failed to convert values to YAML: %w", err)
		}
		args = append(args, "--set-string", valuesYAML)
	}

	// Execute helm command
	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Error("Helm install failed",
			zap.String("chart", chart),
			zap.String("namespace", namespace),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return fmt.Errorf("helm install failed: %w", err)
	}

	h.logger.Info("Helm chart installed successfully",
		zap.String("chart", chart),
		zap.String("namespace", namespace),
		zap.String("release", releaseName),
	)

	return nil
}

// Uninstall uninstalls a Helm release
func (h *HelmProvisioner) Uninstall(ctx context.Context, releaseName, namespace string) error {
	h.logger.Info("Uninstalling Helm release",
		zap.String("release", releaseName),
		zap.String("namespace", namespace),
	)

	args := []string{
		"uninstall",
		releaseName,
		"--namespace", namespace,
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Error("Helm uninstall failed",
			zap.String("release", releaseName),
			zap.String("namespace", namespace),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return fmt.Errorf("helm uninstall failed: %w", err)
	}

	h.logger.Info("Helm release uninstalled successfully",
		zap.String("release", releaseName),
		zap.String("namespace", namespace),
	)

	return nil
}

// GetStatus gets the status of a Helm release
func (h *HelmProvisioner) GetStatus(ctx context.Context, releaseName, namespace string) (string, error) {
	args := []string{
		"status",
		releaseName,
		"--namespace", namespace,
		"--output", "json",
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("helm status failed: %w", err)
	}

	// Parse status from JSON output
	status := h.parseStatusFromJSON(string(output))
	return status, nil
}

// Upgrade upgrades a Helm release
func (h *HelmProvisioner) Upgrade(ctx context.Context, releaseName, chart, namespace string, values map[string]interface{}) error {
	h.logger.Info("Upgrading Helm release",
		zap.String("release", releaseName),
		zap.String("chart", chart),
		zap.String("namespace", namespace),
	)

	args := []string{
		"upgrade",
		releaseName,
		chart,
		"--namespace", namespace,
		"--wait",
		"--timeout", "5m",
	}

	// Add values if provided
	if len(values) > 0 {
		valuesYAML, err := h.convertToYAML(values)
		if err != nil {
			return fmt.Errorf("failed to convert values to YAML: %w", err)
		}
		args = append(args, "--set-string", valuesYAML)
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.logger.Error("Helm upgrade failed",
			zap.String("release", releaseName),
			zap.String("chart", chart),
			zap.String("namespace", namespace),
			zap.String("output", string(output)),
			zap.Error(err),
		)
		return fmt.Errorf("helm upgrade failed: %w", err)
	}

	h.logger.Info("Helm release upgraded successfully",
		zap.String("release", releaseName),
		zap.String("chart", chart),
		zap.String("namespace", namespace),
	)

	return nil
}

// generateReleaseName generates a release name from chart name
func (h *HelmProvisioner) generateReleaseName(chart string) string {
	// Extract chart name from chart string (e.g., "bitnami/postgresql" -> "postgresql")
	parts := strings.Split(chart, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return chart
}

// convertToYAML converts values map to YAML string for --set-string
func (h *HelmProvisioner) convertToYAML(values map[string]interface{}) (string, error) {
	var parts []string
	for key, value := range values {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(parts, ","), nil
}

// parseStatusFromJSON parses status from Helm JSON output
func (h *HelmProvisioner) parseStatusFromJSON(jsonOutput string) string {
	// Simple parsing - in a real implementation, you'd use proper JSON parsing
	if strings.Contains(jsonOutput, `"status":"deployed"`) {
		return "running"
	}
	if strings.Contains(jsonOutput, `"status":"failed"`) {
		return "failed"
	}
	if strings.Contains(jsonOutput, `"status":"pending"`) {
		return "provisioning"
	}
	return "unknown"
}

// KubernetesSecretManager manages Kubernetes secrets
type KubernetesSecretManager struct {
	clientset *kubernetes.Clientset
	logger    *zap.Logger
}

// NewKubernetesSecretManager creates a new Kubernetes secret manager
func NewKubernetesSecretManager(clientset *kubernetes.Clientset, logger *zap.Logger) *KubernetesSecretManager {
	return &KubernetesSecretManager{
		clientset: clientset,
		logger:    logger,
	}
}

// CreateSecret creates a Kubernetes secret
func (k *KubernetesSecretManager) CreateSecret(ctx context.Context, namespace, name string, data map[string]string) error {
	k.logger.Info("Creating Kubernetes secret",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: data,
	}

	_, err := k.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		k.logger.Error("Failed to create Kubernetes secret",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create secret: %w", err)
	}

	k.logger.Info("Kubernetes secret created successfully",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	return nil
}

// GetSecret gets a Kubernetes secret
func (k *KubernetesSecretManager) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	secret, err := k.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return secret, nil
}

// UpdateSecret updates a Kubernetes secret
func (k *KubernetesSecretManager) UpdateSecret(ctx context.Context, namespace, name string, data map[string]string) error {
	k.logger.Info("Updating Kubernetes secret",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	secret, err := k.GetSecret(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update secret data
	for key, value := range data {
		secret.StringData[key] = value
	}

	_, err = k.clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		k.logger.Error("Failed to update Kubernetes secret",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update secret: %w", err)
	}

	k.logger.Info("Kubernetes secret updated successfully",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	return nil
}

// DeleteSecret deletes a Kubernetes secret
func (k *KubernetesSecretManager) DeleteSecret(ctx context.Context, namespace, name string) error {
	k.logger.Info("Deleting Kubernetes secret",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	err := k.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		k.logger.Error("Failed to delete Kubernetes secret",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	k.logger.Info("Kubernetes secret deleted successfully",
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	return nil
}

// ServiceProvisioner orchestrates service provisioning
type ServiceProvisioner struct {
	provisioner Provisioner
	secretMgr   *KubernetesSecretManager
	logger      *zap.Logger
}

// NewServiceProvisioner creates a new service provisioner
func NewServiceProvisioner(provisioner Provisioner, secretMgr *KubernetesSecretManager, logger *zap.Logger) *ServiceProvisioner {
	return &ServiceProvisioner{
		provisioner: provisioner,
		secretMgr:   secretMgr,
		logger:      logger,
	}
}

// ProvisionService provisions a service with its configuration
func (s *ServiceProvisioner) ProvisionService(ctx context.Context, service *domain.Service, configs []domain.ServiceConfig) error {
	s.logger.Info("Provisioning service",
		zap.String("service", service.Name),
		zap.String("chart", service.Chart),
		zap.String("namespace", service.Namespace),
	)

	// Generate Helm values from configurations
	values := make(map[string]interface{})
	secrets := make(map[string]string)

	for _, config := range configs {
		if config.IsSecret {
			secrets[config.Key] = config.Value
		} else {
			values[config.Key] = config.Value
		}
	}

	// Create Kubernetes secrets if any
	if len(secrets) > 0 {
		secretName := fmt.Sprintf("%s-secrets", service.Name)
		err := s.secretMgr.CreateSecret(ctx, service.Namespace, secretName, secrets)
		if err != nil {
			return fmt.Errorf("failed to create secrets: %w", err)
		}
	}

	// Install Helm chart
	err := s.provisioner.Install(ctx, service.Chart, service.Namespace, values)
	if err != nil {
		return fmt.Errorf("failed to install chart: %w", err)
	}

	s.logger.Info("Service provisioned successfully",
		zap.String("service", service.Name),
		zap.String("chart", service.Chart),
		zap.String("namespace", service.Namespace),
	)

	return nil
}

// UnprovisionService removes a provisioned service
func (s *ServiceProvisioner) UnprovisionService(ctx context.Context, service *domain.Service) error {
	s.logger.Info("Unprovisioning service",
		zap.String("service", service.Name),
		zap.String("namespace", service.Namespace),
	)

	// Generate release name
	releaseName := s.generateReleaseName(service.Chart)

	// Uninstall Helm chart
	err := s.provisioner.Uninstall(ctx, releaseName, service.Namespace)
	if err != nil {
		return fmt.Errorf("failed to uninstall chart: %w", err)
	}

	// Delete secrets
	secretName := fmt.Sprintf("%s-secrets", service.Name)
	err = s.secretMgr.DeleteSecret(ctx, service.Namespace, secretName)
	if err != nil {
		s.logger.Warn("Failed to delete secrets",
			zap.String("service", service.Name),
			zap.String("secret", secretName),
			zap.Error(err),
		)
		// Don't fail the unprovisioning if secret deletion fails
	}

	s.logger.Info("Service unprovisioned successfully",
		zap.String("service", service.Name),
		zap.String("namespace", service.Namespace),
	)

	return nil
}

// generateReleaseName generates a release name from chart name
func (s *ServiceProvisioner) generateReleaseName(chart string) string {
	parts := strings.Split(chart, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return chart
}
