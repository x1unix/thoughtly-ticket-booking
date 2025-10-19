package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
	"github.com/x1unix/thoughtly-ticket-booking/internal/server"
)

func main() {
	if err := config.LoadEnvFile(); err != nil {
		die(err)
	}

	cfg, err := config.FromEnv()
	if err != nil {
		die(err)
	}

	logger, err := cfg.Log.BuildZapLogger()
	if err != nil {
		die(err)
	}

	defer logger.Sync()
	if err := run(logger, cfg); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		logger.Sugar().Fatal(err)
	}
}

func die(args ...any) {
	msg := make([]any, 1, len(args)+1)
	msg[0] = "Error: "
	msg = append(msg, args...)
	fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}

func run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancelFn := signal.NotifyContext(context.Background())
	defer cancelFn()

	srv, err := server.NewServer(ctx, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to build server: %w", err)
	}

	return srv.Start(ctx)
}
