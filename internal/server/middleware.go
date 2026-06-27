package server

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

// requestIDHeader is the canonical header carrying the request id in and out.
const requestIDHeader = "X-Request-Id"

// ctxKey is an unexported context key type to avoid collisions.
type ctxKey int

const requestIDKey ctxKey = iota

// RequestIDFromContext returns the request id stored in ctx, or "" if absent.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// RequestID generates (or adopts an inbound) request id, stores it in the
// context, and echoes it on the response header so clients and logs can
// correlate a request end-to-end.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID returns a random 16-byte hex id (falls back to a timestamp if the
// system RNG is somehow unavailable).
func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "ts-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}

// statusRecorder wraps http.ResponseWriter to capture the status code and the
// number of bytes written for request logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	n, err := sr.ResponseWriter.Write(b)
	sr.bytes += n
	return n, err
}

// Hijack forwards to the wrapped ResponseWriter so protocol switching (e.g. the
// Vite HMR websocket proxied through Go in dev) keeps working. Without this the
// reverse proxy fails with "non-Hijacker ResponseWriter".
func (sr *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := sr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
	}
	return hj.Hijack()
}

// Flush forwards flushes (streaming responses / SSE) to the wrapped writer.
func (sr *statusRecorder) Flush() {
	if fl, ok := sr.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

// RequestLogger logs one structured line per request including the request id,
// method, path, status, response size, latency and client ip.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sr, r)

			logger.LogAttrs(r.Context(), slog.LevelInfo, "http request",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sr.status),
				slog.Int("bytes", sr.bytes),
				slog.Duration("latency", time.Since(start)),
				slog.String("remote_ip", clientIP(r)),
			)
		})
	}
}

// Recoverer turns a panicking handler into a clean 500 + JSON error envelope,
// logging the stack server-side without leaking internals to the client.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered",
						slog.String("request_id", RequestIDFromContext(r.Context())),
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
					)
					WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders sets conservative security headers on every response.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("X-XSS-Protection", "0") // modern browsers; rely on CSP instead
		// NOTE: a strict Content-Security-Policy can be added here once the SPA
		// asset shape is locked down.
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the best-effort client IP for logging and rate limiting.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// rateLimiter is a simple in-memory per-IP token bucket.
//
// NOTE: this is intentionally process-local and lossy across restarts/instances.
// Swap it for a Redis-backed limiter when running more than one replica.
type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens added per second
	burst    float64 // maximum tokens (and initial allowance)
	ttl      time.Duration
}

type visitor struct {
	tokens   float64
	lastFill time.Time
	lastSeen time.Time
}

// newRateLimiter creates a limiter allowing `burst` requests immediately and
// refilling at `rps` tokens/second per client IP.
func newRateLimiter(rps, burst float64) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rps,
		burst:    burst,
		ttl:      10 * time.Minute,
	}
	go rl.janitor()
	return rl
}

// allow reports whether a request from ip may proceed, consuming a token.
func (rl *rateLimiter) allow(ip string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[ip]
	if !ok {
		rl.visitors[ip] = &visitor{tokens: rl.burst - 1, lastFill: now, lastSeen: now}
		return true
	}

	// Refill based on elapsed time, capped at burst.
	elapsed := now.Sub(v.lastFill).Seconds()
	v.tokens = min(rl.burst, v.tokens+elapsed*rl.rate)
	v.lastFill = now
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// janitor periodically evicts idle visitors to bound memory.
func (rl *rateLimiter) janitor() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-rl.ttl)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if v.lastSeen.Before(cutoff) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns middleware that applies a per-IP token bucket, replying 429
// with the JSON error envelope when a client exceeds its allowance.
func RateLimit(rps, burst float64) func(http.Handler) http.Handler {
	rl := newRateLimiter(rps, burst)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(clientIP(r)) {
				w.Header().Set("Retry-After", "1")
				WriteError(w, r, http.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
