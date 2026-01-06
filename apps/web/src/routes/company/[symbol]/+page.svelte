<script lang="ts">
	let { data } = $props();
	let company = $derived(data.company);
</script>

{#if company}
	<div>{company.name}</div>
	<div>{company.symbol}</div>
	<div>{company.instrumentType}</div>
	<div>{company.sector}</div>
{/if}
