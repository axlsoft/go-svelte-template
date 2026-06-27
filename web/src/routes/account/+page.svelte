<script lang="ts">
	import { auth } from '$lib/auth';
	import {
		Avatar,
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui';

	const u = $derived(auth.user);

	function formatSince(ts?: string): string {
		if (!ts) return '—';
		const d = new Date(ts);
		return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString();
	}
</script>

<Card>
	<CardHeader>
		<CardTitle>Profile</CardTitle>
		<CardDescription>Read-only — sourced from your identity provider.</CardDescription>
	</CardHeader>
	<CardContent class="space-y-6">
		{#if u}
			<div class="flex items-center gap-4">
				<Avatar src={u.picture} name={u.name} email={u.email} class="h-14 w-14" />
				<div class="min-w-0">
					<p class="truncate text-lg font-medium">{u.name}</p>
					{#if u.preferred_username}
						<p class="text-muted-foreground truncate text-sm">@{u.preferred_username}</p>
					{/if}
				</div>
			</div>

			<dl class="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-2">
				<div>
					<dt class="text-muted-foreground text-xs tracking-wide uppercase">Display name</dt>
					<dd class="text-sm">{u.name || '—'}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground text-xs tracking-wide uppercase">Email</dt>
					<dd class="text-sm break-all">{u.email || '—'}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground text-xs tracking-wide uppercase">Username</dt>
					<dd class="text-sm">{u.preferred_username || '—'}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground text-xs tracking-wide uppercase">Signed in since</dt>
					<dd class="text-sm">{formatSince(u.signed_in_since)}</dd>
				</div>
				<div class="sm:col-span-2">
					<dt class="text-muted-foreground text-xs tracking-wide uppercase">Roles</dt>
					<dd class="mt-1 flex flex-wrap items-center gap-1.5">
						{#if u.roles.length}
							{#each u.roles as role (role)}
								<span
									class="bg-secondary text-secondary-foreground rounded-full px-2 py-0.5 text-xs font-medium"
								>
									{role}
								</span>
							{/each}
						{:else}
							<span class="text-sm">—</span>
						{/if}
						{#if u.is_admin}
							<span
								class="bg-primary text-primary-foreground rounded-full px-2 py-0.5 text-xs font-medium"
							>
								admin
							</span>
						{/if}
					</dd>
				</div>
				<div class="sm:col-span-2">
					<dt class="text-muted-foreground text-xs tracking-wide uppercase">Subject (sub)</dt>
					<dd class="font-mono text-xs break-all">{u.sub}</dd>
				</div>
			</dl>
		{/if}
	</CardContent>
</Card>
