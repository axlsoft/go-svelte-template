package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/OWNER/REPO/internal/config"
)

// csrfCookieName is the cookie carrying the (JS-readable) CSRF token.
const csrfCookieName = "myapp_csrf"

// CSRFHeader is the request header the SPA echoes the CSRF token in.
const CSRFHeader = "X-CSRF-Token"

// CSRF implements signed double-submit-cookie CSRF protection.
//
// On safe requests the middleware ensures a signed token cookie exists (the SPA
// reads it and echoes it in the X-CSRF-Token header). On state-changing
// requests it requires the header to equal the cookie and the signature to be
// valid; otherwise it replies 403.
type CSRF struct {
	key    []byte
	secure bool
}

// NewCSRF builds a CSRF protector keyed by the configured secret.
func NewCSRF(cfg *config.Config) *CSRF {
	return &CSRF{
		key:    []byte(cfg.CSRFSecret),
		secure: cfg.IsProd(),
	}
}

// Middleware enforces CSRF on unsafe methods and seeds the token cookie on safe
// ones.
func (c *CSRF) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) {
			c.ensureCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(csrfCookieName)
		if err != nil || cookie.Value == "" {
			writeJSONError(w, http.StatusForbidden, "csrf_missing", "missing CSRF cookie")
			return
		}
		header := r.Header.Get(CSRFHeader)
		if header == "" {
			writeJSONError(w, http.StatusForbidden, "csrf_missing", "missing CSRF header")
			return
		}
		if subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 || !c.valid(cookie.Value) {
			writeJSONError(w, http.StatusForbidden, "csrf_invalid", "invalid CSRF token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ensureCookie issues a fresh signed token cookie when none (or an invalid one)
// is present.
func (c *CSRF) ensureCookie(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(csrfCookieName); err == nil && c.valid(cookie.Value) {
		return
	}
	token, err := c.issue()
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // must be readable by the SPA to echo in the header
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// issue creates a new signed token: base64(random).base64(hmac(random)).
func (c *CSRF) issue() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(b)
	return body + "." + c.sign(body), nil
}

// valid reports whether token is well-formed and carries a valid signature.
func (c *CSRF) valid(token string) bool {
	body, sig, ok := strings.Cut(token, ".")
	if !ok || body == "" || sig == "" {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(c.sign(body)))
}

// sign returns the base64 HMAC-SHA256 of body under the CSRF key.
func (c *CSRF) sign(body string) string {
	mac := hmac.New(sha256.New, c.key)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// isSafeMethod reports whether m is a read-only HTTP method exempt from CSRF.
func isSafeMethod(m string) bool {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
