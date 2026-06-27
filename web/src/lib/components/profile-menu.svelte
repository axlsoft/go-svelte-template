<script lang="ts">
	import { CircleUser, LogOut, Shield } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import {
		Avatar,
		Button,
		DropdownMenu,
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuLabel,
		DropdownMenuSeparator,
		DropdownMenuTrigger
	} from '$lib/components/ui';

	// Header profile control: an avatar button that opens a menu. The theme
	// control deliberately lives in account → Preferences, never here.
</script>

{#if auth.status === 'authenticated' && auth.user}
	<DropdownMenu>
		<DropdownMenuTrigger
			class="focus-visible:ring-ring rounded-full focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
			aria-label="Open account menu"
		>
			<Avatar src={auth.user.picture} name={auth.user.name} email={auth.user.email} />
		</DropdownMenuTrigger>
		<DropdownMenuContent>
			<DropdownMenuLabel>
				<div class="flex flex-col">
					<span class="truncate text-sm font-medium">{auth.user.name}</span>
					<span class="text-muted-foreground truncate text-xs font-normal">{auth.user.email}</span>
				</div>
			</DropdownMenuLabel>
			<DropdownMenuSeparator />
			<DropdownMenuItem onSelect={() => goto('/account')}>
				<CircleUser class="size-4" /> My account
			</DropdownMenuItem>
			{#if auth.user.is_admin}
				<DropdownMenuItem onSelect={() => goto('/admin')}>
					<Shield class="size-4" /> Admin
				</DropdownMenuItem>
			{/if}
			<DropdownMenuSeparator />
			<DropdownMenuItem onSelect={() => auth.logout()}>
				<LogOut class="size-4" /> Log out
			</DropdownMenuItem>
		</DropdownMenuContent>
	</DropdownMenu>
{:else if auth.status === 'anonymous'}
	<Button size="sm" onclick={() => auth.login('/dashboard')}>Log in</Button>
{/if}
