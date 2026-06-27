// shadcn-svelte-style DropdownMenu: thin styled wrappers over the bits-ui menu
// primitive (keyboard nav, focus management and portalling come from bits-ui).
import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';

import Content from './dropdown-menu-content.svelte';
import Item from './dropdown-menu-item.svelte';
import Separator from './dropdown-menu-separator.svelte';
import Label from './dropdown-menu-label.svelte';

const Root = DropdownMenuPrimitive.Root;
const Trigger = DropdownMenuPrimitive.Trigger;
const Group = DropdownMenuPrimitive.Group;

export {
	Root as DropdownMenu,
	Trigger as DropdownMenuTrigger,
	Content as DropdownMenuContent,
	Item as DropdownMenuItem,
	Separator as DropdownMenuSeparator,
	Label as DropdownMenuLabel,
	Group as DropdownMenuGroup
};
