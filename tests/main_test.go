package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
	"github.com/x1unix/thoughtly-ticket-booking/internal/server"
)

var client *Client

func TestMain(m *testing.M) {
	code, err := runTests(m)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(code)
}

func runTests(m *testing.M) (int, error) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	srv, err := initServer(ctx)
	if err != nil {
		return 1, fmt.Errorf("failed to init test environment: %w", err)
	}
	defer srv.Close()

	if err := client.WaitForServer(3, 300*time.Millisecond); err != nil {
		return 1, fmt.Errorf("failed to ping server: %w", err)
	}

	return m.Run(), nil
}

func initServer(ctx context.Context) (*server.Server, error) {
	if err := config.LoadEnvFile(); err != nil {
		return nil, err
	}

	cfg, err := config.FromEnv()
	if err != nil {
		return nil, err
	}

	cfg.Log.IsProduction = false
	logger, err := cfg.Log.BuildZapLogger()
	if err != nil {
		return nil, err
	}

	client, err = NewClient(cfg.HTTP.ListenAddress)
	if err != nil {
		return nil, err
	}

	// TODO: create a scratch DB instead of truncating main one
	if err := truncateDB(ctx, cfg.DB); err != nil {
		return nil, err
	}

	// TODO: spawn server at a random port (:0)
	srv, err := server.NewServer(ctx, logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create server")
	}

	srv.Listen(ctx)
	return srv, nil
}

func truncateDB(ctx context.Context, cfg config.DBConfig) error {
	db, err := cfg.NewPgxPool(ctx)
	if err != nil {
		return err
	}

	defer db.Close()
	queries := []string{
		`TRUNCATE TABLE tickets`,
		`TRUNCATE TABLE ticket_tiers`,
		`TRUNCATE TABLE events`,
	}

	for _, q := range queries {
		_, err := db.Exec(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to exec %q", q)
		}
	}

	return nil
}
