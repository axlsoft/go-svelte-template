package auth

import (
	"context"
	"net/http"
)

// ctxKey is an unexported context key type to avoid collisions.
type ctxKey int

const userKey ctxKey = iota

// UserFromContext returns the authenticated user stored by RequireAuth, or
// (nil, false) when the request is unauthenticated.
func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}

// RequireAuth is route-guard middleware: it resolves the session into a User in
// the request context, or replies 401 with the JSON error envelope when the
// request is unauthenticated.
func (m *SessionManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := m.Resolve(r.Context(), r)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
