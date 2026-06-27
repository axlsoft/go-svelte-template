<script lang="ts">
	import { cn } from '$lib/utils';
	import { initialsFrom } from './initials';

	// Avatar renders the OIDC `picture` when present and falls back to initials —
	// both when no picture is given and when the image fails to load. Initials are
	// derived deterministically from the name (then the email).
	type Props = {
		src?: string | null;
		name?: string | null;
		email?: string | null;
		alt?: string;
		class?: string;
	};

	let { src, name, email, alt, class: className }: Props = $props();

	let failed = $state(false);

	// Reset the failure flag whenever the source changes so a new picture gets a
	// fresh chance to load.
	$effect(() => {
		void src;
		failed = false;
	});

	const initials = $derived(initialsFrom(name, email));
	const showImage = $derived(!!src && !failed);
</script>

<span
	class={cn(
		'bg-muted relative flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-full',
		className
	)}
>
	{#if showImage}
		<img
			src={src ?? undefined}
			alt={alt ?? name ?? email ?? 'avatar'}
			class="aspect-square h-full w-full object-cover"
			onerror={() => (failed = true)}
		/>
	{:else}
		<span class="text-muted-foreground text-sm font-medium select-none">{initials}</span>
	{/if}
</span>
