// Package config provides app configuration primitives.
package config

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type DBConfig struct {
	URL string `required:"true"`
}

type RedisConfig struct {
	URL string `required:"true"`
}

type Config struct {
	DB    DBConfig    `envconfig:"DB"`
	Redis RedisConfig `envconfig:"REDIS"`
}

// LoadEnvFile populates environment variables from env file (if specified in a flag).
func LoadEnvFile() error {
	var envFilePath string
	flag.StringVar(&envFilePath, "e", "", "Path to env file to load (optional)")
	flag.Parse()
	if envFilePath == "" {
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
