<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';

	type Props = { class?: string; children?: Snippet } & HTMLAttributes<HTMLParagraphElement>;
	let { class: className, children, ...rest }: Props = $props();
</script>

<p class={cn('text-muted-foreground text-sm', className)} {...rest}>
	{@render children?.()}
</p>
