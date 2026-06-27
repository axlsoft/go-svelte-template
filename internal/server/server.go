// Package server wires the HTTP server: chi router, timeouts and graceful
// shutdown with connection draining.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/OWNER/REPO/internal/config"
)

// Timeouts for the HTTP server. Conservative defaults that protect against slow
// clients while leaving room for normal SPA asset loads.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 20 * time.Second
)

// Server owns the http.Server and its lifecycle.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// New builds a Server: it constructs the router and wires timeouts. It does not
// start listening — call Run for that. pool may be nil (e.g. in tests); when
// provided it backs the /readyz DB check.
func New(cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) (*Server, error) {
	handler, err := NewRouter(cfg, logger, pool)
	if err != nil {
		return nil, err
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           handler,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		},
		logger: logger,
	}, nil
}

// Run starts the server and blocks until a SIGINT/SIGTERM is received, then
// gracefully shuts down, draining in-flight requests up to shutdownTimeout.
func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		s.logger.Info("http server listening", slog.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		// Listener failed to start (e.g. port in use).
		return err
	case <-ctx.Done():
		s.logger.Info("shutdown signal received, draining connections",
			slog.Duration("timeout", shutdownTimeout))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("graceful shutdown failed; forcing close", slog.Any("err", err))
		_ = s.httpServer.Close()
		return err
	}

	s.logger.Info("server stopped cleanly")
	return nil
}
