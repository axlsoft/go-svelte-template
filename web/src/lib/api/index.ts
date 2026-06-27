// Typed client for the Go backend (/api, /auth). See web/AGENTS.md: all
// server-side logic lives in Go; this is the browser's view of it.
export { apiFetch, readCsrfToken } from './client';
export type { ApiRequestOptions } from './client';
export { ApiError, isApiErrorEnvelope, errorBodyFromEnvelope } from './errors';
export type { ApiErrorBody, ApiErrorEnvelope, User, LogoutResponse } from './types';
