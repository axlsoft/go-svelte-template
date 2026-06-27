package server

// The TWO-MODE SPA handler:
//
//   - prod: serve the embedded SPA filesystem (//go:embed spa_dist); unknown,
//     non-API GET routes fall back to index.html for client-side routing.
//   - dev:  reverse-proxy unknown routes to the Vite dev server for HMR.
//
// API/system prefixes (/api, /auth, /healthz, /readyz) must never fall through
// to the SPA.
//
// The embed wrinkle: Go's embed can only reach files at or below this package's
// directory, so `task build` copies web/build into ./spa_dist before `go build`.
// An empty spa_dist/.gitkeep is committed so the package compiles before any
// frontend exists.

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

// spaFS embeds the built SPA. `all:` keeps the committed .gitkeep so the package
// compiles before any frontend is built. `task build` copies web/build here.
//
//go:embed all:spa_dist
var spaFS embed.FS

// spaPlaceholder is served in prod when no index.html has been built yet (e.g.
// running the binary before `task build` has produced the SPA). It keeps the
// server useful during Phases 1–3, before the frontend exists.
const spaPlaceholder = `<!doctype html>
<meta charset="utf-8">
<title>myapp</title>
<h1>myapp</h1>
<p>The SPA has not been built into this binary yet. Run <code>task build</code>
to embed the frontend, or use <code>task dev</code> for HMR.</p>`

// reservedPrefixes are owned by the Go server and must never fall through to the
// SPA handler. The router mounts them explicitly; the SPA handler also guards
// against them defensively.
var reservedPrefixes = []string{"/api", "/auth", "/healthz", "/readyz"}

// IsReservedPath reports whether p belongs to a server-owned prefix (API/system)
// rather than the client-routed SPA. This is the route-classification rule the
// SPA handler relies on.
func IsReservedPath(p string) bool {
	for _, pre := range reservedPrefixes {
		if p == pre || strings.HasPrefix(p, pre+"/") {
			return true
		}
	}
	return false
}

// NewSPAHandler returns the catch-all SPA handler for the given environment.
//
// In dev it reverse-proxies to the Vite dev server (HMR); in prod it serves the
// embedded SPA with index.html fallback for client-side routes.
func NewSPAHandler(dev bool, viteDevURL string, logger *slog.Logger) (http.Handler, error) {
	if dev {
		return newDevProxy(viteDevURL, logger)
	}
	return newProdSPA(logger)
}

// --- dev mode: reverse-proxy to Vite ---------------------------------------

func newDevProxy(viteDevURL string, logger *slog.Logger) (http.Handler, error) {
	target, err := url.Parse(viteDevURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.ErrorContext(r.Context(), "vite proxy error",
			slog.String("request_id", RequestIDFromContext(r.Context())),
			slog.String("target", viteDevURL),
			slog.Any("err", err),
		)
		WriteError(w, r, http.StatusBadGateway, "vite_unreachable",
			"dev: cannot reach the Vite dev server (is `task dev` running?)")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defensive: reserved paths are mounted on the router and must not be
		// proxied to Vite.
		if IsReservedPath(r.URL.Path) {
			WriteError(w, r, http.StatusNotFound, "not_found", "not found")
			return
		}
		proxy.ServeHTTP(w, r)
	}), nil
}

// --- prod mode: serve embedded SPA -----------------------------------------

type prodSPA struct {
	fsys       fs.FS
	fileServer http.Handler
	indexHTML  []byte
	hasIndex   bool
}

func newProdSPA(logger *slog.Logger) (http.Handler, error) {
	sub, err := fs.Sub(spaFS, "spa_dist")
	if err != nil {
		return nil, err
	}

	h := &prodSPA{
		fsys:       sub,
		fileServer: http.FileServer(http.FS(sub)),
	}

	if data, err := readIndex(sub); err == nil {
		h.indexHTML = data
		h.hasIndex = true
	} else {
		logger.Warn("no embedded SPA index.html; serving placeholder until `task build` runs")
	}

	return h, nil
}

func readIndex(fsys fs.FS) ([]byte, error) {
	f, err := fsys.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(f)
}

func (h *prodSPA) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Defensive: reserved paths must never be served by the SPA.
	if IsReservedPath(r.URL.Path) {
		WriteError(w, r, http.StatusNotFound, "not_found", "not found")
		return
	}

	// Only GET/HEAD are eligible for static assets or the SPA fallback.
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		WriteError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	upath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if upath != "" && h.fileExists(upath) {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	// Unknown route → hand back index.html so the client router takes over
	// (deep-link refresh returns the app, not a 404).
	h.serveIndex(w, r)
}

func (h *prodSPA) fileExists(name string) bool {
	f, err := h.fsys.Open(name)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		return false
	}
	return true
}

func (h *prodSPA) serveIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if h.hasIndex {
		_, _ = w.Write(h.indexHTML)
		return
	}
	_, _ = w.Write([]byte(spaPlaceholder))
}
