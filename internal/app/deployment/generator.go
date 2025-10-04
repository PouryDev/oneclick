package deployment

import (
	"fmt"
	"strings"

	"github.com/PouryDev/oneclick/internal/domain"
)

// DeploymentConfig represents configuration for generating Kubernetes deployments
type DeploymentConfig struct {
	AppName     string
	Namespace   string
	Image       string
	Tag         string
	Replicas    int32
	Port        int32
	Environment map[string]string
	Config      map[string]string
	Resources   *ResourceConfig
	HealthCheck *HealthCheckConfig
}

// ResourceConfig represents resource limits and requests
type ResourceConfig struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	LivenessPath  string
	ReadinessPath string
	Port          int32
}

// DeploymentGenerator generates Kubernetes deployment YAML
type DeploymentGenerator struct{}

// NewDeploymentGenerator creates a new deployment generator
func NewDeploymentGenerator() *DeploymentGenerator {
	return &DeploymentGenerator{}
}

// GenerateDeployment generates a Kubernetes Deployment YAML
func (g *DeploymentGenerator) GenerateDeployment(config *DeploymentConfig) (string, error) {
	if config.AppName == "" {
		return "", fmt.Errorf("app name is required")
	}
	if config.Image == "" {
		return "", fmt.Errorf("image is required")
	}
	if config.Tag == "" {
		return "", fmt.Errorf("tag is required")
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	replicas := config.Replicas
	if replicas == 0 {
		replicas = 1
	}

	port := config.Port
	if port == 0 {
		port = 8080
	}

	imageName := fmt.Sprintf("%s:%s", config.Image, config.Tag)

	var envVars strings.Builder
	for key, value := range config.Environment {
		envVars.WriteString(fmt.Sprintf("        - name: %s\n", key))
		envVars.WriteString(fmt.Sprintf("          value: \"%s\"\n", value))
	}

	var configVars strings.Builder
	for key, value := range config.Config {
		configVars.WriteString(fmt.Sprintf("        - name: %s\n", key))
		configVars.WriteString(fmt.Sprintf("          value: \"%s\"\n", value))
	}

	var resources strings.Builder
	if config.Resources != nil {
		resources.WriteString("        resources:\n")
		if config.Resources.CPURequest != "" || config.Resources.MemoryRequest != "" {
			resources.WriteString("          requests:\n")
			if config.Resources.CPURequest != "" {
				resources.WriteString(fmt.Sprintf("            cpu: %s\n", config.Resources.CPURequest))
			}
			if config.Resources.MemoryRequest != "" {
				resources.WriteString(fmt.Sprintf("            memory: %s\n", config.Resources.MemoryRequest))
			}
		}
		if config.Resources.CPULimit != "" || config.Resources.MemoryLimit != "" {
			resources.WriteString("          limits:\n")
			if config.Resources.CPULimit != "" {
				resources.WriteString(fmt.Sprintf("            cpu: %s\n", config.Resources.CPULimit))
			}
			if config.Resources.MemoryLimit != "" {
				resources.WriteString(fmt.Sprintf("            memory: %s\n", config.Resources.MemoryLimit))
			}
		}
	}

	var healthChecks strings.Builder
	if config.HealthCheck != nil {
		if config.HealthCheck.LivenessPath != "" {
			healthChecks.WriteString("        livenessProbe:\n")
			healthChecks.WriteString("          httpGet:\n")
			healthChecks.WriteString(fmt.Sprintf("            path: %s\n", config.HealthCheck.LivenessPath))
			healthChecks.WriteString(fmt.Sprintf("            port: %d\n", config.HealthCheck.Port))
			healthChecks.WriteString("          initialDelaySeconds: 30\n")
			healthChecks.WriteString("          periodSeconds: 10\n")
		}
		if config.HealthCheck.ReadinessPath != "" {
			healthChecks.WriteString("        readinessProbe:\n")
			healthChecks.WriteString("          httpGet:\n")
			healthChecks.WriteString(fmt.Sprintf("            path: %s\n", config.HealthCheck.ReadinessPath))
			healthChecks.WriteString(fmt.Sprintf("            port: %d\n", config.HealthCheck.Port))
			healthChecks.WriteString("          initialDelaySeconds: 5\n")
			healthChecks.WriteString("          periodSeconds: 5\n")
		}
	}

	yaml := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
    version: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
        version: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
        env:%s%s%s%s
`,
		config.AppName,
		namespace,
		config.AppName,
		config.Tag,
		replicas,
		config.AppName,
		config.AppName,
		config.Tag,
		config.AppName,
		imageName,
		port,
		envVars.String(),
		configVars.String(),
		resources.String(),
		healthChecks.String(),
	)

	return yaml, nil
}

// GenerateService generates a Kubernetes Service YAML
func (g *DeploymentGenerator) GenerateService(config *DeploymentConfig) (string, error) {
	if config.AppName == "" {
		return "", fmt.Errorf("app name is required")
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	port := config.Port
	if port == 0 {
		port = 8080
	}

	yaml := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s-service
  namespace: %s
  labels:
    app: %s
spec:
  selector:
    app: %s
  ports:
  - port: 80
    targetPort: %d
    protocol: TCP
  type: ClusterIP
`,
		config.AppName,
		namespace,
		config.AppName,
		config.AppName,
		port,
	)

	return yaml, nil
}

// GenerateIngress generates a Kubernetes Ingress YAML
func (g *DeploymentGenerator) GenerateIngress(config *DeploymentConfig, domains []string) (string, error) {
	if config.AppName == "" {
		return "", fmt.Errorf("app name is required")
	}
	if len(domains) == 0 {
		return "", fmt.Errorf("at least one domain is required")
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	var hosts strings.Builder
	var paths strings.Builder
	for i, domain := range domains {
		hosts.WriteString(fmt.Sprintf("    - host: %s\n", domain))
		if i == 0 {
			paths.WriteString(fmt.Sprintf("      http:\n        paths:\n        - path: /\n          pathType: Prefix\n          backend:\n            service:\n              name: %s-service\n              port:\n                number: 80\n", config.AppName))
		}
	}

	yaml := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s-ingress
  namespace: %s
  labels:
    app: %s
spec:
  rules:%s%s
`,
		config.AppName,
		namespace,
		config.AppName,
		hosts.String(),
		paths.String(),
	)

	return yaml, nil
}

// GenerateConfigMap generates a Kubernetes ConfigMap YAML
func (g *DeploymentGenerator) GenerateConfigMap(config *DeploymentConfig) (string, error) {
	if config.AppName == "" {
		return "", fmt.Errorf("app name is required")
	}
	if len(config.Config) == 0 {
		return "", nil // No config map needed if no config
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = "default"
	}

	var data strings.Builder
	for key, value := range config.Config {
		data.WriteString(fmt.Sprintf("  %s: \"%s\"\n", key, value))
	}

	yaml := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-config
  namespace: %s
  labels:
    app: %s
data:
%s
`,
		config.AppName,
		namespace,
		config.AppName,
		data.String(),
	)

	return yaml, nil
}

// GenerateFromApplication creates deployment configuration from application and release
func (g *DeploymentGenerator) GenerateFromApplication(app *domain.Application, release *domain.Release, meta *domain.ReleaseMeta) *DeploymentConfig {
	config := &DeploymentConfig{
		AppName:   app.Name,
		Namespace: app.Name, // Use app name as namespace
		Image:     release.Image,
		Tag:       release.Tag,
		Replicas:  1,
		Port:      8080,
	}

	if meta != nil {
		config.Environment = meta.Environment
		config.Config = meta.Config
	}

	// Set default health check paths
	config.HealthCheck = &HealthCheckConfig{
		LivenessPath:  "/health",
		ReadinessPath: "/ready",
		Port:          8080,
	}

	// Set default resources
	config.Resources = &ResourceConfig{
		CPURequest:    "100m",
		CPULimit:      "500m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "512Mi",
	}

	return config
}

// GenerateAllManifests generates all Kubernetes manifests for an application
func (g *DeploymentGenerator) GenerateAllManifests(config *DeploymentConfig, domains []string) (map[string]string, error) {
	manifests := make(map[string]string)

	// Generate Deployment
	deployment, err := g.GenerateDeployment(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate deployment: %w", err)
	}
	manifests["deployment.yaml"] = deployment

	// Generate Service
	service, err := g.GenerateService(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service: %w", err)
	}
	manifests["service.yaml"] = service

	// Generate ConfigMap if needed
	if len(config.Config) > 0 {
		configMap, err := g.GenerateConfigMap(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate configmap: %w", err)
		}
		manifests["configmap.yaml"] = configMap
	}

	// Generate Ingress if domains provided
	if len(domains) > 0 {
		ingress, err := g.GenerateIngress(config, domains)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ingress: %w", err)
		}
		manifests["ingress.yaml"] = ingress
	}

	return manifests, nil
}
