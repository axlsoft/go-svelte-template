<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';

	let {
		class: className,
		sideOffset = 4,
		align = 'end',
		children,
		...rest
	}: DropdownMenuPrimitive.ContentProps = $props();
</script>

<DropdownMenuPrimitive.Content
	{sideOffset}
	{align}
	class={cn(
		'bg-popover text-popover-foreground z-50 min-w-[12rem] overflow-hidden rounded-md border p-1 shadow-md',
		className
	)}
	{...rest}
>
	{@render children?.()}
</DropdownMenuPrimitive.Content>
