package main

import (
	"context"
	"fmt"

	"github.com/OWNER/REPO/internal/config"
	"github.com/OWNER/REPO/internal/db"
)

// runMigrate dispatches the `migrate` subcommand: up | down | status | create.
//
// Migrations are an explicit deploy step (`app migrate up`) — never run on boot.
func runMigrate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: myapp migrate <up|down|status|create> [name]")
	}
	command := args[0]

	var name string
	if command == "create" {
		if len(args) < 2 {
			return fmt.Errorf("usage: myapp migrate create <name>")
		}
		name = args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return db.Migrate(context.Background(), cfg.DatabaseURL, command, name)
}
