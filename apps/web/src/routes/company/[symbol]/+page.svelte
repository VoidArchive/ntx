<script lang="ts">
	import { generateStory } from '$lib/story';
	import type { CompanyStory } from '$lib/story';
	import { Sector, CompanyStatus } from '$lib/gen/ntx/v1/common_pb';

	let { data } = $props();
	let company = $derived(data.company);
	let fundamentals = $derived(data.fundamentals);
	let priceData = $derived(data.price);
	let priceHistory = $derived(data.priceHistory);
	let sectorStats = $derived(data.sectorStats);

	// Generate story from data
	let story = $derived.by(() => {
		if (!company || !priceData || !fundamentals) return null;
		return generateStory({
			company,
			price: priceData,
			priceHistory: priceHistory ?? [],
			fundamentals,
			fundamentalsHistory: data.fundamentalsHistory ?? [],
			sectorStats
		});
	});

	// Sector enum to display name mapping
	const sectorDisplayNames: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Commercial Banking',
		[Sector.DEVELOPMENT_BANK]: 'Development Banking',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'Microfinance',
		[Sector.LIFE_INSURANCE]: 'Life Insurance',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
		[Sector.HYDROPOWER]: 'Hydropower',
		[Sector.MANUFACTURING]: 'Manufacturing',
		[Sector.HOTEL]: 'Hotels & Tourism',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Investment',
		[Sector.MUTUAL_FUND]: 'Mutual Fund',
		[Sector.OTHERS]: 'Others'
	};

	function formatNumber(value: number | undefined): string {
		if (value === undefined) return '-';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function formatSector(sector: number | undefined): string {
		if (sector === undefined) return '';
		return sectorDisplayNames[sector] ?? 'Unknown';
	}
</script>

{#if company && priceData}
	<article class="story">
		<!-- Act 1: The Introduction -->
		<header class="act act-intro">
			<h1 class="company-symbol">{company.symbol}</h1>
			<p class="company-name">{company.name}</p>
			<p class="company-meta">
				{formatSector(company.sector)} · {company.status === CompanyStatus.ACTIVE ? 'Active' : 'Inactive'}
			</p>
		</header>

		<!-- Act 2: The Current State -->
		<section class="act act-price">
			<p class="price-main">Rs. {formatNumber(priceData.ltp)}</p>
			{#if story}
				<p class="price-context">{story.price.positionSentence}</p>
				{#if story.price.volumeContext}
					<p class="price-volume">{story.price.trendSentence} {story.price.volumeContext}</p>
				{:else}
					<p class="price-volume">{story.price.trendSentence}</p>
				{/if}
			{/if}
			<div class="price-change" class:positive={priceData.changePercent && priceData.changePercent > 0} class:negative={priceData.changePercent && priceData.changePercent < 0}>
				{priceData.change && priceData.change > 0 ? '+' : ''}{formatNumber(priceData.change)} ({priceData.changePercent && priceData.changePercent > 0 ? '+' : ''}{formatNumber(priceData.changePercent)}%) today
			</div>
		</section>

		<!-- Act 3: The Journey (Price Chart placeholder) -->
		<section class="act act-journey">
			<h2>The Journey</h2>
			<div class="chart-placeholder">
				<p>📈 Price chart coming soon</p>
				<p class="chart-note">1Y price history with {priceHistory?.length ?? 0} data points</p>
			</div>
			{#if story}
				<p class="journey-narrative">
					{story.price.trendSentence}. Currently trading at Rs. {formatNumber(priceData.ltp)}.
				</p>
			{/if}
		</section>

		<!-- Act 4: The Fundamentals Story -->
		{#if fundamentals && story}
			<section class="act act-fundamentals">
				<div class="fundamentals-section">
					<h2>Earnings</h2>
					<p class="fundamentals-headline">{story.earnings.headline}</p>
					<p class="fundamentals-detail">{story.earnings.detail}</p>
					<p class="fundamentals-trend">{story.earnings.trendSentence}</p>
				</div>

				<div class="fundamentals-section">
					<h2>Valuation</h2>
					<p class="fundamentals-headline">{story.valuation.headline}</p>
					<p class="fundamentals-detail">{story.valuation.peContext}</p>
				</div>

				<!-- Key metrics as supporting data -->
				<div class="metrics-grid">
					<div class="metric">
						<span class="metric-label">EPS</span>
						<span class="metric-value">Rs. {formatNumber(fundamentals.eps)}</span>
					</div>
					<div class="metric">
						<span class="metric-label">P/E</span>
						<span class="metric-value">{formatNumber(fundamentals.peRatio)}</span>
					</div>
					<div class="metric">
						<span class="metric-label">Book Value</span>
						<span class="metric-value">Rs. {formatNumber(fundamentals.bookValue)}</span>
					</div>
				</div>
			</section>
		{/if}

		<!-- Act 5: The Verdict -->
		{#if story}
			<section class="act act-verdict" class:opportunity={story.verdict.signal === 'opportunity'} class:caution={story.verdict.signal === 'caution'}>
				<h2>{story.verdict.title}</h2>
				<p class="verdict-summary">{story.verdict.summary}</p>
			</section>
		{/if}
	</article>
{/if}

<style>
	.story {
		max-width: 720px;
		margin: 0 auto;
		padding: 3rem 1.5rem;
	}

	.act {
		margin-bottom: 4rem;
	}

	/* Act 1: Introduction */
	.act-intro {
		text-align: center;
		padding: 2rem 0 3rem;
		border-bottom: 1px solid var(--border);
	}

	.company-symbol {
		font-size: 3rem;
		font-weight: 700;
		margin: 0;
		letter-spacing: -0.02em;
	}

	.company-name {
		font-size: 1.25rem;
		color: var(--muted-foreground);
		margin: 0.5rem 0;
	}

	.company-meta {
		font-size: 0.875rem;
		color: var(--muted-foreground);
		margin: 0;
	}

	/* Act 2: Current State */
	.act-price {
		text-align: center;
		padding: 2rem 0;
	}

	.price-main {
		font-size: 3.5rem;
		font-weight: 700;
		margin: 0;
		letter-spacing: -0.02em;
	}

	.price-context {
		font-size: 1.125rem;
		font-style: italic;
		color: var(--muted-foreground);
		margin: 1rem 0 0.5rem;
	}

	.price-volume {
		font-size: 1rem;
		color: var(--muted-foreground);
		margin: 0.25rem 0;
	}

	.price-change {
		display: inline-block;
		margin-top: 1rem;
		padding: 0.5rem 1rem;
		border-radius: var(--radius);
		font-weight: 500;
		background: var(--muted);
	}

	.price-change.positive {
		color: #16a34a;
		background: #dcfce7;
	}

	.price-change.negative {
		color: #dc2626;
		background: #fee2e2;
	}

	/* Act 3: Journey */
	.act-journey h2 {
		font-size: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.chart-placeholder {
		background: var(--muted);
		border-radius: var(--radius);
		padding: 4rem 2rem;
		text-align: center;
		margin-bottom: 1.5rem;
	}

	.chart-placeholder p {
		margin: 0;
		color: var(--muted-foreground);
	}

	.chart-note {
		font-size: 0.875rem;
		margin-top: 0.5rem !important;
	}

	.journey-narrative {
		font-size: 1.125rem;
		line-height: 1.7;
		color: var(--foreground);
	}

	/* Act 4: Fundamentals */
	.act-fundamentals h2 {
		font-size: 1.25rem;
		color: var(--muted-foreground);
		margin-bottom: 0.75rem;
	}

	.fundamentals-section {
		margin-bottom: 2.5rem;
	}

	.fundamentals-headline {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 0.5rem;
	}

	.fundamentals-detail {
		font-size: 1.125rem;
		line-height: 1.6;
		margin: 0 0 0.5rem;
	}

	.fundamentals-trend {
		font-size: 1rem;
		color: var(--muted-foreground);
		margin: 0;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
		margin-top: 2rem;
		padding-top: 2rem;
		border-top: 1px solid var(--border);
	}

	.metric {
		text-align: center;
	}

	.metric-label {
		display: block;
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--muted-foreground);
		margin-bottom: 0.25rem;
	}

	.metric-value {
		font-size: 1.25rem;
		font-weight: 600;
	}

	/* Act 5: Verdict */
	.act-verdict {
		background: var(--muted);
		padding: 2rem;
		border-radius: var(--radius);
		border-left: 4px solid var(--border);
	}

	.act-verdict.opportunity {
		border-left-color: #16a34a;
		background: #f0fdf4;
	}

	.act-verdict.caution {
		border-left-color: #eab308;
		background: #fefce8;
	}

	.act-verdict h2 {
		font-size: 1.125rem;
		margin: 0 0 1rem;
		color: var(--muted-foreground);
	}

	.verdict-summary {
		font-size: 1.125rem;
		line-height: 1.7;
		margin: 0;
	}

	/* Dark mode adjustments */
	:global(.dark) .price-change.positive {
		background: #166534;
		color: #bbf7d0;
	}

	:global(.dark) .price-change.negative {
		background: #991b1b;
		color: #fecaca;
	}

	:global(.dark) .act-verdict.opportunity {
		background: #14532d;
		border-left-color: #22c55e;
	}

	:global(.dark) .act-verdict.caution {
		background: #422006;
		border-left-color: #facc15;
	}
</style>
