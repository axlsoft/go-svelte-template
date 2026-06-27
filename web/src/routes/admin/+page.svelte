<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, requireAuth } from '$lib/auth';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui';

	let ready = $state(false);

	// NOTE: This is a COSMETIC guard only. It hides the page from non-admins in
	// the UI — it is NOT authorization. `is_admin` comes from the client's copy
	// of /auth/me and must never be trusted for access control. Any real admin
	// endpoint MUST enforce the admin role server-side in Go; the client check is
	// purely for UX (don't show links/screens the user can't use).
	onMount(async () => {
		const ok = await requireAuth('/admin');
		if (!ok) return;
		if (!auth.user?.is_admin) {
			void goto('/account');
			return;
		}
		ready = true;
	});
</script>

{#if ready}
	<section class="space-y-6">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Admin</h1>
			<p class="text-muted-foreground">Placeholder admin area, visible only to admins.</p>
		</div>

		<Card>
			<CardHeader>
				<CardTitle>Server-side enforcement required</CardTitle>
				<CardDescription>This screen is gated client-side for UX only.</CardDescription>
			</CardHeader>
			<CardContent class="text-muted-foreground space-y-2 text-sm">
				<p>
					Real admin functionality belongs behind Go endpoints that verify the admin role on every
					request. Treat this client-side guard as cosmetic.
				</p>
				<p>Build admin features here once the backing endpoints enforce the role.</p>
			</CardContent>
		</Card>
	</section>
{/if}
