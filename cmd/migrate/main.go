package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/x1unix/thoughtly-ticket-booking/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	if err := config.LoadEnvFile(); err != nil {
		return err
	}

	cfg, err := config.DBConfigFromEnv()
	if err != nil {
		return err
	}

	args := flag.Args()
	if len(args) == 0 {
		return errors.New("missing goose command")
	}

	db, err := goose.OpenDBWithDriver("pgx", cfg.URL)
	if err != nil {
		return err
	}

	defer db.Close()

	cmd := args[0]
	args = args[1:]
	ctx := context.Background()
	err = goose.RunContext(ctx, cmd, db, "migrations", args...)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
