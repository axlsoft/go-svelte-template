# Auth

`coreos/go-oidc` handles discovery/JWKS/ID-token verification and
`golang.org/x/oauth2` runs the auth-code flow with **PKCE** (plus `state` +
`nonce`). Sessions are server-side in Postgres with an httpOnly + Secure +
SameSite cookie and idle + absolute timeouts. CSRF protection covers
state-changing requests. `/auth/me` exposes a normalized identity: `name`,
`preferred_username`, `picture`, `roles`, and `is_admin`.

## Endpoints

| Endpoint         | Method | Purpose                                             |
| ---------------- | ------ | --------------------------------------------------- |
| `/auth/login`    | GET    | Redirect to the IdP (PKCE + state + nonce)          |
| `/auth/callback` | GET    | Validate, exchange code, create session, set cookie |
| `/auth/me`       | GET    | Current user as JSON (401 if no session)            |
| `/auth/logout`   | POST   | Clear session + cookie, return IdP logout URL       |
| `/api/me`        | GET    | Sample **protected** route (401 without a session)  |

## CSRF

Mutations (`POST/PUT/PATCH/DELETE`) require the CSRF token: the SPA reads the
`myapp_csrf` cookie and echoes it in the `X-CSRF-Token` header.

## Roles

Role extraction is configurable via `MYAPP_OIDC_ROLES_CLAIM` (dot-path, default
`realm_access.roles`) and `MYAPP_OIDC_ADMIN_ROLE` (default `admin`), and
tolerates IdPs that emit a single role as a string instead of an array.

The client-side `is_admin` check is cosmetic only — any real admin endpoint must
enforce the role server-side in Go.
