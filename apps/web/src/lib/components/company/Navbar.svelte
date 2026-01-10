<script lang="ts">
	import SearchCommand from '$lib/components/SearchCommand.svelte';
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		companies?: Company[];
		prices?: Price[];
	}

	let { companies = [], prices = [] }: Props = $props();
</script>

<nav class="border-b border-border/60 bg-background">
	<div class="mx-auto max-w-6xl px-6 py-4">
		<div class="flex items-center justify-between gap-4">
			<!-- Logo -->
			<a href="/" class="flex items-center gap-2 transition-opacity hover:opacity-80">
				<img src="/logo.png" alt="NTX" class="size-6" />
				<span class="font-serif text-xl tracking-tight">NTX</span>
			</a>

			<!-- Search -->
			<SearchCommand {companies} {prices} variant="compact" placeholder="Search stocks..." />
		</div>
	</div>
</nav>
