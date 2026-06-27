package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"

	"github.com/OWNER/REPO/internal/config"
	"github.com/OWNER/REPO/internal/store"
)

// ErrNoSession indicates there is no valid session for the request (missing,
// expired, or unknown cookie).
var ErrNoSession = errors.New("auth: no session")

// User is the authenticated identity surfaced to handlers and the frontend.
type User struct {
	Sub    string          `json:"sub"`
	Email  string          `json:"email,omitempty"`
	Name   string          `json:"name,omitempty"`
	Claims json.RawMessage `json:"claims,omitempty"`
	// SignedInSince is the session creation time. It is not part of the raw
	// claims; it comes from the session row and is used by the normalized
	// /auth/me identity. Zero for a freshly built (pre-persist) User.
	SignedInSince time.Time `json:"-"`
}

// SessionManager persists server-side sessions in Postgres and manages the
// session cookie, enforcing idle and absolute lifetimes.
type SessionManager struct {
	queries *store.Queries
	auth    *Authenticator
	logger  *slog.Logger

	cookieName   string
	cookieDomain string
	secure       bool
	idle         time.Duration
	absolute     time.Duration
}

// NewSessionManager builds a SessionManager. secure should be true in prod so
// the cookie carries the Secure attribute.
func NewSessionManager(q *store.Queries, a *Authenticator, cfg *config.Config, logger *slog.Logger) *SessionManager {
	return &SessionManager{
		queries:      q,
		auth:         a,
		logger:       logger,
		cookieName:   cfg.SessionCookieName,
		cookieDomain: cfg.SessionCookieDomain,
		secure:       cfg.IsProd(),
		idle:         cfg.SessionIdleTimeout,
		absolute:     cfg.SessionAbsoluteTimeout,
	}
}

// Create persists a new session for user, stores the refresh token, and sets the
// session cookie.
func (m *SessionManager) Create(ctx context.Context, w http.ResponseWriter, user *User, refreshToken string) error {
	id, err := randomToken()
	if err != nil {
		return err
	}

	claims := user.Claims
	if len(claims) == 0 {
		claims = json.RawMessage("{}")
	}

	expires := time.Now().Add(m.absolute)
	if _, err := m.queries.CreateSession(ctx, store.CreateSessionParams{
		ID:           id,
		UserID:       user.Sub,
		UserEmail:    user.Email,
		UserName:     user.Name,
		Claims:       claims,
		RefreshToken: refreshToken,
		ExpiresAt:    pgtype.Timestamptz{Time: expires, Valid: true},
	}); err != nil {
		return err
	}

	m.setCookie(w, id)
	return nil
}

// Resolve looks up the session referenced by the request cookie, enforcing the
// absolute and idle timeouts. On success it slides the idle window (touches
// last_seen_at) and returns the user. Expired/unknown sessions are deleted and
// ErrNoSession is returned.
func (m *SessionManager) Resolve(ctx context.Context, r *http.Request) (*User, error) {
	c, err := r.Cookie(m.cookieName)
	if err != nil || c.Value == "" {
		return nil, ErrNoSession
	}

	s, err := m.queries.GetSession(ctx, c.Value)
	if err != nil {
		return nil, ErrNoSession
	}

	now := time.Now()
	if s.ExpiresAt.Valid && now.After(s.ExpiresAt.Time) {
		_ = m.queries.DeleteSession(ctx, s.ID)
		return nil, ErrNoSession
	}
	if s.LastSeenAt.Valid && now.After(s.LastSeenAt.Time.Add(m.idle)) {
		_ = m.queries.DeleteSession(ctx, s.ID)
		return nil, ErrNoSession
	}

	if _, err := m.queries.TouchSession(ctx, s.ID); err != nil {
		// Touch failure is non-fatal for this request; log and continue.
		m.logger.WarnContext(ctx, "touch session failed", slog.Any("err", err))
	}

	return userFromSession(s), nil
}

// Destroy deletes the session referenced by the request cookie (if any) and
// clears the cookie. It returns the deleted session's refresh token so the
// caller can attempt RP-initiated logout, or "" when there was no session.
func (m *SessionManager) Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(m.cookieName); err == nil && c.Value != "" {
		if err := m.queries.DeleteSession(ctx, c.Value); err != nil {
			m.logger.WarnContext(ctx, "delete session failed", slog.Any("err", err))
		}
	}
	m.clearCookie(w)
}

// RefreshSession uses the stored refresh token to obtain a fresh access token
// for downstream API calls. When the IdP rotates the refresh token, the session
// id and refresh token are rotated and the cookie is reissued. Call this when an
// access token is needed and may have expired (silent refresh).
func (m *SessionManager) RefreshSession(ctx context.Context, w http.ResponseWriter, r *http.Request) (*oauth2.Token, error) {
	c, err := r.Cookie(m.cookieName)
	if err != nil || c.Value == "" {
		return nil, ErrNoSession
	}
	s, err := m.queries.GetSession(ctx, c.Value)
	if err != nil {
		return nil, ErrNoSession
	}
	if s.RefreshToken == "" {
		return nil, errors.New("auth: session has no refresh token")
	}

	src, err := m.auth.TokenSource(ctx, s.RefreshToken)
	if err != nil {
		return nil, err
	}
	tok, err := src.Token()
	if err != nil {
		return nil, err
	}

	if tok.RefreshToken != "" && tok.RefreshToken != s.RefreshToken {
		newID, err := randomToken()
		if err != nil {
			return nil, err
		}
		expires := time.Now().Add(m.absolute)
		if _, err := m.queries.RotateSession(ctx, store.RotateSessionParams{
			NewID:        newID,
			RefreshToken: tok.RefreshToken,
			ExpiresAt:    pgtype.Timestamptz{Time: expires, Valid: true},
			OldID:        s.ID,
		}); err != nil {
			return nil, err
		}
		m.setCookie(w, newID)
	}

	return tok, nil
}

// setCookie writes the session cookie with the absolute lifetime as max-age.
func (m *SessionManager) setCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    id,
		Path:     "/",
		Domain:   m.cookieDomain,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.absolute.Seconds()),
	})
}

// clearCookie expires the session cookie on the client.
func (m *SessionManager) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.cookieDomain,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// userFromSession projects a stored session row into the public User identity.
func userFromSession(s store.Session) *User {
	u := &User{
		Sub:    s.UserID,
		Email:  s.UserEmail,
		Name:   s.UserName,
		Claims: json.RawMessage(s.Claims),
	}
	if s.CreatedAt.Valid {
		u.SignedInSince = s.CreatedAt.Time
	}
	return u
}
