package auth

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestNewIdentity_NamesAndPicture(t *testing.T) {
	t.Parallel()

	claims := mustClaims(t, map[string]any{
		"given_name":         "Ada",
		"family_name":        "Lovelace",
		"preferred_username": "ada",
		"email":              "ada@example.com",
		"picture":            "https://img.example.com/ada.png",
	})
	since := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	u := &User{Sub: "user-1", Claims: claims, SignedInSince: since}

	id := NewIdentity(u, "realm_access.roles", "admin")

	if id.Sub != "user-1" {
		t.Errorf("sub = %q", id.Sub)
	}
	if id.GivenName != "Ada" || id.FamilyName != "Lovelace" {
		t.Errorf("name parts = %q %q", id.GivenName, id.FamilyName)
	}
	if id.PreferredUsername != "ada" {
		t.Errorf("preferred_username = %q", id.PreferredUsername)
	}
	if id.Picture != "https://img.example.com/ada.png" {
		t.Errorf("picture = %q", id.Picture)
	}
	// Display name falls back to given+family when no explicit name claim.
	if id.Name != "Ada Lovelace" {
		t.Errorf("name = %q, want %q", id.Name, "Ada Lovelace")
	}
	if !id.SignedInSince.Equal(since) {
		t.Errorf("signed_in_since = %v", id.SignedInSince)
	}
	if id.Roles == nil {
		t.Error("roles must be non-nil")
	}
}

func TestNewIdentity_DisplayNamePrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		claims map[string]any
		user   *User
		want   string
	}{
		{
			name: "explicit name wins",
			user: &User{Name: "Stored Name"},
			want: "Stored Name",
		},
		{
			name:   "name claim",
			claims: map[string]any{"name": "Claim Name", "preferred_username": "u"},
			want:   "Claim Name",
		},
		{
			name:   "username when no name",
			claims: map[string]any{"preferred_username": "ada", "email": "ada@example.com"},
			want:   "ada",
		},
		{
			name:   "email last resort",
			claims: map[string]any{"email": "only@example.com"},
			want:   "only@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.user
			if u == nil {
				u = &User{}
			}
			u.Claims = mustClaims(t, tt.claims)
			id := NewIdentity(u, "realm_access.roles", "admin")
			if id.Name != tt.want {
				t.Errorf("name = %q, want %q", id.Name, tt.want)
			}
		})
	}
}

func TestNewIdentity_RolesArrayOrScalar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		roles     any
		wantRoles []string
		wantAdmin bool
	}{
		{
			name:      "array of roles",
			roles:     []any{"user", "admin"},
			wantRoles: []string{"user", "admin"},
			wantAdmin: true,
		},
		{
			name:      "single scalar string (some IdPs emit a scalar for a single role)",
			roles:     "admin",
			wantRoles: []string{"admin"},
			wantAdmin: true,
		},
		{
			name:      "scalar non-admin",
			roles:     "user",
			wantRoles: []string{"user"},
			wantAdmin: false,
		},
		{
			name:      "admin match is case-insensitive",
			roles:     []any{"Admin"},
			wantRoles: []string{"Admin"},
			wantAdmin: true,
		},
		{
			name:      "drops empty entries",
			roles:     []any{"", "user", "  "},
			wantRoles: []string{"user"},
			wantAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := mustClaims(t, map[string]any{
				"realm_access": map[string]any{"roles": tt.roles},
			})
			id := NewIdentity(&User{Claims: claims}, "realm_access.roles", "admin")
			if !reflect.DeepEqual(id.Roles, tt.wantRoles) {
				t.Errorf("roles = %#v, want %#v", id.Roles, tt.wantRoles)
			}
			if id.IsAdmin != tt.wantAdmin {
				t.Errorf("is_admin = %v, want %v", id.IsAdmin, tt.wantAdmin)
			}
		})
	}
}

func TestNewIdentity_RolesMissingOrBadPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		claims map[string]any
		path   string
	}{
		{name: "no claims at all", claims: nil, path: "realm_access.roles"},
		{name: "missing nested object", claims: map[string]any{"sub": "x"}, path: "realm_access.roles"},
		{name: "leaf is not array/string", claims: map[string]any{"realm_access": map[string]any{"roles": 42}}, path: "realm_access.roles"},
		{name: "segment not an object", claims: map[string]any{"realm_access": "nope"}, path: "realm_access.roles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewIdentity(&User{Claims: mustClaims(t, tt.claims)}, tt.path, "admin")
			if id.Roles == nil {
				t.Fatal("roles must be non-nil even when absent")
			}
			if len(id.Roles) != 0 {
				t.Errorf("roles = %#v, want empty", id.Roles)
			}
			if id.IsAdmin {
				t.Error("is_admin must be false when no roles")
			}
		})
	}
}

func TestNewIdentity_CustomRolesPath(t *testing.T) {
	t.Parallel()

	claims := mustClaims(t, map[string]any{
		"resource_access": map[string]any{
			"myapp": map[string]any{"roles": []any{"editor", "superuser"}},
		},
	})
	id := NewIdentity(&User{Claims: claims}, "resource_access.myapp.roles", "superuser")
	want := []string{"editor", "superuser"}
	if !reflect.DeepEqual(id.Roles, want) {
		t.Errorf("roles = %#v, want %#v", id.Roles, want)
	}
	if !id.IsAdmin {
		t.Error("expected is_admin via custom admin role")
	}
}

// mustClaims marshals a claims map to JSON; nil yields nil claims.
func mustClaims(t *testing.T, m map[string]any) json.RawMessage {
	t.Helper()
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	return b
}
