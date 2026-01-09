<script lang="ts">
	import type { Company, Price, Fundamental } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		company: Company;
		price?: Price;
		fundamentals?: Fundamental;
		miniStory?: string;
	}

	let { company, price, miniStory = 'View details →' }: Props = $props();

	function formatNumber(value: number | undefined): string {
		if (value === undefined) return '-';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}
</script>

<a href="/company/{company.symbol}" class="card">
	<div class="card-header">
		<span class="symbol">{company.symbol}</span>
		<span class="name">{company.name}</span>
	</div>

	{#if price}
		<div class="card-price">
			<span class="ltp">Rs. {formatNumber(price.ltp)}</span>
			<span
				class="change"
				class:positive={price.changePercent && price.changePercent > 0}
				class:negative={price.changePercent && price.changePercent < 0}
			>
				{price.changePercent && price.changePercent > 0 ? '+' : ''}{formatNumber(price.changePercent)}%
			</span>
		</div>
	{/if}

	<p class="mini-story">{miniStory}</p>
</a>

<style>
	.card {
		display: block;
		padding: 1.25rem;
		background: var(--card);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		text-decoration: none;
		color: inherit;
		transition: border-color 0.15s ease, box-shadow 0.15s ease;
	}

	.card:hover {
		border-color: var(--primary);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}

	.card-header {
		margin-bottom: 0.75rem;
	}

	.symbol {
		display: block;
		font-size: 1.125rem;
		font-weight: 600;
		margin-bottom: 0.125rem;
	}

	.name {
		display: block;
		font-size: 0.875rem;
		color: var(--muted-foreground);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.card-price {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		margin-bottom: 0.75rem;
	}

	.ltp {
		font-size: 1.25rem;
		font-weight: 600;
	}

	.change {
		font-size: 0.875rem;
		font-weight: 500;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		background: var(--muted);
	}

	.change.positive {
		color: #16a34a;
		background: #dcfce7;
	}

	.change.negative {
		color: #dc2626;
		background: #fee2e2;
	}

	.mini-story {
		font-size: 0.875rem;
		color: var(--muted-foreground);
		font-style: italic;
		margin: 0;
		line-height: 1.4;
	}

	/* Dark mode */
	:global(.dark) .change.positive {
		background: #166534;
		color: #bbf7d0;
	}

	:global(.dark) .change.negative {
		background: #991b1b;
		color: #fecaca;
	}
</style>
