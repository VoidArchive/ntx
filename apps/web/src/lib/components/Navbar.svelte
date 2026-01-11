<script lang="ts">
	import SearchCommand from './SearchCommand.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import { page } from '$app/stores';

	interface Props {
		companies?: Company[];
		prices?: Price[];
	}

	let { companies = [], prices = [] }: Props = $props();

	let isMutualFunds = $derived($page.url.pathname.startsWith('/mutual-funds'));
</script>

<nav class="sticky top-0 z-50 border-b border-border bg-background/80 backdrop-blur-sm">
	<div class="mx-auto max-w-7xl px-4">
		<div class="flex h-14 items-center justify-between gap-4">
			<!-- Logo + Nav -->
			<div class="flex items-center gap-6">
				<a href="/" class="font-serif text-xl tracking-tight transition-opacity hover:opacity-80">
					NTX
				</a>

				<a
					href="/mutual-funds"
					class="text-sm underline underline-offset-4 transition-colors {isMutualFunds
						? 'text-foreground'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					Mutual Funds
				</a>
			</div>

			<!-- Search + Theme -->
			<div class="flex items-center gap-2">
				{#if companies.length > 0}
					<SearchCommand {companies} {prices} variant="compact" placeholder="Search stocks..." />
				{/if}
				<ThemeToggle />
			</div>
		</div>
	</div>
</nav>
