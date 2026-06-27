// HTTP error envelope returned by the Go server, e.g.
//
//	{ "error": { "code": "unauthorized", "message": "…", "request_id": "…" } }
//
// The typed client mirrors this shape so callers get structured errors.
export interface ApiErrorBody {
	code: string;
	message: string;
	request_id?: string;
}

export interface ApiErrorEnvelope {
	error: ApiErrorBody;
}

// User is the normalized identity returned by GET /auth/me. All claim parsing
// (roles, names, picture) happens in Go, so the client only ever sees this clean
// shape. `roles` is always present (possibly empty) and `is_admin` is derived
// server-side. The sample GET /api/me echoes a subset of this.
export interface User {
	sub: string;
	email: string;
	name: string;
	given_name?: string;
	family_name?: string;
	preferred_username?: string;
	picture?: string;
	roles: string[];
	is_admin: boolean;
	// signed_in_since is an RFC3339 timestamp of when the session was created.
	signed_in_since?: string;
}

// LogoutResponse is returned by POST /auth/logout.
export interface LogoutResponse {
	logout_url?: string;
}
