import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import ts from 'typescript-eslint';

// Flat config. This project keeps its SvelteKit config in vite.config.ts (there
// is no svelte.config.js), so we don't pass a svelteConfig to the parser — we
// have no custom preprocessors for it to learn about.
export default ts.config(
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	{
		languageOptions: {
			globals: { ...globals.browser, ...globals.node }
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
		languageOptions: {
			parserOptions: {
				extraFileExtensions: ['.svelte'],
				parser: ts.parser
			}
		},
		rules: {
			// We serve the SPA at the domain root (no base path), and the Button
			// component accepts arbitrary hrefs, so resolve() buys us nothing here.
			'svelte/no-navigation-without-resolve': 'off'
		}
	},
	{
		ignores: [
			'build/',
			'.svelte-kit/',
			'dist/',
			'node_modules/',
			'eslint.config.js',
			'vite.config.ts'
		]
	}
);
