<script lang="ts">
	let { data } = $props();
	let company = $derived(data.company);
	let fundamentals = $derived(data.fundamentals);
	let priceData = $derived(data.price);

	function formatNumber(value: number | undefined): string {
		if (value === undefined) return '-';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function formatCrore(value: number | undefined): string {
		if (value === undefined) return '-';
		const crore = value / 10_000_000;
		return crore.toLocaleString('en-NP', { maximumFractionDigits: 2 }) + ' Cr';
	}

	function formatVolume(value: number | bigint | undefined): string {
		if (value === undefined) return '-';
		const num = typeof value === 'bigint' ? Number(value) : value;
		if (num >= 1_000_000) {
			return (num / 1_000_000).toFixed(2) + 'M';
		}
		if (num >= 1_000) {
			return (num / 1_000).toFixed(2) + 'K';
		}
		return num.toLocaleString();
	}
</script>

{#if company}
	<div class="company-page">
		<header class="company-header">
			<div class="company-info">
				<h1>{company.name}</h1>
				<span class="symbol">{company.symbol}</span>
			</div>
			{#if priceData}
				<div class="price-display">
					<span class="ltp">Rs. {formatNumber(priceData.ltp)}</span>
					<span class="change" class:positive={priceData.changePercent && priceData.changePercent > 0} class:negative={priceData.changePercent && priceData.changePercent < 0}>
						{priceData.change && priceData.change > 0 ? '+' : ''}{formatNumber(priceData.change)} ({priceData.changePercent && priceData.changePercent > 0 ? '+' : ''}{formatNumber(priceData.changePercent)}%)
					</span>
				</div>
			{/if}
		</header>

		<div class="company-meta">
			<span class="sector">{company.sector}</span>
			<span class="instrument">{company.instrumentType}</span>
		</div>

		{#if priceData}
			<section class="price-details">
				<h2>Today's Trading</h2>
				<div class="price-grid">
					<div class="price-card">
						<span class="label">Open</span>
						<span class="value">{formatNumber(priceData.open)}</span>
					</div>
					<div class="price-card">
						<span class="label">High</span>
						<span class="value">{formatNumber(priceData.high)}</span>
					</div>
					<div class="price-card">
						<span class="label">Low</span>
						<span class="value">{formatNumber(priceData.low)}</span>
					</div>
					<div class="price-card">
						<span class="label">Close</span>
						<span class="value">{formatNumber(priceData.close)}</span>
					</div>
					<div class="price-card">
						<span class="label">Prev Close</span>
						<span class="value">{formatNumber(priceData.previousClose)}</span>
					</div>
					<div class="price-card">
						<span class="label">Volume</span>
						<span class="value">{formatVolume(priceData.volume)}</span>
					</div>
				</div>
				<p class="business-date">As of {priceData.businessDate}</p>
			</section>
		{/if}

		{#if fundamentals}
			<section class="fundamentals">
				<h2>Key Ratios</h2>
				<div class="ratios-grid">
					<div class="ratio-card">
						<span class="label">EPS</span>
						<span class="value">{formatNumber(fundamentals.eps)}</span>
					</div>
					<div class="ratio-card">
						<span class="label">P/E Ratio</span>
						<span class="value">{formatNumber(fundamentals.peRatio)}</span>
					</div>
					<div class="ratio-card">
						<span class="label">Book Value</span>
						<span class="value">{formatNumber(fundamentals.bookValue)}</span>
					</div>
					<div class="ratio-card">
						<span class="label">Paid-up Capital</span>
						<span class="value">{formatCrore(fundamentals.paidUpCapital)}</span>
					</div>
					<div class="ratio-card">
						<span class="label">Net Profit</span>
						<span class="value">{formatCrore(fundamentals.profitAmount)}</span>
					</div>
				</div>
				<p class="fiscal-year">As of FY {fundamentals.fiscalYear}{fundamentals.quarter ? ` (${fundamentals.quarter})` : ''}</p>
			</section>
		{:else}
			<section class="fundamentals">
				<h2>Key Ratios</h2>
				<p class="no-data">No fundamental data available</p>
			</section>
		{/if}
	</div>
{/if}

<style>
	.company-page {
		max-width: 800px;
		margin: 0 auto;
		padding: 2rem;
	}

	.company-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 0.5rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.company-info {
		display: flex;
		align-items: baseline;
		gap: 1rem;
	}

	.company-info h1 {
		margin: 0;
		font-size: 1.75rem;
	}

	.symbol {
		font-size: 1rem;
		color: #666;
		font-weight: 500;
	}

	.price-display {
		text-align: right;
	}

	.ltp {
		font-size: 1.75rem;
		font-weight: 700;
		display: block;
	}

	.change {
		font-size: 1rem;
		font-weight: 500;
	}

	.change.positive {
		color: #16a34a;
	}

	.change.negative {
		color: #dc2626;
	}

	.company-meta {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		color: #666;
		font-size: 0.875rem;
	}

	.price-details h2,
	.fundamentals h2 {
		font-size: 1.25rem;
		margin-bottom: 1rem;
	}

	.price-grid,
	.ratios-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 1rem;
	}

	.price-card,
	.ratio-card {
		background: #f8f9fa;
		border-radius: 8px;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.price-card .label,
	.ratio-card .label {
		font-size: 0.75rem;
		color: #666;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.price-card .value,
	.ratio-card .value {
		font-size: 1.25rem;
		font-weight: 600;
	}

	.business-date,
	.fiscal-year {
		margin-top: 1rem;
		font-size: 0.875rem;
		color: #666;
	}

	.price-details {
		margin-bottom: 2rem;
	}

	.no-data {
		color: #666;
		font-style: italic;
	}
</style>
