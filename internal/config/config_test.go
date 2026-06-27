package config

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	base := func() *Config {
		return &Config{
			Env:         EnvProd,
			HTTPHost:    "127.0.0.1",
			HTTPPort:    8080,
			LogLevel:    "info",
			ViteDevURL:  "",
			DatabaseURL: "postgres://myapp:myapp@127.0.0.1:5432/myapp?sslmode=disable",

			OIDCIssuer:                "http://127.0.0.1:5556/dex",
			OIDCClientID:              "myapp",
			OIDCClientSecret:          "secret",
			OIDCRedirectURL:           "http://127.0.0.1:8080/auth/callback",
			OIDCScopes:                []string{"openid", "profile", "email"},
			OIDCPostLogoutRedirectURL: "http://127.0.0.1:8080/",

			SessionCookieName:      "myapp_session",
			SessionIdleTimeout:     30 * time.Minute,
			SessionAbsoluteTimeout: 12 * time.Hour,

			CSRFSecret: "dev-csrf-secret-change-me-0123456789ab",
		}
	}

	t.Run("valid prod config", func(t *testing.T) {
		if err := base().validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("dev requires vite url", func(t *testing.T) {
		c := base()
		c.Env = EnvDev
		c.ViteDevURL = ""
		if err := c.validate(); err == nil {
			t.Fatal("expected error for missing VITE_DEV_URL in dev")
		}
	})

	t.Run("invalid env", func(t *testing.T) {
		c := base()
		c.Env = "staging"
		if err := c.validate(); err == nil {
			t.Fatal("expected error for invalid env")
		}
	})

	t.Run("port out of range", func(t *testing.T) {
		c := base()
		c.HTTPPort = 70000
		if err := c.validate(); err == nil {
			t.Fatal("expected error for out-of-range port")
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		c := base()
		c.LogLevel = "trace"
		if err := c.validate(); err == nil {
			t.Fatal("expected error for invalid log level")
		}
	})

	t.Run("invalid database dsn", func(t *testing.T) {
		c := base()
		c.DatabaseURL = "mysql://nope"
		if err := c.validate(); err == nil {
			t.Fatal("expected error for non-postgres DSN")
		}
	})

	t.Run("short csrf secret", func(t *testing.T) {
		c := base()
		c.CSRFSecret = "too-short"
		if err := c.validate(); err == nil {
			t.Fatal("expected error for short CSRF secret")
		}
	})

	t.Run("invalid oidc issuer", func(t *testing.T) {
		c := base()
		c.OIDCIssuer = "not-a-url"
		if err := c.validate(); err == nil {
			t.Fatal("expected error for non-absolute OIDC issuer")
		}
	})

	t.Run("idle exceeds absolute", func(t *testing.T) {
		c := base()
		c.SessionIdleTimeout = 24 * time.Hour
		c.SessionAbsoluteTimeout = 12 * time.Hour
		if err := c.validate(); err == nil {
			t.Fatal("expected error when idle timeout exceeds absolute")
		}
	})
}
