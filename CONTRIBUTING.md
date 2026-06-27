# Contributing

Thank you for taking the time to contribute. This is a template repository, so
contributions that improve the starting point for everyone are especially
welcome: better defaults, clearer documentation, fixed bugs, and well-scoped
new features.

## What belongs here

This repo is a **starting point**, not a framework. A good contribution makes
the template more useful without adding complexity that every project spawned
from it would have to remove. When in doubt, keep it small.

Good candidates:
- Bug fixes in the Go backend, auth flow, or SPA shell
- Improvements to the bootstrap/stamp workflow
- Documentation clarifications
- Test coverage gaps
- Dependency updates (with rationale)

Out of scope:
- Application-specific business logic
- Opinionated UI components beyond the minimal set already included
- Alternative database engines or auth methods (open an issue to discuss first)

## Getting started

```sh
git clone https://github.com/OWNER/REPO.git
cd REPO
cp .env.example .env
task stack-up       # start Postgres + Dex
task migrate-up     # create the sessions table
task dev            # Go (air) + Vite (HMR)
```

Run the full test suite before opening a PR:

```sh
task test           # Go tests
task lint           # golangci-lint
task web-check      # svelte-check (types)
task web-lint       # eslint + prettier
task web-test       # vitest
task build          # full embedded build
```

CI runs all of the above on every push and PR.

## Submitting a pull request

1. Fork the repo and create a branch from `main`.
2. Make your change. Keep commits small and conventional
   (`fix:`, `feat:`, `docs:`, `chore:`, etc.).
3. Add or update tests where relevant. Don't leave CI red.
4. Update `README.md` or `Taskfile.yml` if you add a workflow a developer
   would run.
5. Open a PR against `main`. Describe *what* changed and *why*.

There is no formal review SLA. Small, focused PRs with clear descriptions move
fastest.

## Code style

- **Go:** standard `gofmt` formatting; `golangci-lint` must pass.
- **TypeScript/Svelte:** Prettier and ESLint (run via `task web-lint`).
- **Comments:** explain *why*, not *what*. One line max per comment block.
- **No new dependencies** without a note explaining why the standard library or
  an existing dependency doesn't cover the need.

## Reporting bugs

Please use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md). Include
the Go version, Node version, and OS, plus enough detail to reproduce the issue.

## License

By contributing you agree that your contributions will be licensed under the
[MIT License](LICENSE).
