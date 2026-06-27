// Auth store + route guard. The store hydrates from /auth/me; the guard sends
// anonymous users into the OIDC login flow.
export { auth } from './auth.svelte';
export { requireAuth } from './guard';
