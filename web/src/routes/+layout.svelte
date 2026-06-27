<script lang="ts">
	import { onMount } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import ProfileMenu from '$lib/components/profile-menu.svelte';
	import { auth } from '$lib/auth';
	import { theme } from '$lib/theme';
	import '../app.css';

	let { children } = $props();

	// Hydrate the session and the theme once on first load. The anti-flash script
	// in app.html has already applied the theme before paint; theme.init() takes
	// over to keep it in sync with OS changes and user choices.
	onMount(() => {
		theme.init();
		void auth.load();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="flex min-h-screen flex-col">
	<header class="border-b">
		<nav class="mx-auto flex h-14 w-full max-w-5xl items-center justify-between px-4">
			<a href="/" class="text-base font-semibold">myapp</a>
			<div class="flex items-center gap-4">
				{#if auth.status === 'authenticated'}
					<a href="/dashboard" class="text-muted-foreground hover:text-foreground text-sm">
						Dashboard
					</a>
				{/if}
				<ProfileMenu />
			</div>
		</nav>
	</header>

	<main class="mx-auto w-full max-w-5xl flex-1 px-4 py-10">
		{@render children()}
	</main>
</div>
