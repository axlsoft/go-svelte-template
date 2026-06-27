import { describe, expect, it } from 'vitest';
import { initialsFrom } from './initials';

describe('initialsFrom', () => {
	it('uses first + last initial of a multi-word name', () => {
		expect(initialsFrom('Ada Lovelace')).toBe('AL');
		expect(initialsFrom('Grace Brewster Murray Hopper')).toBe('GH');
	});

	it('uses the first two letters of a single-word name', () => {
		expect(initialsFrom('ada')).toBe('AD');
		expect(initialsFrom('x')).toBe('X');
	});

	it('collapses extra whitespace', () => {
		expect(initialsFrom('  Ada   Lovelace  ')).toBe('AL');
	});

	it('falls back to the email first character when no name', () => {
		expect(initialsFrom('', 'ada@example.com')).toBe('A');
		expect(initialsFrom(null, 'zoe@example.com')).toBe('Z');
	});

	it('falls back to ? when nothing is available', () => {
		expect(initialsFrom()).toBe('?');
		expect(initialsFrom('', '')).toBe('?');
		expect(initialsFrom(null, null)).toBe('?');
	});

	it('is deterministic', () => {
		expect(initialsFrom('Ada Lovelace')).toBe(initialsFrom('Ada Lovelace'));
	});
});
