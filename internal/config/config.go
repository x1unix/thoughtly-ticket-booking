// Package config provides app configuration primitives.
package config

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

type DBConfig struct {
	URL string `required:"true"`
}

func (cfg DBConfig) NewPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
	// TODO: tracing
	conn, err := pgxpool.New(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to build Postgres client: %w", err)
	}

	return conn, nil
}

type RedisConfig struct {
	URL            string        `required:"true"`
	ConnectTimeout time.Duration `default:"10s"`
}

func (cfg RedisConfig) NewRedisClient(ctx context.Context) (redis.UniversalClient, error) {
	// TODO: support cluster + tracing
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)
	err = doWithTimeoutCtx(ctx, cfg.ConnectTimeout, func(ctx context.Context) error {
		return rdb.Ping(ctx).Err()
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return rdb, nil
}

func doWithTimeoutCtx(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	opCtx, cancelFn := context.WithTimeout(ctx, timeout)
	defer cancelFn()

	return fn(opCtx)
}

type Config struct {
	DB    DBConfig    `envconfig:"DB"`
	Redis RedisConfig `envconfig:"REDIS"`
	Log   LogConfig   `envconfig:"LOG"`
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
