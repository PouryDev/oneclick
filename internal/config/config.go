package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./cmd/server")

	// Set default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("jwt.expiration", "24h")
	viper.SetDefault("logging.level", "info")

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if os.Getenv("ONECLICK_MASTER_KEY") == "" {
		return fmt.Errorf("ONECLICK_MASTER_KEY is required")
	}

	return nil
}

func GetMasterKey() string {
	return os.Getenv("ONECLICK_MASTER_KEY")
}
