package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_SetAndPropagate(t *testing.T) {
	t.Parallel()

	var seen string
	h := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen == "" {
		t.Fatal("request id not present in context")
	}
	if got := rec.Header().Get(requestIDHeader); got != seen {
		t.Errorf("response header %s = %q, want %q", requestIDHeader, got, seen)
	}
}

func TestRequestID_AdoptsInbound(t *testing.T) {
	t.Parallel()

	const inbound = "abc-123"
	var seen string
	h := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, inbound)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen != inbound {
		t.Errorf("request id = %q, want inbound %q", seen, inbound)
	}
}

func TestRecoverer_PanicToCleanError(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Wrap with RequestID so the envelope carries a request_id.
	h := RequestID(Recoverer(logger)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not the JSON error envelope: %v (body=%q)", err, rec.Body.String())
	}
	if resp.Error.Code != "internal_error" {
		t.Errorf("error.code = %q, want internal_error", resp.Error.Code)
	}
	if resp.Error.RequestID == "" {
		t.Error("error.request_id is empty")
	}
	// The internal panic value must not leak to the client.
	if resp.Error.Message == "boom" {
		t.Error("panic value leaked into the response message")
	}
}
