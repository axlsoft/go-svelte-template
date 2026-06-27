package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/OWNER/REPO/internal/config"
)

// flowCookieName carries the signed, short-lived OIDC login-flow state from
// /auth/login to /auth/callback. Scoped to /auth so it is only sent on the
// callback.
const flowCookieName = "myapp_oidc_flow"

// Handler serves the /auth endpoints: login, callback, logout, me.
type Handler struct {
	auth                  *Authenticator
	sessions              *SessionManager
	logger                *slog.Logger
	flowKey               []byte
	postLogoutRedirectURL string
	rolesClaim            string
	adminRole             string
	secure                bool
}

// NewHandler builds the /auth handler set.
func NewHandler(a *Authenticator, sessions *SessionManager, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		auth:                  a,
		sessions:              sessions,
		logger:                logger,
		flowKey:               []byte(cfg.CSRFSecret),
		postLogoutRedirectURL: cfg.OIDCPostLogoutRedirectURL,
		rolesClaim:            cfg.OIDCRolesClaim,
		adminRole:             cfg.OIDCAdminRole,
		secure:                cfg.IsProd(),
	}
}

// Login starts the auth-code + PKCE flow: it generates state/nonce/verifier,
// stores them in a signed flow cookie, and redirects to the IdP.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	returnTo := safeReturnTo(r.URL.Query().Get("return_to"))

	url, params, err := h.auth.AuthCodeURL(r.Context(), returnTo)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "auth code url", slog.Any("err", err))
		writeJSONError(w, http.StatusBadGateway, "idp_unreachable", "cannot reach the identity provider")
		return
	}

	if err := h.setFlowCookie(w, params); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal_error", "could not start login")
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

// Callback completes the flow: validate state, exchange the code, verify the ID
// token (and nonce), create a session, set the cookie, and redirect into the app.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	flow, err := h.readFlowCookie(r)
	h.clearFlowCookie(w)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid_state", "missing or invalid login state")
		return
	}

	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		writeJSONError(w, http.StatusUnauthorized, "idp_error", e)
		return
	}
	if q.Get("state") != flow.State {
		writeJSONError(w, http.StatusForbidden, "state_mismatch", "state parameter mismatch")
		return
	}
	code := q.Get("code")
	if code == "" {
		writeJSONError(w, http.StatusBadRequest, "missing_code", "missing authorization code")
		return
	}

	token, err := h.auth.Exchange(r.Context(), code, flow.Verifier)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "code exchange", slog.Any("err", err))
		writeJSONError(w, http.StatusBadGateway, "exchange_failed", "could not exchange authorization code")
		return
	}

	idToken, err := h.auth.VerifyIDToken(r.Context(), token, flow.Nonce)
	if err != nil {
		h.logger.WarnContext(r.Context(), "id token verification", slog.Any("err", err))
		writeJSONError(w, http.StatusUnauthorized, "invalid_id_token", "could not verify id token")
		return
	}

	user, err := userFromIDToken(idToken)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "claims_error", "could not read identity claims")
		return
	}

	if err := h.sessions.Create(r.Context(), w, user, token.RefreshToken); err != nil {
		h.logger.ErrorContext(r.Context(), "create session", slog.Any("err", err))
		writeJSONError(w, http.StatusInternalServerError, "session_error", "could not create session")
		return
	}

	http.Redirect(w, r, flow.ReturnTo, http.StatusFound)
}

// logoutResponse tells the SPA where to navigate to finish logout at the IdP.
type logoutResponse struct {
	LogoutURL string `json:"logout_url"`
}

// Logout clears the local session and cookie, then returns the IdP RP-initiated
// logout URL (or the post-logout redirect) for the SPA to navigate to.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Destroy(r.Context(), w, r)

	logoutURL := h.postLogoutRedirectURL
	if u, ok := h.auth.LogoutURL(r.Context(), "", h.postLogoutRedirectURL); ok {
		logoutURL = u
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(logoutResponse{LogoutURL: logoutURL})
}

// Me returns the current user as a normalized identity for frontend hydration,
// or 401 when there is no session. All claim parsing (roles, names, picture)
// happens here in Go so the client only ever sees the clean shape.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.sessions.Resolve(r.Context(), r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(NewIdentity(user, h.rolesClaim, h.adminRole))
}

// --- flow cookie (signed) ---------------------------------------------------

func (h *Handler) setFlowCookie(w http.ResponseWriter, params *FlowParams) error {
	payload, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	value := body + "." + h.signFlow(body)
	http.SetCookie(w, &http.Cookie{
		Name:     flowCookieName,
		Value:    value,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes to complete the login
	})
	return nil
}

func (h *Handler) readFlowCookie(r *http.Request) (*FlowParams, error) {
	c, err := r.Cookie(flowCookieName)
	if err != nil || c.Value == "" {
		return nil, errInvalidFlow
	}
	body, sig, ok := strings.Cut(c.Value, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(h.signFlow(body))) {
		return nil, errInvalidFlow
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil, errInvalidFlow
	}
	var params FlowParams
	if err := json.Unmarshal(payload, &params); err != nil {
		return nil, errInvalidFlow
	}
	params.ReturnTo = safeReturnTo(params.ReturnTo)
	return &params, nil
}

func (h *Handler) clearFlowCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     flowCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (h *Handler) signFlow(body string) string {
	mac := hmac.New(sha256.New, h.flowKey)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// --- helpers ----------------------------------------------------------------

// errInvalidFlow is returned when the login-flow cookie is missing/tampered.
var errInvalidFlow = &authError{"invalid login flow"}

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }

// safeReturnTo restricts the post-login redirect to a local, absolute path to
// prevent open redirects. Anything else collapses to "/".
func safeReturnTo(p string) string {
	if p == "" || !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		return "/"
	}
	if strings.Contains(p, "://") || strings.Contains(p, "\\") {
		return "/"
	}
	return p
}

// userFromIDToken projects verified ID-token claims into a User, preserving the
// raw claims JSON.
func userFromIDToken(idToken *oidc.IDToken) (*User, error) {
	var raw json.RawMessage
	if err := idToken.Claims(&raw); err != nil {
		return nil, err
	}

	var c struct {
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	_ = json.Unmarshal(raw, &c)

	name := c.Name
	if name == "" {
		name = c.PreferredUsername
	}

	return &User{
		Sub:    idToken.Subject,
		Email:  c.Email,
		Name:   name,
		Claims: raw,
	}, nil
}
