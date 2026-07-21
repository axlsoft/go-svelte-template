<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = { class?: string; children?: Snippet } & HTMLAttributes<HTMLDivElement>;
	let { class: className, children, ...rest }: Props = $props();
</script>

<div class={cn('bg-card text-card-foreground rounded-lg border shadow-sm', className)} {...rest}>
	{@render children?.()}
</div>
