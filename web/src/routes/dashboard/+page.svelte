<script lang="ts">
	import { onMount } from 'svelte';
	import { apiFetch, ApiError, type User } from '$lib/api';
	import { auth, requireAuth } from '$lib/auth';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui';

	let ready = $state(false);
	let apiUser = $state<User | null>(null);
	let apiError = $state<string | null>(null);

	onMount(async () => {
		// Guard: anonymous users are sent through the login flow and returned here.
		const ok = await requireAuth('/dashboard');
		if (!ok) return;
		ready = true;

		// Demonstrate the guarded API: GET /api/me requires a valid session.
		try {
			apiUser = await apiFetch<User>('/api/me', { redirectOnUnauthorized: false });
		} catch (err) {
			apiError = err instanceof ApiError ? `${err.code}: ${err.message}` : String(err);
		}
	});
</script>

{#if !ready}
	<p class="text-muted-foreground">Checking your session…</p>
{:else}
	<section class="space-y-6">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
			<p class="text-muted-foreground">You're signed in. This route is session-protected.</p>
		</div>

		<div class="grid gap-4 sm:grid-cols-2">
			<Card>
				<CardHeader>
					<CardTitle>Session (/auth/me)</CardTitle>
					<CardDescription>Hydrated by the auth store.</CardDescription>
				</CardHeader>
				<CardContent class="space-y-1 text-sm">
					<p><span class="text-muted-foreground">Name:</span> {auth.user?.name}</p>
					<p><span class="text-muted-foreground">Email:</span> {auth.user?.email}</p>
					<p class="break-all">
						<span class="text-muted-foreground">Subject:</span>
						{auth.user?.sub}
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>Guarded API (/api/me)</CardTitle>
					<CardDescription>Fetched with the typed client.</CardDescription>
				</CardHeader>
				<CardContent class="space-y-1 text-sm">
					{#if apiError}
						<p class="text-destructive">{apiError}</p>
					{:else if apiUser}
						<p><span class="text-muted-foreground">Email:</span> {apiUser.email}</p>
						<p class="text-muted-foreground">200 OK — the RequireAuth guard passed.</p>
					{:else}
						<p class="text-muted-foreground">Loading…</p>
					{/if}
				</CardContent>
			</Card>
		</div>
	</section>
{/if}
