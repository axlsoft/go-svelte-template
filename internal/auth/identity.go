package auth

import (
	"encoding/json"
	"strings"
	"time"
)

// Identity is the normalized identity the frontend consumes from /auth/me. All
// claim parsing happens in Go so the client never has to touch raw IdP claims or
// cope with provider quirks. Roles is always non-nil (never JSON null).
type Identity struct {
	Sub               string    `json:"sub"`
	Email             string    `json:"email,omitempty"`
	Name              string    `json:"name,omitempty"`
	GivenName         string    `json:"given_name,omitempty"`
	FamilyName        string    `json:"family_name,omitempty"`
	PreferredUsername string    `json:"preferred_username,omitempty"`
	Picture           string    `json:"picture,omitempty"`
	Roles             []string  `json:"roles"`
	IsAdmin           bool      `json:"is_admin"`
	SignedInSince     time.Time `json:"signed_in_since,omitempty"`
}

// NewIdentity builds the normalized identity from a session User. rolesClaimPath
// is a dot-path into the claims (e.g. "realm_access.roles" or "roles") and
// adminRole is the role that grants is_admin. The roles value is tolerated
// whether the IdP emits an array or a single scalar string.
func NewIdentity(u *User, rolesClaimPath, adminRole string) *Identity {
	id := &Identity{
		Sub:           u.Sub,
		Email:         u.Email,
		Name:          u.Name,
		SignedInSince: u.SignedInSince,
		Roles:         []string{},
	}

	var claims map[string]any
	if len(u.Claims) > 0 {
		_ = json.Unmarshal(u.Claims, &claims)
	}

	id.GivenName = claimString(claims, "given_name")
	id.FamilyName = claimString(claims, "family_name")
	id.PreferredUsername = claimString(claims, "preferred_username")
	id.Picture = claimString(claims, "picture")
	if id.Email == "" {
		id.Email = claimString(claims, "email")
	}

	// Display-name precedence: explicit name -> given+family -> username -> email.
	if id.Name == "" {
		id.Name = claimString(claims, "name")
	}
	if id.Name == "" {
		id.Name = strings.TrimSpace(id.GivenName + " " + id.FamilyName)
	}
	if id.Name == "" {
		id.Name = id.PreferredUsername
	}
	if id.Name == "" {
		id.Name = id.Email
	}

	id.Roles = rolesFromClaims(claims, rolesClaimPath)
	id.IsAdmin = containsFold(id.Roles, adminRole)

	return id
}

// claimString returns claims[key] when it is a string, else "".
func claimString(claims map[string]any, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

// lookupPath walks a dot-separated path through nested JSON objects and returns
// the value at the leaf, or nil if any segment is missing or not an object.
func lookupPath(claims map[string]any, path string) any {
	if claims == nil || path == "" {
		return nil
	}
	var cur any = claims
	for _, seg := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = m[seg]
		if !ok {
			return nil
		}
	}
	return cur
}

// rolesFromClaims extracts roles at path, tolerating the value being a single
// string, a []any of strings (the common JSON-decoded shape), or a []string.
// Empty/whitespace entries are dropped. The result is always non-nil.
func rolesFromClaims(claims map[string]any, path string) []string {
	out := []string{}
	switch v := lookupPath(claims, path).(type) {
	case string:
		if s := strings.TrimSpace(v); s != "" {
			out = append(out, s)
		}
	case []any:
		for _, e := range v {
			if s, ok := e.(string); ok {
				if s = strings.TrimSpace(s); s != "" {
					out = append(out, s)
				}
			}
		}
	case []string:
		for _, s := range v {
			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

// containsFold reports whether want is in roles, case-insensitively.
func containsFold(roles []string, want string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, want) {
			return true
		}
	}
	return false
}
