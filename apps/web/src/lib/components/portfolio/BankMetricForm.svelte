<script lang="ts">
	import { enhance } from '$app/forms';

	// One quarter of a lender's figures, typed from the filing with its date and URL
	// (ADR 0002). Shared by the company page and a needs-data candidate row.
	let {
		symbol,
		fiscalYear = '',
		quarter = 'Q1'
	}: { symbol: string; fiscalYear?: string; quarter?: string } = $props();

	const field = 'h-8 rounded-md border border-input bg-background px-2 text-sm';
	const bankColumns = [
		['nplPercent', 'NPL %'],
		['carPercent', 'CAR %'],
		['loanGrowthPercent', 'Loan growth %'],
		['distributableProfit', 'Distributable profit Rs']
	] as const;
</script>

<form
	method="POST"
	action="?/bankMetric"
	use:enhance
	class="mt-3 flex flex-wrap items-end gap-2 text-sm"
>
	<input type="hidden" name="symbol" value={symbol} />
	<input
		type="text"
		name="fiscal_year"
		placeholder="2083/84"
		value={fiscalYear}
		required
		aria-label="Fiscal year"
		class="{field} w-24"
	/>
	<select name="quarter" aria-label="Quarter" value={quarter} class={field}>
		{#each ['Q1', 'Q2', 'Q3', 'Q4'] as q (q)}<option>{q}</option>{/each}
	</select>
	{#each bankColumns as [key, label] (key)}
		<input
			type="number"
			step="any"
			name={key}
			placeholder={label}
			aria-label={label}
			class="{field} w-36"
		/>
	{/each}
	<input type="date" name="published_date" required aria-label="Published date" class={field} />
	<input
		type="url"
		name="source_url"
		placeholder="https://… filing"
		required
		aria-label="Source URL"
		class="{field} min-w-64 flex-1"
	/>
	<button
		type="submit"
		class="h-8 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent"
		>Save quarter</button
	>
</form>
