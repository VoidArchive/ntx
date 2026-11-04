<script lang="ts">
	import type { Company } from '#lib/gen/ntx/v1/common_pb.js';
	import type { GetCompanyStoryResponse } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		company: Company;
		story?: GetCompanyStoryResponse;
	}

	let { company, story }: Props = $props();

	let claims = $derived(story?.claims ?? []);
	let quote = $derived(story?.quote);

	// Claim text is computed in the API; only its numbers become links to their evidence.
	const NUMBER = /(Rs [\d,]+(?:\.\d+)? crore|Rs [\d,]+(?:\.\d+)?|[\d,]+(?:\.\d+)?(?:%|x))/;

	function fmt(value: number): string {
		return value.toLocaleString('en-NP', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}

	function signed(value: number): string {
		return `${value > 0 ? '+' : ''}${fmt(value)}`;
	}
</script>

<header class="py-4">
	<div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
		<h1 class="font-serif text-3xl font-medium tracking-tight">{company.symbol}</h1>
		<span class="text-lg text-muted-foreground">{company.name}</span>
	</div>

	<p class="mt-4 max-w-3xl text-base leading-relaxed">
		{#if claims.length === 0}
			Not enough filings to say anything yet. Last filing: {story?.lastFiling || 'none'}.
		{:else}
			{#each claims as claim, i (i)}
				{#each claim.text.split(NUMBER) as part, j (j)}
					{#if j % 2 === 1}
						<a
							href="#{claim.anchor}"
							class="underline decoration-dotted underline-offset-4 hover:text-primary">{part}</a
						>
					{:else}{part}{/if}
				{/each}{i < claims.length - 1 ? ' ' : ''}
			{/each}
		{/if}
	</p>

	{#if quote}
		<p class="mt-2 text-sm text-muted-foreground tabular-nums">
			Rs {fmt(quote.close)}
			{#if quote.change !== undefined}
				<span class={quote.change > 0 ? 'text-positive' : quote.change < 0 ? 'text-negative' : ''}>
					{quote.change > 0 ? '▲' : quote.change < 0 ? '▼' : '='}
					{signed(quote.change)}{#if quote.changePercent !== undefined}
						({signed(quote.changePercent)}%){/if}
				</span>
			{/if}
			· close of {quote.businessDate}{#if quote.businessDateBs}
				({quote.businessDateBs}){/if}
		</p>
	{/if}
</header>
