import { ApiError, apiFetch } from '$lib/api';
import type { LogoutResponse, User } from '$lib/api';

// AuthState is hydrated once from GET /auth/me and then kept in sync by the
// login/logout actions. It's a runes-based singleton so any component can read
// `auth.user` / `auth.status` reactively.
//
// SvelteKit here is client-only (see web/AGENTS.md) — there is no server load,
// so the session is resolved by calling the Go backend from the browser.
type AuthStatus = 'unknown' | 'authenticated' | 'anonymous';

class AuthState {
	user = $state<User | null>(null);
	status = $state<AuthStatus>('unknown');

	get isAuthenticated(): boolean {
		return this.status === 'authenticated';
	}

	// load resolves the current session. A 401 means "not logged in" (not an
	// error); anything else propagates.
	async load(): Promise<void> {
		try {
			const user = await apiFetch<User>('/auth/me', { redirectOnUnauthorized: false });
			this.user = user;
			this.status = 'authenticated';
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				this.user = null;
				this.status = 'anonymous';
				return;
			}
			throw err;
		}
	}

	// login sends the browser into the OIDC auth-code flow, returning to
	// `returnTo` (defaults to the current location) afterwards.
	login(returnTo?: string): void {
		const target =
			returnTo ??
			(typeof window !== 'undefined' ? window.location.pathname + window.location.search : '/');
		window.location.assign(`/auth/login?return_to=${encodeURIComponent(target)}`);
	}

	// logout destroys the server session and follows the IdP logout URL when the
	// backend provides one.
	async logout(): Promise<void> {
		let logoutUrl = '/';
		try {
			const res = await apiFetch<LogoutResponse>('/auth/logout', {
				method: 'POST',
				redirectOnUnauthorized: false
			});
			if (res?.logout_url) logoutUrl = res.logout_url;
		} finally {
			this.user = null;
			this.status = 'anonymous';
		}
		if (typeof window !== 'undefined') window.location.assign(logoutUrl);
	}
}

export const auth = new AuthState();
