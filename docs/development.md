# Development

## Requirements

- [Go](https://go.dev/dl/) — see the version in [go.mod](../go.mod)
- [Task](https://taskfile.dev) — `brew install go-task`
- [golangci-lint](https://golangci-lint.run) v2 — `brew install golangci-lint`
- [Docker](https://www.docker.com) or [Podman](https://podman.io) — local
  Postgres + Dex via compose
- [sqlc](https://sqlc.dev) — `brew install sqlc` (only to regenerate query code)
- [Node](https://nodejs.org) 22+ and npm — builds the SPA

## Local environment

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
   [internal/db/migrations](../internal/db/migrations).
4. **Run.** `task dev`, then open `http://127.0.0.1:8080`. Dev credentials:
   `dev@example.com` / `password`.

`air` is pinned as a `go tool` dependency; fetch it once with
`go get -tool github.com/air-verse/air@latest`. If it isn't present, `task dev`
falls back to a plain `go run` (no Go live-reload; Vite HMR still works).

## Common tasks

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

## Testing

```sh
# Backend
task test          # go test ./...
task lint          # golangci-lint run

# Frontend (or run npm scripts directly inside web/)
task web-test      # vitest unit + component tests
task web-check     # svelte-check (types)
task web-lint      # eslint + prettier --check

# Whole thing, with the SPA embedded
task build         # svelte build → embed → go build
```

Backend tests cover config validation, auth (sessions, CSRF, normalized
identity), and the server middleware/SPA handler. Frontend tests cover the API
client, theme resolution, and the components with logic (avatar initials and
image-error fallback). CI ([.github/workflows/ci.yml](../.github/workflows/ci.yml))
runs the Go build/test/lint and the frontend check/lint/format/test on every push
and PR.
