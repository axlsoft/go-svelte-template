package auth

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/OWNER/REPO/internal/config"
	"github.com/OWNER/REPO/internal/store"
)

// Module bundles the wired auth components for the router to mount.
type Module struct {
	Handler  *Handler
	Sessions *SessionManager
	CSRF     *CSRF
}

// NewModule constructs the auth stack (OIDC authenticator, session manager, CSRF
// protector and /auth handlers) over the given connection pool.
func NewModule(pool *pgxpool.Pool, cfg *config.Config, logger *slog.Logger) *Module {
	q := store.New(pool)
	authenticator := NewAuthenticator(cfg)
	sessions := NewSessionManager(q, authenticator, cfg, logger)
	csrf := NewCSRF(cfg)
	handler := NewHandler(authenticator, sessions, cfg, logger)

	return &Module{
		Handler:  handler,
		Sessions: sessions,
		CSRF:     csrf,
	}
}
