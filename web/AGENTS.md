# web/ — SvelteKit SPA

**This is a routing + build tool only.** SvelteKit here uses
`@sveltejs/adapter-static` in SPA mode (`fallback: 'index.html'`,
`prerender = false`). There is **no Node server** in production — the built
assets are embedded into the Go binary.

Therefore, **do not** use any server-side SvelteKit features:

- no `+page.server.ts` / `+layout.server.ts`
- no form actions
- no `hooks.server.ts`
- no `$app/server`

Everything server-side lives in **Go** (`/api`, `/auth`). The frontend talks to
it over HTTP via a typed client (`src/lib/api`).
