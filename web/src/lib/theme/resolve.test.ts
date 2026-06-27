import { describe, expect, it } from 'vitest';
import { normalizePreference, resolveEffective } from './resolve';

describe('normalizePreference', () => {
	it('accepts valid preferences', () => {
		expect(normalizePreference('light')).toBe('light');
		expect(normalizePreference('dark')).toBe('dark');
		expect(normalizePreference('system')).toBe('system');
	});

	it('defaults to system for anything else', () => {
		expect(normalizePreference(null)).toBe('system');
		expect(normalizePreference(undefined)).toBe('system');
		expect(normalizePreference('')).toBe('system');
		expect(normalizePreference('bogus')).toBe('system');
	});
});

describe('resolveEffective', () => {
	it('system follows the OS', () => {
		expect(resolveEffective('system', true)).toBe('dark');
		expect(resolveEffective('system', false)).toBe('light');
	});

	it('explicit choices ignore the OS', () => {
		expect(resolveEffective('light', true)).toBe('light');
		expect(resolveEffective('dark', false)).toBe('dark');
	});
});
