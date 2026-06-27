<script lang="ts">
	import {
		Button,
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui';
	import { auth } from '$lib/auth';
</script>

<section class="space-y-8">
	<div class="space-y-3">
		<h1 class="text-3xl font-bold tracking-tight">Go + Svelte single-binary scaffold</h1>
		<p class="max-w-2xl text-muted-foreground">
			A Svelte 5 SPA embedded into one static Go binary — in-app OIDC, Postgres, embedded
			migrations, and a typed API client that speaks the Go error envelope.
		</p>
		<div class="flex gap-3">
			{#if auth.status === 'authenticated'}
				<Button href="/dashboard">Go to dashboard</Button>
			{:else}
				<Button onclick={() => auth.login('/dashboard')}>Log in</Button>
			{/if}
			<Button variant="outline" href="/dashboard">Open dashboard</Button>
		</div>
	</div>

	<Card class="max-w-md">
		<CardHeader>
			<CardTitle>Protected routes</CardTitle>
			<CardDescription>The dashboard is guarded by the session.</CardDescription>
		</CardHeader>
		<CardContent class="text-sm text-muted-foreground">
			Visiting <code>/dashboard</code> while signed out redirects you through the OIDC login flow and
			returns you here afterwards.
		</CardContent>
	</Card>
</section>
