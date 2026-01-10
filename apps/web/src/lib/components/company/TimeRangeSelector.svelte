<script lang="ts">
	interface TimeRange {
		days: number;
		label: string;
	}

	interface Props {
		ranges?: TimeRange[];
		selected: number;
		onSelect: (days: number) => void;
		totalDays?: number;
	}

	const defaultRanges: TimeRange[] = [
		{ days: 1, label: '1D' },
		{ days: 7, label: '1W' },
		{ days: 30, label: '1M' },
		{ days: 90, label: '3M' },
		{ days: 180, label: '6M' },
		{ days: 365, label: '1YR' },
		{ days: 0, label: 'All' }
	];

	let { ranges = defaultRanges, selected, onSelect, totalDays = 365 }: Props = $props();

	function getDays(range: TimeRange): number {
		return range.days === 0 ? totalDays : range.days;
	}

	function isSelected(range: TimeRange): boolean {
		return selected === getDays(range);
	}
</script>

<div class="flex gap-1.5 px-6">
	{#each ranges as range (range.label)}
		<button
			onclick={() => onSelect(getDays(range))}
			class="rounded-md border px-3 py-1.5 text-sm font-medium transition-colors
				{isSelected(range)
					? 'border-foreground bg-foreground text-background'
					: 'border-border bg-background text-muted-foreground hover:border-foreground/50 hover:text-foreground'}"
		>
			{range.label}
		</button>
	{/each}
</div>
