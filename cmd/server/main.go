package main

import (
	"context"
	"fmt"
	"os"

	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
)

func main() {
	if err := config.LoadEnvFile(); err != nil {
		die(err)
		return
	}

	cfg, err := config.FromEnv()
	if err != nil {
		die(err)
		return
	}

	fmt.Printf("%#v\n", cfg)
	// ctx, cancelFn := signal.NotifyContext(context.Background())
	// defer cancelFn()
}

func die(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func run(ctx context.Context) error {
	return nil
}
