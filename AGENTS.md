# AGENTS.md

Context for any coding agent (or human contributor) working in this repo.
Read this before making changes.

## What this is

A reusable template for building a web application as a **Svelte 5 SPA embedded
into a single static Go binary**. One artifact to ship, one process to run, no
Node runtime in production. It ships wired with OIDC auth, Postgres-backed
sessions, embedded migrations, a typed frontend API client, and light/dark
theming.

When stamping a new project from this template, `gonew` rewrites the module
path and `bootstrap.sh` rewrites the remaining placeholders. See the
placeholder table in `README.md`.

## Project layout

```
cmd/server/          # binary entrypoint — wires config → db → router → shutdown
internal/
  config/            # typed env config, validated on boot (fail fast)
  version/           # Version/Commit/Date set via -ldflags
  server/            # http.Server, middleware, routes, JSON errors, two-mode SPA handler
    spa_dist/        # gitignored; web/build is copied here at build time
  auth/              # OIDC flow, sessions, CSRF, guard, handlers, identity
  db/                # pgxpool, goose runner, embedded migrations
  store/             # sqlc-generated queries + models
  health/            # /healthz (liveness), /readyz (DB ping)
web/                 # SvelteKit SPA — adapter-static, SPA mode, no server features
config/dex/          # Dev IdP config (static dev@example.com user for local login)
deploy/ansible/      # Placeholder for deployment automation
.github/workflows/   # CI: build · test · lint on every push/PR
```

## Non-negotiable constraints

- **SvelteKit is a routing + build tool only.** No `+page.server.ts`, no form
  actions, no `hooks.server.ts`, no `$app/server`. Everything server-side lives
  in Go. (See also `web/AGENTS.md`.)
- **Config is validated on boot — fail fast.** Missing or invalid env vars log
  a clear message and exit non-zero. Do not add silent defaults for required
  values.
- **Migrations are never run on boot.** They run as an explicit deploy step
  (`app migrate up`). Do not add auto-migration logic to the server startup.
- **The embed wrinkle.** Go's `embed` can only reach files at or below the
  embedding package's directory. `task build` copies `web/build` into
  `internal/server/spa_dist/` before `go build`. Do not change this flow
  without updating the `build` task and README.
- **Placeholders must stay intact.** `myapp`, `MYAPP_`, and
  `github.com/OWNER/REPO` are rewritten by `bootstrap.sh` / `gonew`. Do not
  hardcode a real app name anywhere in the template source.

## Key decisions (don't contradict these)

| Area | Decision |
|---|---|
| Router | `net/http` + `chi` |
| Config | Typed struct, env-sourced, prefix `MYAPP_`, validate on boot |
| Logging | `slog` — JSON in prod, text in dev, request-id on every request |
| Database | Postgres via `pgxpool`; queries via `sqlc` + `pgx/v5` |
| Migrations | `goose`, embedded in binary, explicit deploy step |
| Auth | In-app OIDC: `coreos/go-oidc` + `golang.org/x/oauth2`, auth-code + PKCE |
| Sessions | Server-side in Postgres; httpOnly + Secure + SameSite cookie |
| Dev IdP | Dex in `docker-compose`; prod points at any OIDC-compliant IdP |
| Frontend | SvelteKit + `adapter-static`, SPA mode, Tailwind v4, Bits UI |
| Build | `task build` = svelte build → copy → `go build` |
| Dev | `task dev` = `air` (Go reload) + `vite` (HMR); developer hits Go |

## Working in this repo

- Run the full suite before opening a PR: `task test && task lint && task web-test && task web-check && task build`.
- CI must stay green. If you add behaviour, add a test.
- Update `README.md` and `Taskfile.yml` if you add a workflow a developer
  would run.
- Keep commits small and conventional (`fix:`, `feat:`, `docs:`, `chore:`).
- Prefer the standard library and already-chosen libraries. Don't add
  dependencies without explaining why.
- If a decision is genuinely ambiguous, make the smallest reasonable choice,
  implement it, and leave a short `// NOTE:` explaining it.

## Auth endpoints (for reference)

| Endpoint | Method | Purpose |
|---|---|---|
| `/auth/login` | GET | Redirect to IdP (PKCE + state + nonce) |
| `/auth/callback` | GET | Validate, exchange code, create session, set cookie |
| `/auth/me` | GET | Current user as JSON (401 if no session) |
| `/auth/logout` | POST | Clear session + cookie, return IdP logout URL |
| `/api/me` | GET | Sample protected route (401 without session) |

CSRF token (`myapp_csrf` cookie → `X-CSRF-Token` header) is required on all
state-changing requests. `/auth/callback` is exempt (the `state` param covers
it).

## Roles and admin

Roles are extracted from the ID-token claims at a configurable dot-path
(`MYAPP_OIDC_ROLES_CLAIM`, default `realm_access.roles`). The value is
tolerated as either an array or a single string. `MYAPP_OIDC_ADMIN_ROLE`
(default `admin`) sets the admin role. All claim parsing happens in Go;
`/auth/me` exposes a clean `{ roles, is_admin }` shape — the frontend never
touches raw claims.

The client-side `is_admin` check is **cosmetic only**. Any real admin endpoint
must enforce the role server-side in Go.
