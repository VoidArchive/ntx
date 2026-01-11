<script lang="ts">
	import SearchCommand from '$lib/components/SearchCommand.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		companies?: Company[];
		prices?: Price[];
	}

	let { companies = [], prices = [] }: Props = $props();
</script>

<nav class="border-b border-border/60 bg-background">
	<div class="mx-auto max-w-7xl px-2 py-4">
		<div class="flex items-center justify-between gap-4">
			<!-- Logo -->
			<a href="/" class="font-serif text-xl tracking-tight transition-opacity hover:opacity-80">
				NTX
			</a>

			<!-- Search + Theme -->
			<div class="flex items-center gap-2">
				<SearchCommand {companies} {prices} variant="compact" placeholder="Search stocks..." />
				<ThemeToggle />
			</div>
		</div>
	</div>
</nav>
