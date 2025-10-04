package deployment

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/PouryDev/oneclick/internal/domain"
)

func TestDeploymentGenerator_GenerateDeployment(t *testing.T) {
	generator := NewDeploymentGenerator()

	config := &DeploymentConfig{
		AppName:   "test-app",
		Namespace: "test-namespace",
		Image:     "myapp",
		Tag:       "latest",
		Replicas:  2,
		Port:      8080,
		Environment: map[string]string{
			"ENV": "production",
		},
		Config: map[string]string{
			"CONFIG_KEY": "config_value",
		},
		Resources: &ResourceConfig{
			CPURequest:    "100m",
			CPULimit:      "500m",
			MemoryRequest: "128Mi",
			MemoryLimit:   "512Mi",
		},
		HealthCheck: &HealthCheckConfig{
			LivenessPath:  "/health",
			ReadinessPath: "/ready",
			Port:          8080,
		},
	}

	yaml, err := generator.GenerateDeployment(config)
	assert.NoError(t, err)
	assert.Contains(t, yaml, "kind: Deployment")
	assert.Contains(t, yaml, "name: test-app")
	assert.Contains(t, yaml, "namespace: test-namespace")
	assert.Contains(t, yaml, "image: myapp:latest")
	assert.Contains(t, yaml, "replicas: 2")
	assert.Contains(t, yaml, "containerPort: 8080")
	assert.Contains(t, yaml, "ENV")
	assert.Contains(t, yaml, "production")
	assert.Contains(t, yaml, "CONFIG_KEY")
	assert.Contains(t, yaml, "config_value")
	assert.Contains(t, yaml, "livenessProbe")
	assert.Contains(t, yaml, "readinessProbe")
}

func TestDeploymentGenerator_GenerateService(t *testing.T) {
	generator := NewDeploymentGenerator()

	config := &DeploymentConfig{
		AppName: "test-app",
		Port:    8080,
	}

	yaml, err := generator.GenerateService(config)
	assert.NoError(t, err)
	assert.Contains(t, yaml, "kind: Service")
	assert.Contains(t, yaml, "name: test-app-service")
	assert.Contains(t, yaml, "port: 80")
	assert.Contains(t, yaml, "targetPort: 8080")
}

func TestDeploymentGenerator_GenerateIngress(t *testing.T) {
	generator := NewDeploymentGenerator()

	config := &DeploymentConfig{
		AppName: "test-app",
	}

	domains := []string{"example.com", "app.example.com"}

	yaml, err := generator.GenerateIngress(config, domains)
	assert.NoError(t, err)
	assert.Contains(t, yaml, "kind: Ingress")
	assert.Contains(t, yaml, "name: test-app-ingress")
	assert.Contains(t, yaml, "host: example.com")
	assert.Contains(t, yaml, "host: app.example.com")
}

func TestDeploymentGenerator_GenerateConfigMap(t *testing.T) {
	generator := NewDeploymentGenerator()

	config := &DeploymentConfig{
		AppName: "test-app",
		Config: map[string]string{
			"CONFIG_KEY1": "value1",
			"CONFIG_KEY2": "value2",
		},
	}

	yaml, err := generator.GenerateConfigMap(config)
	assert.NoError(t, err)
	assert.Contains(t, yaml, "kind: ConfigMap")
	assert.Contains(t, yaml, "name: test-app-config")
	assert.Contains(t, yaml, "CONFIG_KEY1")
	assert.Contains(t, yaml, "value1")
	assert.Contains(t, yaml, "CONFIG_KEY2")
	assert.Contains(t, yaml, "value2")
}

func TestDeploymentGenerator_GenerateFromApplication(t *testing.T) {
	generator := NewDeploymentGenerator()

	app := &domain.Application{
		ID:            uuid.New(),
		Name:          "test-app",
		DefaultBranch: "main",
	}

	release := &domain.Release{
		ID:    uuid.New(),
		Image: "myapp",
		Tag:   "v1.0.0",
	}

	meta := &domain.ReleaseMeta{
		Environment: map[string]string{
			"ENV": "production",
		},
		Config: map[string]string{
			"CONFIG_KEY": "config_value",
		},
	}

	config := generator.GenerateFromApplication(app, release, meta)

	assert.Equal(t, "test-app", config.AppName)
	assert.Equal(t, "test-app", config.Namespace)
	assert.Equal(t, "myapp", config.Image)
	assert.Equal(t, "v1.0.0", config.Tag)
	assert.Equal(t, int32(1), config.Replicas)
	assert.Equal(t, int32(8080), config.Port)
	assert.Equal(t, "production", config.Environment["ENV"])
	assert.Equal(t, "config_value", config.Config["CONFIG_KEY"])
	assert.NotNil(t, config.HealthCheck)
	assert.NotNil(t, config.Resources)
}

func TestDeploymentGenerator_GenerateAllManifests(t *testing.T) {
	generator := NewDeploymentGenerator()

	config := &DeploymentConfig{
		AppName: "test-app",
		Image:   "myapp",
		Tag:     "latest",
		Config: map[string]string{
			"CONFIG_KEY": "config_value",
		},
	}

	domains := []string{"example.com"}

	manifests, err := generator.GenerateAllManifests(config, domains)
	assert.NoError(t, err)

	// Should have deployment, service, configmap, and ingress
	assert.Contains(t, manifests, "deployment.yaml")
	assert.Contains(t, manifests, "service.yaml")
	assert.Contains(t, manifests, "configmap.yaml")
	assert.Contains(t, manifests, "ingress.yaml")

	// Check deployment content
	deployment := manifests["deployment.yaml"]
	assert.Contains(t, deployment, "kind: Deployment")
	assert.Contains(t, deployment, "name: test-app")

	// Check service content
	service := manifests["service.yaml"]
	assert.Contains(t, service, "kind: Service")
	assert.Contains(t, service, "name: test-app-service")

	// Check configmap content
	configmap := manifests["configmap.yaml"]
	assert.Contains(t, configmap, "kind: ConfigMap")
	assert.Contains(t, configmap, "name: test-app-config")

	// Check ingress content
	ingress := manifests["ingress.yaml"]
	assert.Contains(t, ingress, "kind: Ingress")
	assert.Contains(t, ingress, "name: test-app-ingress")
}

func TestDeploymentGenerator_GenerateAllManifests_NoConfig(t *testing.T) {
	generator := NewDeploymentGenerator()

	config := &DeploymentConfig{
		AppName: "test-app",
		Image:   "myapp",
		Tag:     "latest",
		// No config provided
	}

	manifests, err := generator.GenerateAllManifests(config, []string{})
	assert.NoError(t, err)

	// Should have deployment and service, but no configmap or ingress
	assert.Contains(t, manifests, "deployment.yaml")
	assert.Contains(t, manifests, "service.yaml")
	assert.NotContains(t, manifests, "configmap.yaml")
	assert.NotContains(t, manifests, "ingress.yaml")
}
