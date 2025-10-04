package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/PouryDev/oneclick/internal/domain"
)

// Parser parses infra-config.yml files
type Parser struct{}

// NewParser creates a new infra-config parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseConfig parses the infra-config.yml content
func (p *Parser) ParseConfig(content string) (*domain.InfraConfig, error) {
	var config domain.InfraConfig
	err := yaml.Unmarshal([]byte(content), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// ProcessSecrets detects SECRET:: markers and creates secret references
func (p *Parser) ProcessSecrets(config *domain.InfraConfig) ([]domain.SecretReference, error) {
	var secrets []domain.SecretReference
	secretPattern := regexp.MustCompile(`SECRET::([a-zA-Z0-9_-]+)`)

	// Process service environment variables
	for serviceName, serviceDef := range config.Services {
		for key, value := range serviceDef.Env {
			matches := secretPattern.FindAllStringSubmatch(value, -1)
			for _, match := range matches {
				if len(match) > 1 {
					secretName := match[1]
					secrets = append(secrets, domain.SecretReference{
						Name: secretName,
						Key:  fmt.Sprintf("%s.%s", serviceName, key),
					})
				}
			}
		}
	}

	// Process app environment variables
	for key, value := range config.App.Env {
		matches := secretPattern.FindAllStringSubmatch(value, -1)
		for _, match := range matches {
			if len(match) > 1 {
				secretName := match[1]
				secrets = append(secrets, domain.SecretReference{
					Name: secretName,
					Key:  fmt.Sprintf("app.%s", key),
				})
			}
		}
	}

	return secrets, nil
}

// GenerateServiceConfigs creates service configurations from parsed config
func (p *Parser) GenerateServiceConfigs(config *domain.InfraConfig, appID string) ([]ServiceConfigData, error) {
	var serviceConfigs []ServiceConfigData

	for serviceName, serviceDef := range config.Services {
		// Create service configuration
		serviceConfig := ServiceConfigData{
			ServiceName: serviceName,
			Chart:       serviceDef.Chart,
			Namespace:   appID, // Use app ID as namespace
			Configs:     make(map[string]ConfigValue),
		}

		// Process environment variables
		for key, value := range serviceDef.Env {
			isSecret := strings.HasPrefix(value, "SECRET::")
			configValue := ConfigValue{
				Value:    value,
				IsSecret: isSecret,
			}

			if isSecret {
				// Extract secret name from SECRET::name format
				secretPattern := regexp.MustCompile(`SECRET::([a-zA-Z0-9_-]+)`)
				matches := secretPattern.FindStringSubmatch(value)
				if len(matches) > 1 {
					configValue.SecretName = matches[1]
					configValue.Value = "" // Clear the actual value for secrets
				}
			}

			serviceConfig.Configs[key] = configValue
		}

		serviceConfigs = append(serviceConfigs, serviceConfig)
	}

	return serviceConfigs, nil
}

// ProcessTemplates processes template substitutions in configuration
func (p *Parser) ProcessTemplates(config *domain.InfraConfig, serviceConfigs map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	// Create template data with service configurations
	templateData := make(map[string]interface{})
	templateData["services"] = make(map[string]interface{})

	// Add service configurations to template data
	for serviceName, configs := range serviceConfigs {
		serviceData := make(map[string]interface{})
		var envMap map[string]interface{}
		if err := json.Unmarshal([]byte(configs), &envMap); err != nil {
			return nil, fmt.Errorf("failed to parse service configs for %s: %w", serviceName, err)
		}
		serviceData["env"] = envMap
		templateData["services"].(map[string]interface{})[serviceName] = serviceData
	}

	// Process app environment variables
	for key, value := range config.App.Env {
		processedValue, err := p.processTemplate(value, templateData)
		if err != nil {
			return nil, fmt.Errorf("failed to process template for app.%s: %w", key, err)
		}
		result[key] = processedValue
	}

	return result, nil
}

// processTemplate processes a single template string
func (p *Parser) processTemplate(templateStr string, data interface{}) (string, error) {
	// Create template with sprig functions
	tmpl, err := template.New("config").Funcs(sprig.TxtFuncMap()).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ValidateConfig validates the infrastructure configuration
func (p *Parser) ValidateConfig(config *domain.InfraConfig) error {
	if config.Services == nil {
		return fmt.Errorf("services section is required")
	}

	for serviceName, serviceDef := range config.Services {
		if serviceDef.Chart == "" {
			return fmt.Errorf("chart is required for service %s", serviceName)
		}

		// Validate service name
		if !isValidServiceName(serviceName) {
			return fmt.Errorf("invalid service name: %s", serviceName)
		}
	}

	return nil
}

// ExtractSecretsFromValue extracts secret references from a value string
func (p *Parser) ExtractSecretsFromValue(value string) []string {
	var secrets []string
	secretPattern := regexp.MustCompile(`SECRET::([a-zA-Z0-9_-]+)`)
	matches := secretPattern.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		if len(match) > 1 {
			secrets = append(secrets, match[1])
		}
	}

	return secrets
}

// ReplaceSecretsInValue replaces secret references with actual values
func (p *Parser) ReplaceSecretsInValue(value string, secretValues map[string]string) string {
	secretPattern := regexp.MustCompile(`SECRET::([a-zA-Z0-9_-]+)`)
	return secretPattern.ReplaceAllStringFunc(value, func(match string) string {
		secretName := secretPattern.FindStringSubmatch(match)[1]
		if actualValue, exists := secretValues[secretName]; exists {
			return actualValue
		}
		return match // Keep original if secret not found
	})
}

// Helper types for internal processing

// ServiceConfigData represents service configuration data
type ServiceConfigData struct {
	ServiceName string
	Chart       string
	Namespace   string
	Configs     map[string]ConfigValue
}

// ConfigValue represents a configuration value
type ConfigValue struct {
	Value      string
	IsSecret   bool
	SecretName string
}

// isValidServiceName validates service name format
func isValidServiceName(name string) bool {
	// Service names should be lowercase alphanumeric with hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, name)
	return matched && len(name) > 0 && len(name) <= 63
}

// GenerateHelmValues generates Helm values from service configuration
func (p *Parser) GenerateHelmValues(serviceConfig ServiceConfigData) (map[string]interface{}, error) {
	values := make(map[string]interface{})

	// Add environment variables
	if len(serviceConfig.Configs) > 0 {
		env := make(map[string]string)
		for key, configValue := range serviceConfig.Configs {
			if !configValue.IsSecret {
				env[key] = configValue.Value
			}
		}
		if len(env) > 0 {
			values["env"] = env
		}
	}

	return values, nil
}

// GenerateKubernetesSecrets generates Kubernetes secret data from service configuration
func (p *Parser) GenerateKubernetesSecrets(serviceConfig ServiceConfigData) (map[string]string, error) {
	secrets := make(map[string]string)

	for key, configValue := range serviceConfig.Configs {
		if configValue.IsSecret && configValue.Value != "" {
			secrets[key] = configValue.Value
		}
	}

	return secrets, nil
}
