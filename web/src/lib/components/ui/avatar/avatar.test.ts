import { describe, expect, it } from 'vitest';
import { fireEvent, render } from '@testing-library/svelte';
import Avatar from './avatar.svelte';

describe('Avatar', () => {
	it('renders the image when a src is provided', () => {
		const { container } = render(Avatar, {
			src: 'https://img.example/ada.png',
			name: 'Ada Lovelace'
		});
		const img = container.querySelector('img');
		expect(img).not.toBeNull();
		expect(img?.getAttribute('src')).toBe('https://img.example/ada.png');
	});

	it('falls back to initials when the image fails to load', async () => {
		const { container } = render(Avatar, {
			src: 'https://img.example/broken.png',
			name: 'Ada Lovelace'
		});
		const img = container.querySelector('img');
		expect(img).not.toBeNull();

		await fireEvent.error(img!);

		expect(container.querySelector('img')).toBeNull();
		expect(container.textContent).toContain('AL');
	});

	it('shows initials when there is no src', () => {
		const { container } = render(Avatar, { name: 'Grace Hopper' });
		expect(container.querySelector('img')).toBeNull();
		expect(container.textContent).toContain('GH');
	});

	it('uses the email initial when no name', () => {
		const { container } = render(Avatar, { email: 'zoe@example.com' });
		expect(container.textContent).toContain('Z');
	});
});
