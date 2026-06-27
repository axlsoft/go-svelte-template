package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsReservedPath(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"/api":             true,
		"/api/":            true,
		"/api/users":       true,
		"/auth":            true,
		"/auth/callback":   true,
		"/healthz":         true,
		"/readyz":          true,
		"/":                false,
		"/dashboard":       false,
		"/login":           false,
		"/apixyz":          false, // not a prefix boundary
		"/healthzz":        false,
		"/assets/app.js":   false,
		"/auth-not-really": false,
	}

	for path, want := range cases {
		if got := IsReservedPath(path); got != want {
			t.Errorf("IsReservedPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// discardLogger returns a logger that drops output, for tests.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestProdSPA_RouteClassification(t *testing.T) {
	t.Parallel()

	h, err := NewSPAHandler(false /* prod */, "", discardLogger())
	if err != nil {
		t.Fatalf("NewSPAHandler: %v", err)
	}

	cases := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantHTML   bool // SPA fallback serves text/html
	}{
		{"root serves SPA", http.MethodGet, "/", http.StatusOK, true},
		{"unknown client route falls back to SPA", http.MethodGet, "/dashboard/deep", http.StatusOK, true},
		{"reserved API path never hits SPA", http.MethodGet, "/api/users", http.StatusNotFound, false},
		{"reserved auth path never hits SPA", http.MethodGet, "/auth/login", http.StatusNotFound, false},
		{"reserved health path never hits SPA", http.MethodGet, "/healthz", http.StatusNotFound, false},
		{"non-GET on SPA route is rejected", http.MethodPost, "/dashboard", http.StatusMethodNotAllowed, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("%s %s: status = %d, want %d", tc.method, tc.path, rec.Code, tc.wantStatus)
			}
			if tc.wantHTML {
				if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
					t.Errorf("%s %s: Content-Type = %q, want text/html", tc.method, tc.path, ct)
				}
			}
		})
	}
}

func TestDevProxy_ReservedPathNotProxied(t *testing.T) {
	t.Parallel()

	// Point at an unroutable target; reserved paths must 404 without ever
	// attempting the proxy (a proxy attempt would yield 502).
	h, err := NewSPAHandler(true /* dev */, "http://127.0.0.1:1", discardLogger())
	if err != nil {
		t.Fatalf("NewSPAHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("reserved path in dev: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
