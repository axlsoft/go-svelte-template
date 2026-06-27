// Package auth implements in-app OIDC (auth-code + PKCE), Postgres-backed
// server-side sessions, CSRF protection, a route guard and the /auth endpoints.
//
// Components: oidc.go (discovery, PKCE, ID-token verification), session.go
// (store-backed sessions + cookie + timeouts), csrf.go (signed double-submit),
// middleware.go (RequireAuth guard), handlers.go (login/callback/logout/me),
// module.go (wiring). Use NewModule to construct the stack.
package auth
