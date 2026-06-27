# Security Policy

## Supported versions

Only the latest commit on `main` is actively maintained. This is a template
repository — apply security fixes to your own project once you have stamped it.

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report vulnerabilities privately via
[GitHub Security Advisories](../../security/advisories/new). Include:

- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof of concept
- The Go version, Node version, and OS (if relevant)

You can expect an acknowledgement within 72 hours and a status update within
7 days. If the report is confirmed, a fix will be prepared and a security
advisory published before or alongside disclosure.

## Scope

This is a template. The following areas are in scope for this repository:

- The Go backend (auth flow, session handling, CSRF, middleware)
- The Svelte frontend (API client, auth guard, cookie handling)
- The build and bootstrapping tooling

Vulnerabilities in the template's **dev-only components** (Dex, Postgres
running without TLS) are out of scope — they are intentionally insecure for
local development only. Do not expose them publicly.

## Security model notes

- OIDC auth-code flow with **PKCE** — access tokens are never sent to the
  browser.
- Sessions are **server-side** in Postgres; the browser only holds an httpOnly
  + Secure + SameSite session cookie.
- CSRF protection via double-submit cookie on all state-changing requests.
- The dev CSRF secret and OIDC client secret in `.env.example` are plaintext
  placeholders — replace them with real random values before any production use.
