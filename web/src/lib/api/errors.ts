import type { ApiErrorBody, ApiErrorEnvelope } from './types';

// ApiError is thrown by the client for any non-2xx response. It carries the
// HTTP status plus the structured fields from the Go error envelope.
export class ApiError extends Error {
	readonly status: number;
	readonly code: string;
	readonly requestId?: string;

	constructor(status: number, body: ApiErrorBody) {
		super(body.message || `request failed (${status})`);
		this.name = 'ApiError';
		this.status = status;
		this.code = body.code || 'unknown';
		this.requestId = body.request_id;
	}
}

// isApiErrorEnvelope is a pure type guard for the Go error envelope. Exported so
// it can be unit-tested without touching the network.
export function isApiErrorEnvelope(value: unknown): value is ApiErrorEnvelope {
	if (typeof value !== 'object' || value === null) return false;
	const err = (value as Record<string, unknown>).error;
	if (typeof err !== 'object' || err === null) return false;
	const body = err as Record<string, unknown>;
	return typeof body.code === 'string' && typeof body.message === 'string';
}

// errorBodyFromEnvelope extracts the inner body, falling back to a synthetic
// body when the payload isn't a recognizable envelope. Pure and testable.
export function errorBodyFromEnvelope(status: number, payload: unknown): ApiErrorBody {
	if (isApiErrorEnvelope(payload)) {
		return payload.error;
	}
	return { code: 'unknown', message: `request failed (${status})` };
}
