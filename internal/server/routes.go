package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/OWNER/REPO/internal/auth"
	"github.com/OWNER/REPO/internal/config"
	"github.com/OWNER/REPO/internal/db"
	"github.com/OWNER/REPO/internal/health"
)

// Default per-IP rate limit. Generous enough for a SPA loading many assets in
// dev, while still bounding abusive clients.
//
// NOTE: in-memory and process-local; swap for a Redis-backed limiter when
// running more than one replica (see middleware.go).
const (
	rateLimitRPS   = 50
	rateLimitBurst = 100
)

// NewRouter builds the application router: global middleware, system/health
// endpoints, the (future) /api and /auth mounts, and the two-mode SPA handler
// as the catch-all. pool may be nil, in which case /readyz always reports ready.
func NewRouter(cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) (http.Handler, error) {
	spa, err := NewSPAHandler(cfg.IsDev(), cfg.ViteDevURL, logger)
	if err != nil {
		return nil, err
	}

	var readyCheck func(ctx context.Context) error
	if pool != nil {
		readyCheck = func(ctx context.Context) error { return db.Ping(ctx, pool) }
	}

	r := chi.NewRouter()

	// Global middleware (outermost first).
	r.Use(RequestID)
	r.Use(RequestLogger(logger))
	r.Use(Recoverer(logger))
	r.Use(SecureHeaders)
	r.Use(RateLimit(rateLimitRPS, rateLimitBurst))

	// System / health endpoints — never fall through to the SPA.
	r.Get("/healthz", health.Live)
	r.Get("/readyz", health.Ready(readyCheck))

	// Auth and the protected API are mounted only when a DB pool is available
	// (they need session storage). Without a pool (e.g. some tests) the server
	// still serves health and the SPA.
	if pool != nil {
		mountAuth(r, cfg, logger, pool)
	}

	// Catch-all: the two-mode SPA handler (dev proxy / prod embedded).
	r.NotFound(spa.ServeHTTP)
	r.MethodNotAllowed(spa.ServeHTTP)

	return r, nil
}

// mountAuth wires the /auth endpoints and a sample protected /api route guarded
// by the session and CSRF middleware.
func mountAuth(r chi.Router, cfg *config.Config, logger *slog.Logger, pool *pgxpool.Pool) {
	mod := auth.NewModule(pool, cfg, logger)

	r.Route("/auth", func(r chi.Router) {
		r.Use(mod.CSRF.Middleware)
		r.Get("/login", mod.Handler.Login)
		r.Get("/callback", mod.Handler.Callback)
		r.Post("/logout", mod.Handler.Logout)
		r.Get("/me", mod.Handler.Me)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(mod.CSRF.Middleware)
		r.Use(mod.Sessions.RequireAuth)

		// Sample protected endpoint: echoes the authenticated user. Demonstrates
		// the guard (401 without a session, 200 with one).
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok {
				WriteError(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(user)
		})
	})
}
