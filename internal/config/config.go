// Package config provides app configuration primitives.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type HTTPConfig struct {
	ListenAddress string `envconfig:"ADDR" default:":8000"`
}

type Config struct {
	DB    DBConfig    `envconfig:"DB"`
	Redis RedisConfig `envconfig:"REDIS"`
	Log   LogConfig   `envconfig:"LOG"`
	HTTP  HTTPConfig  `envconfig:"ADDR"`
}

// LoadEnvFile populates environment variables from env file (if specified in a flag).
func LoadEnvFile() error {
	envFilePath, ok := os.LookupEnv("ENV_FILE")
	if !ok || envFilePath == "" {
		return nil
	}

	err := godotenv.Load(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to load env from file %q: %w", envFilePath, err)
	}

	return nil
}

// FromEnv loads and returns config from environment variables.
func FromEnv() (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process("APP", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// DBConfigFromEnv loads and returns database config from environment variables.
//
// Used for migration by goose.
func DBConfigFromEnv() (*DBConfig, error) {
	cfg := &DBConfig{}

	err := envconfig.Process("APP_DB", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}
