<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';

	let { class: className, children, ...rest }: DropdownMenuPrimitive.ItemProps = $props();
</script>

<DropdownMenuPrimitive.Item
	class={cn(
		'relative flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none select-none',
		'data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground',
		'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
		className
	)}
	{...rest}
>
	{@render children?.()}
</DropdownMenuPrimitive.Item>
