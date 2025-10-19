package server

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
)

const (
	shutdownTimeout = 5 * time.Second
	idleTimeout     = 10 * time.Second
)

type Server struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
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
	_ = rdb
	_ = db

	return nil, nil
}

func (srv *Server) Listen(ctx context.Context) {
	app := fiber.New(fiber.Config{
		IdleTimeout: idleTimeout,
	})

	srv.logger.Infof("listening on %q", srv.cfg.HTTP.ListenAddress)
	go func() {
		err := app.Listen(srv.cfg.HTTP.ListenAddress, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			srv.logger.Fatal("failed to start HTTP server:", err)
		}
	}()

	<-ctx.Done()
	app.ShutdownWithTimeout(shutdownTimeout)
}
