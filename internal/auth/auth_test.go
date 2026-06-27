package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testCSRF() *CSRF {
	return &CSRF{key: []byte("0123456789abcdef0123456789abcdef"), secure: false}
}

// csrfToken drives a GET through the middleware and returns the seeded token.
func csrfToken(t *testing.T, h http.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, c := range rec.Result().Cookies() {
		if c.Name == csrfCookieName {
			return c.Value
		}
	}
	t.Fatal("expected a CSRF cookie to be seeded on GET")
	return ""
}

func TestCSRFMiddleware(t *testing.T) {
	t.Parallel()

	c := testCSRF()
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := c.Middleware(ok)

	token := csrfToken(t, h)
	if !c.valid(token) {
		t.Fatal("seeded token should be valid")
	}

	t.Run("valid double-submit passes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
		req.Header.Set(CSRFHeader, token)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", rec.Code)
		}
	})

	t.Run("missing header is rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", rec.Code)
		}
	})

	t.Run("mismatched header is rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
		req.Header.Set(CSRFHeader, token+"x")
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", rec.Code)
		}
	})

	t.Run("tampered signature is rejected", func(t *testing.T) {
		forged := "forged-body.forged-sig"
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: forged})
		req.Header.Set(CSRFHeader, forged)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", rec.Code)
		}
	})
}

func TestSafeReturnTo(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"/dashboard":        "/dashboard",
		"/a/b?x=1":          "/a/b?x=1",
		"":                  "/",
		"//evil.com":        "/",
		"https://evil.com":  "/",
		"http://evil.com/x": "/",
		"/path://weird":     "/",
		"\\\\evil":          "/",
		"relative/no/slash": "/",
	}
	for in, want := range cases {
		if got := safeReturnTo(in); got != want {
			t.Errorf("safeReturnTo(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFlowCookieRoundTrip(t *testing.T) {
	t.Parallel()

	h := &Handler{flowKey: []byte("0123456789abcdef0123456789abcdef")}
	params := &FlowParams{State: "st", Nonce: "no", Verifier: "ve", ReturnTo: "/home"}

	rec := httptest.NewRecorder()
	if err := h.setFlowCookie(rec, params); err != nil {
		t.Fatalf("setFlowCookie: %v", err)
	}
	cookie := rec.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)
	req.AddCookie(cookie)
	got, err := h.readFlowCookie(req)
	if err != nil {
		t.Fatalf("readFlowCookie: %v", err)
	}
	if got.State != params.State || got.Nonce != params.Nonce ||
		got.Verifier != params.Verifier || got.ReturnTo != params.ReturnTo {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	t.Run("tampered flow cookie is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)
		req.AddCookie(&http.Cookie{Name: flowCookieName, Value: cookie.Value + "x"})
		if _, err := h.readFlowCookie(req); err == nil {
			t.Fatal("expected error for tampered flow cookie")
		}
	})
}
