<script lang="ts">
	import ThemeToggle from '$lib/components/theme-toggle.svelte';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui';
</script>

<Card>
	<CardHeader>
		<CardTitle>Preferences</CardTitle>
		<CardDescription>
			Personalize how myapp looks. More preferences will appear here over time.
		</CardDescription>
	</CardHeader>
	<CardContent class="space-y-6">
		<div class="space-y-2">
			<div>
				<p class="text-sm font-medium">Theme</p>
				<p class="text-muted-foreground text-sm">
					Choose a light or dark appearance, or follow your system.
				</p>
			</div>
			<ThemeToggle />
		</div>
	</CardContent>
</Card>
