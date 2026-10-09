# Deployment

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
