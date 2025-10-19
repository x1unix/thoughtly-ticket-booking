package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
)

type Server struct {
	logger *zap.SugaredLogger
	db     pgxpool.Conn
	rdb    redis.UniversalClient
}

func NewServer(ctx context.Context, logger *zap.Logger, cfg *config.Config) (*Server, error) {
	rdb, err := cfg.Redis.NewRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	db, err := cfg.DB.NewPgxPool(ctx)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (srv *Server) Start(ctx context.Context) error {
	// TODO
	return nil
}
