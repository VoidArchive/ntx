<script lang="ts">
	import type { PlantState } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { STATE } from '#lib/hydro.js';

	// Glyph, label and colour together, so colour is never the only carrier of a state.
	let { state, stale = false }: { state: PlantState; stale?: boolean } = $props();
	let s = $derived(STATE[state]);
</script>

<span aria-hidden="true" style="color: {s.color}">{s.glyph}</span>
{s.label}{#if stale}
	<span class="text-xs text-muted-foreground">(stale)</span>{/if}
