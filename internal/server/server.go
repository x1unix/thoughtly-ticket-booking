package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/x1unix/thoughtly-ticket-booking/internal/booking"
	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
)

const (
	shutdownTimeout = 5 * time.Second
	idleTimeout     = 10 * time.Second
)

type Server struct {
	logger *zap.SugaredLogger
	cfg    *config.Config
	db     *pgxpool.Pool
	rdb    redis.UniversalClient
	svc    *booking.Service
	app    *fiber.App
}

func NewServer(ctx context.Context, logger *zap.Logger, cfg *config.Config) (*Server, error) {
	db, err := cfg.DB.NewPgxPool(ctx)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to conn to db: %w", err)
	}

	rdb, err := cfg.Redis.NewRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger: logger.Sugar(),
		cfg:    cfg,
		db:     db,
		rdb:    rdb,
		svc:    booking.NewService(db, rdb),
	}, nil
}

func (srv *Server) mountRoutes(app *fiber.App) {
	// Endpoints for tests
	app.Post("/api/events", srv.handleCreateEvent)
	app.Get("/api/ping", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	// Client API
	app.Get("/api/events", srv.handleListEvents)
	app.Get("/api/events/:eventID/tiers", srv.handleListTiersSummary)
	app.Post("/api/events/:eventID/reserve", srv.handleReserveTickets)
}

func (srv *Server) Listen(ctx context.Context) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		IdleTimeout:           idleTimeout,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			if code == fiber.StatusInternalServerError {
				srv.logger.Error(err.Error())
			}

			return c.Status(code).JSON(&ErrorResponse{
				Error: err.Error(),
			})
		},
	})

	app.Use(fiberRecover.New())
	srv.mountRoutes(app)
	srv.app = app

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
}

func (srv *Server) Close() {
	defer srv.db.Close()
	defer srv.rdb.Close()

	if srv.app != nil {
		srv.app.ShutdownWithTimeout(shutdownTimeout)
	}
}

func (srv *Server) ListenAndWait(ctx context.Context) {
	srv.Listen(ctx)
	<-ctx.Done()
	srv.Close()
}
