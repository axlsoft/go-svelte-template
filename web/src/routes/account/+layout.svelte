<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { requireAuth } from '$lib/auth';
	import { cn } from '$lib/utils';

	let { children } = $props();

	let ready = $state(false);

	// Settings sub-nav. New settings areas slot in here as additional entries.
	const nav = [
		{ href: '/account', label: 'Profile' },
		{ href: '/account/preferences', label: 'Preferences' }
	];

	onMount(async () => {
		const ok = await requireAuth(page.url.pathname);
		if (ok) ready = true;
	});
</script>

{#if !ready}
	<p class="text-muted-foreground">Checking your session…</p>
{:else}
	<div class="space-y-6">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Account</h1>
			<p class="text-muted-foreground">Manage your profile and preferences.</p>
		</div>

		<div class="flex flex-col gap-8 sm:flex-row">
			<aside class="shrink-0 sm:w-48">
				<nav class="flex gap-1 sm:flex-col" aria-label="Account settings">
					{#each nav as item (item.href)}
						<a
							href={item.href}
							aria-current={page.url.pathname === item.href ? 'page' : undefined}
							class={cn(
								'rounded-md px-3 py-2 text-sm',
								page.url.pathname === item.href
									? 'bg-secondary text-secondary-foreground font-medium'
									: 'text-muted-foreground hover:text-foreground'
							)}
						>
							{item.label}
						</a>
					{/each}
				</nav>
			</aside>

			<div class="min-w-0 flex-1">
				{@render children()}
			</div>
		</div>
	</div>
{/if}
