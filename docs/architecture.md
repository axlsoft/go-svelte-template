# Architecture

The Go process owns every request. A single `chi` router classifies them, and
everything that isn't an API/auth/health route is handed to the SPA.

```
            ┌─────────────────────────────────────────────┐
  client ──▶│  myapp (single Go binary)                    │──▶ Postgres
            │                                              │
            │  /healthz /readyz   → health probes          │──▶ OIDC IdP
            │  /api/*             → JSON API (auth-guarded) │   (Dex in dev / your IdP in prod)
            │  /auth/*            → OIDC login/callback/me  │
            │  everything else    → embedded SPA + fallback │
            └─────────────────────────────────────────────┘
```

## Two-mode SPA handler

[internal/server/spa.go](../internal/server/spa.go) — the same catch-all behaves
differently per environment:

- **prod:** serves the embedded SPA filesystem; unknown non-API `GET` routes
  return `index.html` so client-side routing works.
- **dev:** reverse-proxies unknown routes to the **Vite dev server** for HMR
  while Go still serves `/api` and `/auth`. You hit **Go**, not Vite.

## The embed wrinkle

Go's `embed` can only see files at or below the embedding package's directory,
so it can't reach `web/build/` from `internal/server/`. The build copies the
frontend in first:

```
svelte build  →  web/build/
copy          →  internal/server/spa_dist/   (gitignored)
go build      →  embeds spa_dist into the binary
```

`internal/server/spa_dist/.gitkeep` is committed so the `server` package compiles
before any frontend is built.

## Backend

Typed config validated on boot (prefix `MYAPP_`, fail fast on missing/invalid);
`slog` (JSON in prod, text in dev) with a request-id on every request and panic
recovery; one JSON error envelope
(`{ "error": { "code", "message", "request_id" } }`); `/healthz` (liveness, always
200) and `/readyz` (pings the DB); version/commit/date stamped via `-ldflags`.

## Data

Postgres via `pgxpool`; hand-written SQL in
[internal/store/queries](../internal/store/queries) compiled to type-safe Go by
**sqlc**; **goose** migrations embedded via `//go:embed` and run with the binary's
`migrate` subcommand (`up`/`down`/`status`/`create`).

## Auth

See [auth.md](auth.md).

## Project layout

```
cmd/server/          # binary entrypoint (wiring + migrate subcommand)
internal/
  config/            # typed env config, validate-on-boot
  version/           # -ldflags target (Version/Commit/Date)
  server/            # http.Server, middleware, routes, errors, two-mode SPA
    spa_dist/        # gitignored; web/build is copied here at build time
  auth/              # oidc, session, csrf, guard, handlers, normalized identity
  db/                # pgxpool, goose runner, embedded migrations
  store/             # sqlc queries + generated code
  health/            # /healthz, /readyz
web/                 # SvelteKit SPA (build-only; no server features)
config/dex/          # dev IdP config (static dev@example.com user)
deploy/ansible/      # placeholder for deployment automation (not included)
docs/                # these docs
.github/workflows/   # CI (build · test · lint)
```
