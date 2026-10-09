# myapp — a Go+Svelte template

[![CI](https://github.com/OWNER/REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A **Svelte 5 SPA embedded into a single static Go binary**. One artifact to ship,
one process to run, no Node runtime in production. Ships with in-app OIDC login,
Postgres-backed sessions, embedded migrations, a typed API client, and
light/dark theming — a fresh clone logs in and runs end to end.

## Quickstart

```sh
cp .env.example .env     # local config; defaults target the dev stack
task stack-up            # start Postgres + Dex (docker or podman compose)
task migrate-up          # create the sessions table
task dev                 # Go (air) + Vite (HMR) together
```

Open **<http://127.0.0.1:8080>** (Go, not Vite) and log in as
`dev@example.com` / `password`. `task build` produces `bin/myapp` with the SPA
embedded.

## Docs

- [Architecture](docs/architecture.md) — request flow, SPA embedding, project layout
- [Development](docs/development.md) — requirements, local setup, tasks, testing
- [Auth](docs/auth.md) — OIDC flow, endpoints, CSRF, roles
- [Bootstrapping](docs/bootstrapping.md) — start a new project from this template
- [Deployment](docs/deployment.md) — build, migrate, run

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
