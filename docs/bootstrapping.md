# Bootstrapping a new project

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
