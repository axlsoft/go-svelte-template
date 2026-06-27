package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorBody is the inner object of the JSON error envelope.
type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// ErrorResponse is the consistent JSON error envelope used by all handlers:
//
//	{ "error": { "code", "message", "request_id" } }
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// WriteError writes status with the standard JSON error envelope. The request id
// is pulled from the request context (set by the RequestID middleware) so every
// error is correlatable with its logs.
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	resp := ErrorResponse{Error: ErrorBody{
		Code:      code,
		Message:   message,
		RequestID: RequestIDFromContext(r.Context()),
	}}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// The header/status are already written; just record the failure.
		slog.ErrorContext(r.Context(), "failed to encode error envelope", "err", err)
	}
}
