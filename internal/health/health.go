// Package health serves liveness and readiness probes:
//
//   - /healthz: liveness, always 200.
//   - /readyz:  readiness, pings the DB (200 healthy / 503 unhealthy).
package health

import (
	"context"
	"net/http"
)

// Live is the liveness handler: always 200. A reachable process is "live".
func Live(w http.ResponseWriter, _ *http.Request) {
	writePlain(w, http.StatusOK, "ok")
}

// Ready returns a readiness handler. The check probes a downstream dependency
// (the DB); a nil error means ready (200), any error means unavailable (503).
// A nil check always reports ready (useful before a DB exists).
func Ready(check func(ctx context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check == nil {
			writePlain(w, http.StatusOK, "ready")
			return
		}
		if err := check(r.Context()); err != nil {
			writePlain(w, http.StatusServiceUnavailable, "unavailable")
			return
		}
		writePlain(w, http.StatusOK, "ready")
	}
}

func writePlain(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
