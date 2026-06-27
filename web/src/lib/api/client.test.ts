import { describe, expect, it } from 'vitest';
import { errorBodyFromEnvelope, isApiErrorEnvelope } from './errors';
import { readCsrfToken } from './client';

describe('isApiErrorEnvelope', () => {
	it('accepts a well-formed Go error envelope', () => {
		const payload = {
			error: { code: 'unauthorized', message: 'authentication required', request_id: 'abc' }
		};
		expect(isApiErrorEnvelope(payload)).toBe(true);
	});

	it('rejects payloads missing the error object', () => {
		expect(isApiErrorEnvelope({})).toBe(false);
		expect(isApiErrorEnvelope(null)).toBe(false);
		expect(isApiErrorEnvelope('nope')).toBe(false);
		expect(isApiErrorEnvelope({ error: { code: 1, message: 'x' } })).toBe(false);
	});
});

describe('errorBodyFromEnvelope', () => {
	it('returns the inner body for a valid envelope', () => {
		const body = errorBodyFromEnvelope(403, {
			error: { code: 'forbidden', message: 'nope', request_id: 'r1' }
		});
		expect(body).toEqual({ code: 'forbidden', message: 'nope', request_id: 'r1' });
	});

	it('synthesizes a body for unrecognized payloads', () => {
		const body = errorBodyFromEnvelope(500, 'boom');
		expect(body.code).toBe('unknown');
		expect(body.message).toContain('500');
	});
});

describe('readCsrfToken', () => {
	it('extracts the CSRF cookie value', () => {
		expect(readCsrfToken('foo=1; myapp_csrf=tok123; bar=2')).toBe('tok123');
	});

	it('returns null when absent', () => {
		expect(readCsrfToken('foo=1; bar=2')).toBeNull();
		expect(readCsrfToken('')).toBeNull();
	});

	it('url-decodes the value', () => {
		expect(readCsrfToken('myapp_csrf=a%2Bb')).toBe('a+b');
	});
});
