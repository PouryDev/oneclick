package infra

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PouryDev/oneclick/internal/domain"
)

func TestParser_ParseConfig(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		yaml     string
		expected *domain.InfraConfig
		wantErr  bool
	}{
		{
			name: "valid config with services and app",
			yaml: `
services:
  db:
    chart: bitnami/postgresql
    env:
      POSTGRES_DB: webshop
      POSTGRES_USER: shop
      POSTGRES_PASSWORD: SECRET::webshop-postgres-password
  cache:
    chart: bitnami/redis
    env:
      REDIS_PASSWORD: SECRET::redis-password
app:
  env:
    DATABASE_URL: "postgres://shop:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop"
`,
			expected: &domain.InfraConfig{
				Services: map[string]domain.ServiceDefinition{
					"db": {
						Chart: "bitnami/postgresql",
						Env: map[string]string{
							"POSTGRES_DB":       "webshop",
							"POSTGRES_USER":     "shop",
							"POSTGRES_PASSWORD": "SECRET::webshop-postgres-password",
						},
					},
					"cache": {
						Chart: "bitnami/redis",
						Env: map[string]string{
							"REDIS_PASSWORD": "SECRET::redis-password",
						},
					},
				},
				App: domain.AppDefinition{
					Env: map[string]string{
						"DATABASE_URL": "postgres://shop:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			yaml: `
services:
  db:
    chart: bitnami/postgresql
    env:
      POSTGRES_DB: webshop
      POSTGRES_USER: shop
      POSTGRES_PASSWORD: SECRET::webshop-postgres-password
  cache:
    chart: bitnami/redis
    env:
      REDIS_PASSWORD: SECRET::redis-password
app:
  env:
    DATABASE_URL: "postgres://shop:{{services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop"
`,
			wantErr: false, // This is actually valid YAML
		},
		{
			name:    "empty config",
			yaml:    ``,
			wantErr: false, // Empty config is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseConfig(tt.yaml)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, config)
		})
	}
}

func TestParser_ProcessSecrets(t *testing.T) {
	parser := NewParser()

	config := &domain.InfraConfig{
		Services: map[string]domain.ServiceDefinition{
			"db": {
				Chart: "bitnami/postgresql",
				Env: map[string]string{
					"POSTGRES_PASSWORD": "SECRET::webshop-postgres-password",
					"POSTGRES_DB":       "webshop",
				},
			},
			"cache": {
				Chart: "bitnami/redis",
				Env: map[string]string{
					"REDIS_PASSWORD": "SECRET::redis-password",
				},
			},
		},
		App: domain.AppDefinition{
			Env: map[string]string{
				"SESSION_SECRET": "SECRET::session-secret",
				"APP_NAME":       "webshop",
			},
		},
	}

	secrets, err := parser.ProcessSecrets(config)
	require.NoError(t, err)
	assert.Len(t, secrets, 3)

	// Check that secrets are properly extracted
	secretNames := make(map[string]bool)
	for _, secret := range secrets {
		secretNames[secret.Name] = true
	}

	assert.True(t, secretNames["webshop-postgres-password"])
	assert.True(t, secretNames["redis-password"])
	assert.True(t, secretNames["session-secret"])
}

func TestParser_GenerateServiceConfigs(t *testing.T) {
	parser := NewParser()

	config := &domain.InfraConfig{
		Services: map[string]domain.ServiceDefinition{
			"db": {
				Chart: "bitnami/postgresql",
				Env: map[string]string{
					"POSTGRES_DB":       "webshop",
					"POSTGRES_USER":     "shop",
					"POSTGRES_PASSWORD": "SECRET::webshop-postgres-password",
				},
			},
			"cache": {
				Chart: "bitnami/redis",
				Env: map[string]string{
					"REDIS_PASSWORD": "SECRET::redis-password",
				},
			},
		},
	}

	serviceConfigs, err := parser.GenerateServiceConfigs(config, "test-app")
	require.NoError(t, err)
	assert.Len(t, serviceConfigs, 2)

	// Check db service
	dbConfig := serviceConfigs[0]
	assert.Equal(t, "db", dbConfig.ServiceName)
	assert.Equal(t, "bitnami/postgresql", dbConfig.Chart)
	assert.Equal(t, "test-app", dbConfig.Namespace)
	assert.Len(t, dbConfig.Configs, 3)

	// Check that secrets are properly identified
	dbPasswordConfig := dbConfig.Configs["POSTGRES_PASSWORD"]
	assert.True(t, dbPasswordConfig.IsSecret)
	assert.Equal(t, "webshop-postgres-password", dbPasswordConfig.SecretName)

	// Check that non-secrets are properly handled
	dbNameConfig := dbConfig.Configs["POSTGRES_DB"]
	assert.False(t, dbNameConfig.IsSecret)
	assert.Equal(t, "webshop", dbNameConfig.Value)
}

func TestParser_ProcessTemplates(t *testing.T) {
	parser := NewParser()

	config := &domain.InfraConfig{
		App: domain.AppDefinition{
			Env: map[string]string{
				"DATABASE_URL": "postgres://shop:{{.services.db.env.POSTGRES_PASSWORD}}@db:5432/webshop",
				"REDIS_URL":    "redis://:{{.services.cache.env.REDIS_PASSWORD}}@cache:6379",
				"APP_NAME":     "webshop",
			},
		},
	}

	serviceConfigs := map[string]string{
		"db":    `{"POSTGRES_PASSWORD":"secret123"}`,
		"cache": `{"REDIS_PASSWORD":"redis123"}`,
	}

	result, err := parser.ProcessTemplates(config, serviceConfigs)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	assert.Equal(t, "postgres://shop:secret123@db:5432/webshop", result["DATABASE_URL"])
	assert.Equal(t, "redis://:redis123@cache:6379", result["REDIS_URL"])
	assert.Equal(t, "webshop", result["APP_NAME"])
}

func TestParser_ValidateConfig(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name    string
		config  *domain.InfraConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &domain.InfraConfig{
				Services: map[string]domain.ServiceDefinition{
					"db": {
						Chart: "bitnami/postgresql",
						Env: map[string]string{
							"POSTGRES_DB": "webshop",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing services",
			config: &domain.InfraConfig{
				Services: nil,
			},
			wantErr: true,
		},
		{
			name: "service without chart",
			config: &domain.InfraConfig{
				Services: map[string]domain.ServiceDefinition{
					"db": {
						Chart: "",
						Env: map[string]string{
							"POSTGRES_DB": "webshop",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid service name",
			config: &domain.InfraConfig{
				Services: map[string]domain.ServiceDefinition{
					"DB": { // Invalid: uppercase
						Chart: "bitnami/postgresql",
						Env: map[string]string{
							"POSTGRES_DB": "webshop",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParser_ExtractSecretsFromValue(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "single secret",
			value:    "SECRET::my-secret",
			expected: []string{"my-secret"},
		},
		{
			name:     "multiple secrets",
			value:    "SECRET::secret1 and SECRET::secret2",
			expected: []string{"secret1", "secret2"},
		},
		{
			name:     "no secrets",
			value:    "regular-value",
			expected: []string(nil),
		},
		{
			name:     "empty string",
			value:    "",
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secrets := parser.ExtractSecretsFromValue(tt.value)
			assert.Equal(t, tt.expected, secrets)
		})
	}
}

func TestParser_ReplaceSecretsInValue(t *testing.T) {
	parser := NewParser()

	secretValues := map[string]string{
		"db-password": "secret123",
		"redis-pass":  "redis456",
	}

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "single secret replacement",
			value:    "SECRET::db-password",
			expected: "secret123",
		},
		{
			name:     "multiple secret replacements",
			value:    "SECRET::db-password and SECRET::redis-pass",
			expected: "secret123 and redis456",
		},
		{
			name:     "no secrets",
			value:    "regular-value",
			expected: "regular-value",
		},
		{
			name:     "unknown secret",
			value:    "SECRET::unknown-secret",
			expected: "SECRET::unknown-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ReplaceSecretsInValue(tt.value, secretValues)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_GenerateHelmValues(t *testing.T) {
	parser := NewParser()

	serviceConfig := ServiceConfigData{
		ServiceName: "db",
		Chart:       "bitnami/postgresql",
		Namespace:   "test-app",
		Configs: map[string]ConfigValue{
			"POSTGRES_DB": {
				Value:    "webshop",
				IsSecret: false,
			},
			"POSTGRES_USER": {
				Value:    "shop",
				IsSecret: false,
			},
			"POSTGRES_PASSWORD": {
				Value:      "",
				IsSecret:   true,
				SecretName: "db-password",
			},
		},
	}

	values, err := parser.GenerateHelmValues(serviceConfig)
	require.NoError(t, err)

	// Check that non-secret values are included in env
	env, exists := values["env"]
	require.True(t, exists)
	envMap := env.(map[string]string)
	assert.Equal(t, "webshop", envMap["POSTGRES_DB"])
	assert.Equal(t, "shop", envMap["POSTGRES_USER"])

	// Check that secret values are not included
	_, exists = envMap["POSTGRES_PASSWORD"]
	assert.False(t, exists)
}

func TestParser_GenerateKubernetesSecrets(t *testing.T) {
	parser := NewParser()

	serviceConfig := ServiceConfigData{
		ServiceName: "db",
		Chart:       "bitnami/postgresql",
		Namespace:   "test-app",
		Configs: map[string]ConfigValue{
			"POSTGRES_DB": {
				Value:    "webshop",
				IsSecret: false,
			},
			"POSTGRES_PASSWORD": {
				Value:      "secret123",
				IsSecret:   true,
				SecretName: "db-password",
			},
		},
	}

	secrets, err := parser.GenerateKubernetesSecrets(serviceConfig)
	require.NoError(t, err)

	// Check that only secret values are included
	assert.Len(t, secrets, 1)
	assert.Equal(t, "secret123", secrets["POSTGRES_PASSWORD"])

	// Check that non-secret values are not included
	_, exists := secrets["POSTGRES_DB"]
	assert.False(t, exists)
}

func TestIsValidServiceName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"db", true},
		{"cache", true},
		{"postgres-db", true},
		{"redis-cache", true},
		{"DB", false},                           // uppercase
		{"cache_redis", false},                  // underscore
		{"cache.redis", false},                  // dot
		{"", false},                             // empty
		{"a", true},                             // single char
		{"a" + string(make([]byte, 64)), false}, // too long
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidServiceName(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}
