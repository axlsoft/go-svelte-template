// Command server is the single static binary for myapp.
//
// It wires config → logger → DB pool → router → HTTP server with graceful
// shutdown, exposes a `migrate` subcommand, and serves the embedded SPA.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/OWNER/REPO/internal/config"
	"github.com/OWNER/REPO/internal/db"
	"github.com/OWNER/REPO/internal/logging"
	"github.com/OWNER/REPO/internal/server"
	"github.com/OWNER/REPO/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			printVersion()
			return
		case "migrate":
			if err := runMigrate(os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	if err := run(); err != nil {
		// Logger may not be installed yet on early failures; stderr is the
		// reliable channel for a fatal, readable message.
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// run loads config, sets up logging, connects the DB, and runs the server until
// shutdown.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger := logging.New(cfg.IsProd(), cfg.LogLevel)
	logger.Info("starting myapp",
		"version", version.Version,
		"commit", version.Commit,
		"env", string(cfg.Env),
		"addr", cfg.Addr(),
	)

	pool, err := db.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer pool.Close()

	srv, err := server.New(cfg, logger, pool)
	if err != nil {
		return fmt.Errorf("server: %w", err)
	}

	return srv.Run()
}

func printVersion() {
	fmt.Printf("myapp %s\n", version.String())
	fmt.Printf("  go:   %s\n", runtime.Version())
	fmt.Printf("  os:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
