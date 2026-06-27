package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/OWNER/REPO/internal/config"
)

// Authenticator wraps OIDC discovery, the OAuth2 auth-code + PKCE flow, and
// ID-token verification.
//
// Provider discovery (the IdP's .well-known endpoint) is performed lazily on
// first use and cached, so the app boots even if the IdP is briefly
// unreachable; discovery is retried until it succeeds.
type Authenticator struct {
	issuer       string
	clientID     string
	clientSecret string
	redirectURL  string
	scopes       []string

	mu       sync.Mutex
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth2   *oauth2.Config
}

// NewAuthenticator builds an Authenticator from config. It does not contact the
// IdP — discovery happens on first use.
func NewAuthenticator(cfg *config.Config) *Authenticator {
	return &Authenticator{
		issuer:       cfg.OIDCIssuer,
		clientID:     cfg.OIDCClientID,
		clientSecret: cfg.OIDCClientSecret,
		redirectURL:  cfg.OIDCRedirectURL,
		scopes:       cfg.OIDCScopes,
	}
}

// ensure performs (or reuses) provider discovery and builds the verifier and
// oauth2 config. Safe for concurrent use.
func (a *Authenticator) ensure(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.provider != nil {
		return nil
	}

	provider, err := oidc.NewProvider(ctx, a.issuer)
	if err != nil {
		return fmt.Errorf("oidc discovery (%s): %w", a.issuer, err)
	}

	a.provider = provider
	a.verifier = provider.Verifier(&oidc.Config{ClientID: a.clientID})
	a.oauth2 = &oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  a.redirectURL,
		Scopes:       a.scopes,
	}
	return nil
}

// FlowParams are the per-login secrets that must round-trip from /auth/login to
// /auth/callback (carried in a signed, short-lived cookie).
type FlowParams struct {
	State    string `json:"state"`
	Nonce    string `json:"nonce"`
	Verifier string `json:"verifier"`
	ReturnTo string `json:"return_to"`
}

// AuthCodeURL generates fresh state, nonce and a PKCE verifier, and returns the
// IdP authorization URL plus the FlowParams to persist for the callback.
func (a *Authenticator) AuthCodeURL(ctx context.Context, returnTo string) (string, *FlowParams, error) {
	if err := a.ensure(ctx); err != nil {
		return "", nil, err
	}

	state, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	nonce, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	verifier := oauth2.GenerateVerifier()

	url := a.oauth2.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(verifier),
	)

	return url, &FlowParams{
		State:    state,
		Nonce:    nonce,
		Verifier: verifier,
		ReturnTo: returnTo,
	}, nil
}

// Exchange swaps the authorization code for a token set, supplying the PKCE
// verifier.
func (a *Authenticator) Exchange(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	if err := a.ensure(ctx); err != nil {
		return nil, err
	}
	return a.oauth2.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}

// VerifyIDToken verifies the raw ID token from a token set and checks the nonce,
// returning the parsed token (claims available via Claims).
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token, nonce string) (*oidc.IDToken, error) {
	if err := a.ensure(ctx); err != nil {
		return nil, err
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("oidc: token response missing id_token")
	}

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc: verify id_token: %w", err)
	}
	if idToken.Nonce != nonce {
		return nil, fmt.Errorf("oidc: nonce mismatch")
	}
	return idToken, nil
}

// TokenSource returns an oauth2 TokenSource seeded with a refresh token, used to
// silently refresh access tokens. The source rotates the refresh token when the
// IdP issues a new one.
func (a *Authenticator) TokenSource(ctx context.Context, refreshToken string) (oauth2.TokenSource, error) {
	if err := a.ensure(ctx); err != nil {
		return nil, err
	}
	return a.oauth2.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}), nil
}

// LogoutURL builds the IdP RP-initiated logout URL when the provider advertises
// an end_session_endpoint. The bool is false when the provider does not support
// it, in which case the caller should fall back to a local logout.
func (a *Authenticator) LogoutURL(ctx context.Context, idTokenHint, postLogoutRedirectURL string) (string, bool) {
	if err := a.ensure(ctx); err != nil {
		return "", false
	}

	var meta struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := a.provider.Claims(&meta); err != nil || meta.EndSessionEndpoint == "" {
		return "", false
	}

	u := meta.EndSessionEndpoint
	sep := "?"
	add := func(k, v string) {
		if v == "" {
			return
		}
		u += sep + k + "=" + url.QueryEscape(v)
		sep = "&"
	}
	add("post_logout_redirect_uri", postLogoutRedirectURL)
	add("id_token_hint", idTokenHint)
	add("client_id", a.clientID)
	return u, true
}

// randomToken returns a URL-safe, 256-bit random token.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
