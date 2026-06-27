package auth

import (
	"encoding/json"
	"net/http"
)

// The auth package writes the same JSON error envelope as the server package.
// It is duplicated here (rather than imported) to avoid an import cycle: the
// server package imports auth to mount the handlers and guard. The request id is
// read from the response header, which the RequestID middleware sets before any
// handler runs.

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// writeJSONError writes status with the standard JSON error envelope.
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rid := w.Header().Get("X-Request-Id")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: errorBody{
		Code:      code,
		Message:   message,
		RequestID: rid,
	}})
}
