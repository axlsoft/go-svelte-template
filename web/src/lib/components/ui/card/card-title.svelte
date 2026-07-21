<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = { class?: string; children?: Snippet } & HTMLAttributes<HTMLHeadingElement>;
	let { class: className, children, ...rest }: Props = $props();
</script>

<h3 class={cn('text-lg leading-none font-semibold tracking-tight', className)} {...rest}>
	{@render children?.()}
</h3>
