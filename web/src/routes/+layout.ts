// Client-only SPA. No SSR, no prerender: every route is served by index.html and
// the client router takes over, so refreshing a deep link returns the app (not a
// 404). This pairs with adapter-static `fallback: 'index.html'` and the Go embed.
//
// See web/AGENTS.md — there is no Node server, so there are no *.server.ts loads.
export const ssr = false;
export const prerender = false;
