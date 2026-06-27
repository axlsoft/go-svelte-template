import { browser } from '$app/environment';
import { ApiError, errorBodyFromEnvelope } from './errors';

// The Go CSRF middleware uses a signed double-submit token: it sets a readable
// cookie and expects the same value echoed in this header on mutating requests.
// These must match internal/auth/csrf.go.
const CSRF_COOKIE = 'myapp_csrf';
const CSRF_HEADER = 'X-CSRF-Token';
const MUTATING_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

// LOGIN_PATH is the Go-owned endpoint that starts the OIDC auth-code flow.
const LOGIN_PATH = '/auth/login';

// readCsrfToken returns the current CSRF token from cookies, or null. Pure
// string parsing — exported for testing.
export function readCsrfToken(cookieHeader: string): string | null {
	for (const part of cookieHeader.split(';')) {
		const [rawName, ...rawValue] = part.trim().split('=');
		if (rawName === CSRF_COOKIE) {
			return decodeURIComponent(rawValue.join('='));
		}
	}
	return null;
}

// onUnauthorized sends the browser through the login flow, preserving where the
// user was trying to go so they land back there after authenticating.
function onUnauthorized(): void {
	if (!browser) return;
	const returnTo = window.location.pathname + window.location.search;
	const url = `${LOGIN_PATH}?return_to=${encodeURIComponent(returnTo)}`;
	window.location.assign(url);
}

export interface ApiRequestOptions extends Omit<RequestInit, 'body'> {
	// json is serialized to the request body and sets the Content-Type header.
	json?: unknown;
	body?: BodyInit | null;
	// redirectOnUnauthorized defaults to true; set false to handle 401 yourself.
	redirectOnUnauthorized?: boolean;
}

// apiFetch is the single entry point for talking to the Go backend. It:
//   - serializes `json` bodies,
//   - attaches the CSRF token on mutating requests,
//   - sends same-origin credentials (the session cookie),
//   - throws a typed ApiError mirroring the Go envelope on failure,
//   - redirects to the login flow on 401 (unless opted out).
export async function apiFetch<T = unknown>(
	path: string,
	options: ApiRequestOptions = {}
): Promise<T> {
	const { json, redirectOnUnauthorized = true, headers, body, ...rest } = options;
	const method = (rest.method ?? 'GET').toUpperCase();
	const finalHeaders = new Headers(headers);
	let finalBody = body ?? null;

	if (json !== undefined) {
		finalHeaders.set('Content-Type', 'application/json');
		finalBody = JSON.stringify(json);
	}

	if (MUTATING_METHODS.has(method) && browser) {
		const token = readCsrfToken(document.cookie);
		if (token) finalHeaders.set(CSRF_HEADER, token);
	}

	const res = await fetch(path, {
		...rest,
		method,
		headers: finalHeaders,
		body: finalBody,
		credentials: 'same-origin'
	});

	if (res.status === 401 && redirectOnUnauthorized) {
		onUnauthorized();
	}

	if (!res.ok) {
		throw new ApiError(res.status, errorBodyFromEnvelope(res.status, await safeJson(res)));
	}

	if (res.status === 204) {
		return undefined as T;
	}

	const contentType = res.headers.get('content-type') ?? '';
	if (contentType.includes('application/json')) {
		return (await res.json()) as T;
	}
	return (await res.text()) as unknown as T;
}

async function safeJson(res: Response): Promise<unknown> {
	try {
		return await res.json();
	} catch {
		return null;
	}
}
