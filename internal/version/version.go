// Package version exposes build metadata stamped into the binary at link time.
//
// The vars below are overridden via -ldflags during `task build` / release, e.g.:
//
//	go build -ldflags "\
//	  -X github.com/OWNER/REPO/internal/version.Version=v1.2.3 \
//	  -X github.com/OWNER/REPO/internal/version.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/OWNER/REPO/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
//
// When built without ldflags (e.g. `go run`, tests) they keep the "dev" defaults.
package version

// Build metadata. Overridden at link time; do not set these elsewhere.
var (
	// Version is the released semver tag (or "dev" for local builds).
	Version = "dev"
	// Commit is the short git SHA the binary was built from.
	Commit = "none"
	// Date is the RFC3339 UTC build timestamp.
	Date = "unknown"
)

// String returns a human-readable one-line build identifier.
func String() string {
	return Version + " (commit " + Commit + ", built " + Date + ")"
}
