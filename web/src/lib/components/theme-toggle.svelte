<script lang="ts">
	import { Monitor, Moon, Sun } from '@lucide/svelte';
	import { theme, type ThemePreference } from '$lib/theme';
	import { cn } from '$lib/utils';

	// Three-way Light / Dark / System segmented control. Reflects the current
	// preference and writes through the theme store. This is the only theme
	// control in the app (it lives in account → Preferences).
	const options: { value: ThemePreference; label: string; icon: typeof Sun }[] = [
		{ value: 'light', label: 'Light', icon: Sun },
		{ value: 'dark', label: 'Dark', icon: Moon },
		{ value: 'system', label: 'System', icon: Monitor }
	];
</script>

<div role="radiogroup" aria-label="Theme" class="inline-flex gap-1 rounded-md border p-1">
	{#each options as opt (opt.value)}
		{@const Icon = opt.icon}
		<button
			type="button"
			role="radio"
			aria-checked={theme.preference === opt.value}
			aria-label={opt.label}
			class={cn(
				'focus-visible:ring-ring flex items-center gap-2 rounded-sm px-3 py-1.5 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none',
				theme.preference === opt.value
					? 'bg-secondary text-secondary-foreground'
					: 'text-muted-foreground hover:text-foreground'
			)}
			onclick={() => theme.set(opt.value)}
		>
			<Icon class="size-4" />
			{opt.label}
		</button>
	{/each}
</div>
