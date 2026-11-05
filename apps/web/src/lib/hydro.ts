import { PlantState } from '#lib/gen/ntx/v1/hydro_pb.js';

// A state is carried by a glyph and a label as well as a colour, in rows and in the legend.
export const STATE: Record<
	PlantState,
	{ key: string; label: string; glyph: string; color: string }
> = {
	[PlantState.UNSPECIFIED]: {
		key: '',
		label: 'no readings',
		glyph: '',
		color: 'var(--color-border)'
	},
	[PlantState.NORMAL]: {
		key: 'normal',
		label: 'normal',
		glyph: '·',
		color: 'var(--color-chart-1)'
	},
	[PlantState.HEAVY_RAIN]: {
		key: 'heavy_rain',
		label: 'heavy rain',
		glyph: '△',
		color: 'var(--color-caution)'
	},
	[PlantState.HIGH_DISCHARGE]: {
		key: 'high_discharge',
		label: 'high discharge',
		glyph: '▲',
		color: 'var(--color-negative)'
	}
};

// The three states a plant can be in, in order of severity; UNSPECIFIED means no readings.
export const STATES = [PlantState.NORMAL, PlantState.HEAVY_RAIN, PlantState.HIGH_DISCHARGE];

export function stateFromKey(key: string): PlantState | undefined {
	const hit = Object.entries(STATE).find(([, s]) => s.key === key && key !== '');
	return hit ? (Number(hit[0]) as PlantState) : undefined;
}

export function num(value: number | undefined, digits = 1): string {
	if (value === undefined) return '—';
	return value.toLocaleString('en-NP', {
		minimumFractionDigits: digits,
		maximumFractionDigits: digits
	});
}

export const STATUS_LABEL: Record<string, string> = {
	operating: 'operating',
	under_construction: 'under construction'
};
