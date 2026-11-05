<script lang="ts">
	import type { Company } from '#lib/gen/ntx/v1/common_pb.js';
	import type { GetCompanyStoryResponse } from '#lib/gen/ntx/v1/story_pb.js';
	import { sectorNames } from '#lib/utils.js';
	import ClaimText from './ClaimText.svelte';

	interface Props {
		company: Company;
		story?: GetCompanyStoryResponse;
	}

	let { company, story }: Props = $props();

	// The lede carries the first earnings sentence; every claim heads its own section below.
	let lede = $derived((story?.claims ?? []).filter((c) => c.kind === 'earnings').slice(0, 1));
	let quote = $derived(story?.quote);
	let rating = $derived(story?.rating);

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
		{#if rating?.letter}
			<!-- The letter alone carries the meaning; no colour (GLOSSARY: Rating). -->
			<span
				class="inline-flex size-8 items-center justify-center self-center rounded border border-foreground font-serif text-lg font-medium"
				title="{rating.holds} of {rating.evaluated} checks hold"
				aria-label="Rating {rating.letter}: {rating.holds} of {rating.evaluated} checks hold"
			>
				{rating.letter}
			</span>
		{:else if rating && story}
			<span
				class="text-xs text-muted-foreground"
				title="{rating.evaluated} of {story.checks
					.length} checks have enough history; too few to rate">Unrated</span
			>
		{/if}
		<span class="text-lg text-muted-foreground">{company.name}</span>
		<span class="text-sm text-muted-foreground">{sectorNames[company.sector] ?? 'Others'}</span>
	</div>

	<p class="mt-4 max-w-3xl text-lg leading-relaxed">
		{#if lede.length === 0}
			Not enough filings to say anything yet. Last filing: {story?.lastFiling || 'none'}.
		{:else}
			<ClaimText claims={lede} />
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
