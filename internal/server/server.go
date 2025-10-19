package server

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
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

	return &Server{
		logger: logger.Sugar(),
		cfg:    cfg,
	}, nil
}

func (srv *Server) Listen(ctx context.Context) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		IdleTimeout:           idleTimeout,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			srv.logger.Error(err)
			return fiber.DefaultErrorHandler(c, err)
		},
	})

	app.Use(fiberRecover.New())
	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	go func() {
		srv.logger.Infof("listening on %q", srv.cfg.HTTP.ListenAddress)
		err := app.Listen(srv.cfg.HTTP.ListenAddress)
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
