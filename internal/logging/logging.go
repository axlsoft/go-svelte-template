// Package logging configures the application's slog logger.
//
// In prod the handler emits JSON; in dev it emits human-readable text. The level
// comes from config. New returns the logger and also installs it as the slog
// default so package-level slog calls share the same configuration.
package logging

import (
	"log/slog"
	"os"
)

// New builds a *slog.Logger for the given environment and level string, installs
// it as the slog default, and returns it.
//
// level is one of debug|info|warn|error (already validated by config); unknown
// values fall back to info. json selects the JSON handler (prod) over text (dev).
func New(json bool, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var h slog.Handler
	if json {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(h)
	slog.SetDefault(logger)
	return logger
}

// parseLevel maps a level string to a slog.Level, defaulting to info.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
