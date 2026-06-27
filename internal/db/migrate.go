package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	// pgx stdlib driver, registered as "pgx" for database/sql (goose).
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// migrationsFS embeds the goose .sql migrations so they ship inside the binary.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	migrationsDir = "migrations"
	// sourceMigrationsDir is the on-disk location used by `migrate create`
	// (a dev-time command run from the repo root).
	sourceMigrationsDir = "internal/db/migrations"
)

// Migrate runs a goose command against the embedded migrations.
//
// Supported commands: "up", "down", "status" (require a reachable DB) and
// "create" (writes a new migration file to the source tree; name required).
func Migrate(ctx context.Context, dsn, command, name string) error {
	if command == "create" {
		return createMigration(name)
	}

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	switch command {
	case "up":
		return goose.UpContext(ctx, sqlDB, migrationsDir)
	case "down":
		return goose.DownContext(ctx, sqlDB, migrationsDir)
	case "status":
		return goose.StatusContext(ctx, sqlDB, migrationsDir)
	default:
		return fmt.Errorf("unknown migrate command %q (want up|down|status|create)", command)
	}
}

// createMigration writes a new, empty goose SQL migration to the source tree,
// using a sequential, zero-padded version one higher than the current max.
func createMigration(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("migrate create: a migration name is required")
	}

	next, err := nextVersion(sourceMigrationsDir)
	if err != nil {
		return err
	}

	slug := slugify(name)
	filename := fmt.Sprintf("%05d_%s.sql", next, slug)
	full := filepath.Join(sourceMigrationsDir, filename)

	content := "-- +goose Up\n-- +goose StatementBegin\n\n-- +goose StatementEnd\n\n" +
		"-- +goose Down\n-- +goose StatementBegin\n\n-- +goose StatementEnd\n"

	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write migration: %w", err)
	}
	fmt.Printf("created %s\n", full)
	return nil
}

// nextVersion returns one higher than the largest NNNNN_ prefix in dir (or 1).
func nextVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir: %w", err)
	}

	versions := make([]int, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		prefix, _, ok := strings.Cut(e.Name(), "_")
		if !ok {
			continue
		}
		if n, err := strconv.Atoi(prefix); err == nil {
			versions = append(versions, n)
		}
	}

	if len(versions) == 0 {
		return 1, nil
	}
	sort.Ints(versions)
	return versions[len(versions)-1] + 1, nil
}

// slugify lowercases and replaces non-alphanumeric runs with underscores.
func slugify(s string) string {
	var b strings.Builder
	prevUnderscore := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevUnderscore = false
		default:
			if !prevUnderscore {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}
