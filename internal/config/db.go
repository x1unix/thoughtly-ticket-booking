package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
