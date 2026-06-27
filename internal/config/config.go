// Package config holds the typed, env-sourced application configuration.
//
// Config is loaded from the environment (prefix MYAPP_) and validated on boot:
// missing or invalid values cause Load to return an error so main can log a
// clear message and exit non-zero (fail fast). Later phases extend Config with
// DB, OIDC, session and CSRF fields.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// envPrefix is prepended to every environment variable name.
//
// NOTE: bootstrap.sh rewrites MYAPP_ to the project's prefix; keep this literal
// in sync with that placeholder.
const envPrefix = "MYAPP_"

// Env is the application environment.
type Env string

const (
	// EnvDev is local development (text logs, Vite proxy enabled).
	EnvDev Env = "dev"
	// EnvProd is production (JSON logs, embedded SPA served).
	EnvProd Env = "prod"
)

// Config is the typed application configuration.
//
// Fields are populated from the environment by Load and validated by validate.
// Later phases append fields (DB DSN, OIDC, sessions, CSRF) — keep this struct
// the single source of truth.
type Config struct {
	// Env selects dev vs prod behaviour.
	Env Env
	// HTTPHost is the interface the server binds (e.g. 127.0.0.1).
	HTTPHost string
	// HTTPPort is the TCP port the server binds.
	HTTPPort int
	// LogLevel is one of debug|info|warn|error.
	LogLevel string
	// ViteDevURL is the Vite dev-server base URL that the SPA handler proxies to
	// in dev. Required in dev, ignored in prod.
	ViteDevURL string
	// DatabaseURL is the Postgres DSN (postgres://user:pass@host:port/db?...).
	DatabaseURL string

	// --- OIDC ---------------------------------------------------------------

	// OIDCIssuer is the IdP issuer URL used for discovery (.well-known).
	OIDCIssuer string
	// OIDCClientID is this app's OAuth2 client id.
	OIDCClientID string
	// OIDCClientSecret is this app's OAuth2 client secret.
	OIDCClientSecret string
	// OIDCRedirectURL is the absolute callback URL registered with the IdP.
	OIDCRedirectURL string
	// OIDCScopes are the OAuth2 scopes requested at login (always includes
	// openid).
	OIDCScopes []string
	// OIDCPostLogoutRedirectURL is where the IdP returns the user after an
	// RP-initiated logout.
	OIDCPostLogoutRedirectURL string
	// OIDCRolesClaim is the dot-path into the ID-token claims where roles live
	// (e.g. "realm_access.roles" or "roles"). The value is tolerated as either
	// an array or a single string (some IdPs emit a scalar for a single role).
	OIDCRolesClaim string
	// OIDCAdminRole is the role name that grants admin (drives is_admin).
	OIDCAdminRole string

	// --- Sessions -----------------------------------------------------------

	// SessionCookieName is the name of the session id cookie.
	SessionCookieName string
	// SessionCookieDomain optionally scopes the session cookie to a domain
	// (empty = host-only).
	SessionCookieDomain string
	// SessionIdleTimeout expires a session after this much inactivity.
	SessionIdleTimeout time.Duration
	// SessionAbsoluteTimeout is the maximum lifetime of a session regardless of
	// activity.
	SessionAbsoluteTimeout time.Duration

	// --- CSRF ---------------------------------------------------------------

	// CSRFSecret is the HMAC key used to sign CSRF tokens and the short-lived
	// OIDC login-flow cookie. Must be at least 32 bytes.
	CSRFSecret string
}

// Addr returns the host:port the HTTP server should bind.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// IsDev reports whether the app is running in the dev environment.
func (c *Config) IsDev() bool { return c.Env == EnvDev }

// IsProd reports whether the app is running in the prod environment.
func (c *Config) IsProd() bool { return c.Env == EnvProd }

// Load reads configuration from the environment, applies defaults, and
// validates it. A non-nil error means the process should exit non-zero.
func Load() (*Config, error) {
	cfg := &Config{
		Env:         Env(getenv("ENV", string(EnvDev))),
		HTTPHost:    getenv("HTTP_HOST", "127.0.0.1"),
		LogLevel:    strings.ToLower(getenv("LOG_LEVEL", "info")),
		ViteDevURL:  getenv("VITE_DEV_URL", "http://127.0.0.1:5173"),
		DatabaseURL: getenv("DB_DSN", "postgres://myapp:myapp@127.0.0.1:5432/myapp?sslmode=disable"),

		OIDCIssuer:                getenv("OIDC_ISSUER", "http://127.0.0.1:5556/dex"),
		OIDCClientID:              getenv("OIDC_CLIENT_ID", "myapp"),
		OIDCClientSecret:          getenv("OIDC_CLIENT_SECRET", "myapp-dev-secret-change-me"),
		OIDCRedirectURL:           getenv("OIDC_REDIRECT_URL", "http://127.0.0.1:8080/auth/callback"),
		OIDCScopes:                splitScopes(getenv("OIDC_SCOPES", "openid,profile,email")),
		OIDCPostLogoutRedirectURL: getenv("OIDC_POST_LOGOUT_REDIRECT_URL", "http://127.0.0.1:8080/"),
		OIDCRolesClaim:            getenv("OIDC_ROLES_CLAIM", "realm_access.roles"),
		OIDCAdminRole:             getenv("OIDC_ADMIN_ROLE", "admin"),

		SessionCookieName:   getenv("SESSION_COOKIE_NAME", "myapp_session"),
		SessionCookieDomain: getenv("SESSION_COOKIE_DOMAIN", ""),

		// NOTE: dev default; override with a real random value in prod.
		CSRFSecret: getenv("CSRF_SECRET", "dev-csrf-secret-change-me-0123456789ab"),
	}

	port, err := getenvInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, err
	}
	cfg.HTTPPort = port

	idle, err := getenvDuration("SESSION_IDLE_TIMEOUT", 30*time.Minute)
	if err != nil {
		return nil, err
	}
	cfg.SessionIdleTimeout = idle

	absolute, err := getenvDuration("SESSION_ABSOLUTE_TIMEOUT", 12*time.Hour)
	if err != nil {
		return nil, err
	}
	cfg.SessionAbsoluteTimeout = absolute

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate enforces invariants. Every failure names the offending env var so
// the operator can fix it immediately.
func (c *Config) validate() error {
	switch c.Env {
	case EnvDev, EnvProd:
	default:
		return fmt.Errorf("%sENV must be %q or %q, got %q", envPrefix, EnvDev, EnvProd, c.Env)
	}

	if c.HTTPHost == "" {
		return fmt.Errorf("%sHTTP_HOST must not be empty", envPrefix)
	}

	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		return fmt.Errorf("%sHTTP_PORT must be 1..65535, got %d", envPrefix, c.HTTPPort)
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("%sLOG_LEVEL must be debug|info|warn|error, got %q", envPrefix, c.LogLevel)
	}

	// The Vite dev-server URL is only meaningful (and required) in dev.
	if c.IsDev() {
		if c.ViteDevURL == "" {
			return fmt.Errorf("%sVITE_DEV_URL must be set in dev", envPrefix)
		}
		u, err := url.Parse(c.ViteDevURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("%sVITE_DEV_URL must be an absolute URL, got %q", envPrefix, c.ViteDevURL)
		}
	}

	if c.DatabaseURL == "" {
		return fmt.Errorf("%sDB_DSN must not be empty", envPrefix)
	}
	if u, err := url.Parse(c.DatabaseURL); err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return fmt.Errorf("%sDB_DSN must be a postgres:// DSN, got %q", envPrefix, c.DatabaseURL)
	}

	if err := c.validateAuth(); err != nil {
		return err
	}

	return nil
}

// validateAuth enforces the OIDC, session and CSRF invariants.
func (c *Config) validateAuth() error {
	if err := requireAbsURL("OIDC_ISSUER", c.OIDCIssuer); err != nil {
		return err
	}
	if c.OIDCClientID == "" {
		return fmt.Errorf("%sOIDC_CLIENT_ID must not be empty", envPrefix)
	}
	if c.OIDCClientSecret == "" {
		return fmt.Errorf("%sOIDC_CLIENT_SECRET must not be empty", envPrefix)
	}
	if err := requireAbsURL("OIDC_REDIRECT_URL", c.OIDCRedirectURL); err != nil {
		return err
	}
	if err := requireAbsURL("OIDC_POST_LOGOUT_REDIRECT_URL", c.OIDCPostLogoutRedirectURL); err != nil {
		return err
	}
	if len(c.OIDCScopes) == 0 {
		return fmt.Errorf("%sOIDC_SCOPES must include at least openid", envPrefix)
	}

	if c.SessionCookieName == "" {
		return fmt.Errorf("%sSESSION_COOKIE_NAME must not be empty", envPrefix)
	}
	if c.SessionIdleTimeout <= 0 {
		return fmt.Errorf("%sSESSION_IDLE_TIMEOUT must be positive", envPrefix)
	}
	if c.SessionAbsoluteTimeout <= 0 {
		return fmt.Errorf("%sSESSION_ABSOLUTE_TIMEOUT must be positive", envPrefix)
	}
	if c.SessionIdleTimeout > c.SessionAbsoluteTimeout {
		return fmt.Errorf("%sSESSION_IDLE_TIMEOUT must be <= SESSION_ABSOLUTE_TIMEOUT", envPrefix)
	}

	if len(c.CSRFSecret) < 32 {
		return fmt.Errorf("%sCSRF_SECRET must be at least 32 bytes", envPrefix)
	}

	return nil
}

// requireAbsURL validates that v is a non-empty absolute URL, naming the env var
// on failure.
func requireAbsURL(key, v string) error {
	if v == "" {
		return fmt.Errorf("%s%s must not be empty", envPrefix, key)
	}
	u, err := url.Parse(v)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%s%s must be an absolute URL, got %q", envPrefix, key, v)
	}
	return nil
}

// getenv returns the prefixed env var or a default.
func getenv(key, def string) string {
	if v, ok := os.LookupEnv(envPrefix + key); ok {
		return v
	}
	return def
}

// getenvInt returns the prefixed env var parsed as an int, or a default.
func getenvInt(key string, def int) (int, error) {
	v, ok := os.LookupEnv(envPrefix + key)
	if !ok {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s%s must be an integer, got %q", envPrefix, key, v)
	}
	return n, nil
}

// getenvDuration returns the prefixed env var parsed as a Go duration, or a
// default.
func getenvDuration(key string, def time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(envPrefix + key)
	if !ok {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s%s must be a duration (e.g. 30m, 12h), got %q", envPrefix, key, v)
	}
	return d, nil
}

// splitScopes parses a comma-separated scope list, trimming blanks and ensuring
// the openid scope is always present.
func splitScopes(s string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	if !seen["openid"] {
		out = append([]string{"openid"}, out...)
	}
	return out
}
