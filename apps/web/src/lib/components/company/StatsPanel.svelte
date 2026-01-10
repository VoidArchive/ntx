<script lang="ts">
	import type { Price, Fundamental } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		price?: Price;
		fundamentals?: Fundamental;
		priceHistory?: Price[];
	}

	let { price, fundamentals, priceHistory }: Props = $props();

	function fmt(value: number | bigint | undefined): string {
		if (value === undefined) return '—';
		return Number(value).toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function fmtLarge(value: number | bigint | undefined): string {
		if (value === undefined) return '—';
		const num = Number(value);
		if (num >= 1_000_000_000) return `${(num / 1_000_000_000).toFixed(2)}B`;
		if (num >= 10_000_000) return `${(num / 10_000_000).toFixed(2)} Cr`;
		if (num >= 100_000) return `${(num / 100_000).toFixed(2)} L`;
		return fmt(num);
	}

	// Calculate 52-week high/low from price history
	let rangeInfo = $derived.by(() => {
		if (!priceHistory || priceHistory.length === 0) return null;
		const highs = priceHistory.map((p) => p.high ?? p.ltp ?? 0).filter((v) => v > 0);
		const lows = priceHistory.map((p) => p.low ?? p.ltp ?? 0).filter((v) => v > 0);
		if (highs.length === 0 || lows.length === 0) return null;

		return {
			high52w: Math.max(...highs),
			low52w: Math.min(...lows)
		};
	});

	// Calculate outstanding shares from paid-up capital (face value = Rs. 100)
	let outstandingShares = $derived.by(() => {
		if (!fundamentals?.paidUpCapital) return null;
		return fundamentals.paidUpCapital / 100;
	});

	// Calculate market cap: price * outstanding shares
	let marketCap = $derived.by(() => {
		const currentPrice = price?.ltp ?? price?.close;
		if (!currentPrice || !outstandingShares) return null;
		return currentPrice * outstandingShares;
	});

	interface StatRow {
		label: string;
		value: string;
	}

	let stats = $derived.by((): StatRow[] => {
		const rows: StatRow[] = [];

		if (marketCap) {
			rows.push({ label: 'Market Cap', value: fmtLarge(marketCap) });
		}

		if (rangeInfo) {
			rows.push({ label: '52w High', value: `Rs. ${fmt(rangeInfo.high52w)}` });
			rows.push({ label: '52w Low', value: `Rs. ${fmt(rangeInfo.low52w)}` });
		}

		if (fundamentals?.peRatio) {
			rows.push({ label: 'P/E', value: fmt(fundamentals.peRatio) });
		}

		if (fundamentals?.eps) {
			rows.push({ label: 'EPS', value: fmt(fundamentals.eps) });
		}

		if (fundamentals?.bookValue) {
			rows.push({ label: 'Book Value', value: fmt(fundamentals.bookValue) });
		}

		if (price?.volume) {
			rows.push({ label: 'Volume', value: fmtLarge(price.volume) });
		}

		if (outstandingShares) {
			rows.push({ label: 'Listed Shares', value: fmtLarge(outstandingShares) });
		}

		return rows;
	});
</script>

<div class="text-sm">
	{#each stats as stat, i (stat.label)}
		<div
			class="flex justify-between py-2.5 {i < stats.length - 1
				? 'border-b border-dotted border-border'
				: ''}"
		>
			<span class="text-muted-foreground">{stat.label}</span>
			<span class="tabular-nums font-medium">{stat.value}</span>
		</div>
	{/each}
</div>
