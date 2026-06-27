# myapp — a Go+Svelte template

[![CI](https://github.com/OWNER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A reusable starting point for building a web application as a **Svelte 5 SPA
embedded into a single static Go binary**. One artifact to ship, one process to
run, no Node runtime in production.

It comes wired with in-app OIDC login, Postgres-backed sessions, embedded
database migrations, a typed frontend API client, and a light/dark theming and
account UX — so a fresh clone logs in and runs end to end.

## Why this exists

Most "full-stack" templates leave you juggling a Node server, a separate static
host, and a pile of glue. This one collapses all of that into **a single Go
binary** that:

- serves your JSON API and OIDC auth endpoints, and
- embeds the compiled Svelte SPA via `//go:embed` and serves it with client-side
  routing (deep-link refreshes return the app, not a 404).

The result is trivial to deploy — copy one file, run one systemd unit — while you
still get a modern Svelte 5 + Tailwind frontend in development.

### What's included

- **One binary.** The built SPA is embedded into the Go binary. Deploy one file.
- **Go backend.** `net/http` + `chi`, typed env config (fail-fast on boot),
  `slog` logging with request-ids, health probes, graceful shutdown, secure
  headers + per-IP rate limiting, and a consistent JSON error envelope.
- **Postgres** via `pgxpool` + `sqlc`/`pgx/v5`, with **goose** migrations
  embedded in the binary and run as an explicit step (`app migrate up`) — never
  on boot.
- **In-app OIDC** (auth-code + PKCE) with Postgres-backed server-side sessions,
  secure cookies, CSRF protection, and a route guard. **Dex** ships in
  `docker-compose` so login works on a fresh clone; production points at your
  OIDC provider by changing env vars.
- **Svelte 5 SPA** (SvelteKit + `adapter-static`, SPA mode) with Tailwind v4,
  a few shadcn-style base components, a typed API client, an auth store, and a
  profile/account UX with light/dark/system theming.

## Architecture

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

**Two-mode SPA handler** ([internal/server/spa.go](internal/server/spa.go)) — the
same catch-all behaves differently per environment:

- **prod:** serves the embedded SPA filesystem; unknown non-API `GET` routes
  return `index.html` so client-side routing works.
- **dev:** reverse-proxies unknown routes to the **Vite dev server** for HMR
  while Go still serves `/api` and `/auth`. You hit **Go**, not Vite.

**The embed wrinkle.** Go's `embed` can only see files at or below the embedding
package's directory, so it can't reach `web/build/` from `internal/server/`. The
build copies the frontend in first:

```
svelte build  →  web/build/
copy          →  internal/server/spa_dist/   (gitignored)
go build      →  embeds spa_dist into the binary
```

`internal/server/spa_dist/.gitkeep` is committed so the `server` package compiles
before any frontend is built.

**Backend shape.** Typed config validated on boot (prefix `MYAPP_`, fail fast on
missing/invalid); `slog` (JSON in prod, text in dev) with a request-id on every
request and panic recovery; one JSON error envelope
(`{ "error": { "code", "message", "request_id" } }`); `/healthz` (liveness, always
200) and `/readyz` (pings the DB); version/commit/date stamped via `-ldflags`.

**Data shape.** Postgres via `pgxpool`; hand-written SQL in
[internal/store/queries](internal/store/queries) compiled to type-safe Go by
**sqlc**; **goose** migrations embedded via `//go:embed` and run with the binary's
`migrate` subcommand (`up`/`down`/`status`/`create`).

**Auth shape.** `coreos/go-oidc` for discovery/JWKS/ID-token verification +
`golang.org/x/oauth2` for the auth-code flow with **PKCE** (plus `state` +
`nonce`); server-side sessions in Postgres with an httpOnly + Secure + SameSite
cookie and idle + absolute timeouts; CSRF protection on state-changing requests;
a normalized identity (`/auth/me`) exposing `name`, `preferred_username`,
`picture`, `roles`, and `is_admin`.

## Requirements

- [Go](https://go.dev/dl/) — see the version in [go.mod](go.mod)
- [Task](https://taskfile.dev) — `brew install go-task`
- [golangci-lint](https://golangci-lint.run) v2 — `brew install golangci-lint`
- [Docker](https://www.docker.com) or [Podman](https://podman.io) — local
  Postgres + Dex via compose
- [sqlc](https://sqlc.dev) — `brew install sqlc` (only to regenerate query code)
- [Node](https://nodejs.org) 22+ and npm — builds the SPA

## Quickstart

```sh
cp .env.example .env     # local config; defaults target the dev stack
task stack-up            # start Postgres + Dex (docker or podman compose)
task migrate-up          # create the sessions table
task dev                 # Go (air) + Vite (HMR) together
```

Open **http://127.0.0.1:8080** (the Go server — not Vite on :5173) and click
**Log in**. Dev credentials: `dev@example.com` / `password`.

To produce the shippable artifact instead:

```sh
task build               # build the SPA, embed it, build bin/myapp
./bin/myapp version      # prints version/commit/date
```

## Local development environment

`task dev` runs the Go server (live-reloaded by [`air`](https://github.com/air-verse/air))
and the Vite dev server (HMR) together. You always hit **Go** at
`http://127.0.0.1:8080`; it serves `/api` and `/auth` itself and proxies
everything else to Vite, so login works end to end in dev.

1. **Config.** `cp .env.example .env`. The defaults already point at the dev
   stack (Postgres + Dex in `docker-compose.yml`); edit only if you change ports
   or wire up a different IdP. Config is validated on boot and fails fast on a
   missing/invalid value.
2. **Dev stack.** `task stack-up` starts Postgres and Dex. The `db-*` / `stack-*`
   tasks auto-detect `docker` or `podman` (override with
   `COMPOSE="podman compose" task stack-up`).
3. **Migrations.** `task migrate-up` applies the embedded goose migrations
   (creates the `sessions` table). Schema lives in
   [internal/db/migrations](internal/db/migrations).
4. **Run.** `task dev`, then open `http://127.0.0.1:8080`.

`air` is pinned as a `go tool` dependency; fetch it once with
`go get -tool github.com/air-verse/air@latest`. If it isn't present, `task dev`
falls back to a plain `go run` (no Go live-reload; Vite HMR still works).

Common tasks (`task` lists them all):

```sh
task               # list every task
task dev           # live-reload dev (Go + Vite)
task build         # build the SPA, embed it, build bin/myapp
task run           # build and run the server
task stack-up      # start Postgres + Dex   (stack-down to stop)
task migrate-up    # apply migrations  (migrate-down / migrate-status / migrate-create)
task db-reset      # drop the volume, recreate, re-migrate
task sqlc-generate # regenerate internal/store/*.go from queries + schema
```

### Auth endpoints

| Endpoint         | Method | Purpose                                             |
| ---------------- | ------ | --------------------------------------------------- |
| `/auth/login`    | GET    | Redirect to the IdP (PKCE + state + nonce)          |
| `/auth/callback` | GET    | Validate, exchange code, create session, set cookie |
| `/auth/me`       | GET    | Current user as JSON (401 if no session)            |
| `/auth/logout`   | POST   | Clear session + cookie, return IdP logout URL       |
| `/api/me`        | GET    | Sample **protected** route (401 without a session)  |

Mutations (`POST/PUT/PATCH/DELETE`) require the CSRF token: the SPA reads the
`myapp_csrf` cookie and echoes it in the `X-CSRF-Token` header. Role extraction
is configurable via `MYAPP_OIDC_ROLES_CLAIM` (dot-path, default
`realm_access.roles`) and `MYAPP_OIDC_ADMIN_ROLE` (default `admin`), and
tolerates IdPs that emit a single role as a string instead of an array.

## Testing

The repo stays green across both stacks. Run the backend and frontend suites:

```sh
# Backend
task test          # go test ./...
task lint          # golangci-lint run

# Frontend (or run npm scripts directly inside web/)
task web-test      # vitest unit + component tests
task web-check     # svelte-check (types)
task web-lint      # eslint + prettier --check
```

Then confirm the whole thing builds with the SPA embedded:

```sh
task build         # svelte build → embed → go build
```

Backend tests cover config validation, auth (sessions, CSRF, normalized
identity), and the server middleware/SPA handler. Frontend tests cover the API
client, theme resolution, and the components with logic (avatar initials and
image-error fallback). CI ([.github/workflows/ci.yml](.github/workflows/ci.yml))
runs the Go build/test/lint and the frontend check/lint/format/test on every push
and PR.

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
.github/workflows/   # CI (build · test · lint)
```

## Bootstrapping a new project

This repo is a template stamped with placeholders. To start a real project from
it, rewrite the module path with `gonew`, then run `bootstrap.sh` for everything
else:

```sh
# Install gonew once:
go install golang.org/x/tools/cmd/gonew@latest

# 1. Copy the template under a new module path:
gonew github.com/OWNER/REPO github.com/you/newproject
cd newproject

# 2. Rewrite the app name, env prefix, and systemd unit name:
./bootstrap.sh newproject

# 3. Verify it still builds:
task build && ./bin/newproject version
```

`gonew` rewrites the Go module path; `bootstrap.sh` rewrites the remaining
placeholders and is idempotent (running it again is a no-op).

| Placeholder             | Meaning           | Rewritten by             |
| ----------------------- | ----------------- | ------------------------ |
| `github.com/OWNER/REPO` | Go module path    | `gonew`                  |
| `myapp`                 | App name          | `bootstrap.sh <newname>` |
| `MYAPP_`                | Env var prefix    | `bootstrap.sh <newname>` |
| `myapp.service`         | systemd unit name | `bootstrap.sh <newname>` |

## Deployment

The deployment model is deliberately boring: build the binary, ship that one
file, run the database migration as an explicit step, then (re)start the service.

```sh
task build                 # produces bin/myapp with the SPA embedded
./bin/myapp migrate up     # apply pending migrations before starting
./bin/myapp                # run (binds MYAPP_HTTP_HOST:MYAPP_HTTP_PORT)
```

In production, set the `MYAPP_*` env vars (point `MYAPP_OIDC_*` at your OIDC
provider), run the binary behind your own TLS-terminating reverse proxy, and
supervise it with systemd. `deploy/ansible/` is a placeholder for automating that
roll-out; the automation itself is left to you.

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before
opening a pull request. For security issues, see [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
