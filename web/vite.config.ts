import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';

// SvelteKit here is a routing + build tool only (see web/AGENTS.md). We use
// adapter-static in SPA mode: every unknown path falls back to index.html and
// the client router takes over. There is no Node server in production — the
// built assets are embedded into the Go binary.
//
// Dev model: the user always hits the Go server (:8080). Go serves /api + /auth
// itself and reverse-proxies everything else to this Vite dev server for HMR.
// Setting hmr.clientPort to the Go port makes the browser open its HMR
// websocket against Go, which proxies it through to Vite.
const GO_PORT = Number(process.env.GO_PORT ?? 8080);
const VITE_PORT = Number(process.env.VITE_PORT ?? 5173);

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// SPA: serve index.html for any route the static files don't match.
			adapter: adapter({ fallback: 'index.html' })
		}),
		// Test-only (no-op outside vitest): resolves Svelte's browser build so
		// component tests can mount, and handles DOM cleanup between tests.
		svelteTesting()
	],
	server: {
		port: VITE_PORT,
		strictPort: true,
		// The browser loads the app from the Go server, so point HMR there.
		hmr: { clientPort: GO_PORT }
	},
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
