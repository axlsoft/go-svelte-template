import { auth } from './auth.svelte';

// requireAuth is the route guard for protected pages. It ensures the session is
// resolved, and if the user is anonymous, sends them through the login flow
// (returning to `returnTo` afterwards). Returns true when authenticated.
//
// Use it from a protected page's onMount (the app is client-only, so there is no
// server-side load to guard in).
export async function requireAuth(returnTo?: string): Promise<boolean> {
	if (auth.status === 'unknown') {
		await auth.load();
	}
	if (!auth.isAuthenticated) {
		auth.login(returnTo);
		return false;
	}
	return true;
}
