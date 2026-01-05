<script lang="ts">
import { InstrumentType } from "$lib/gen/ntx/v1/common_pb.js";
    let {data} = $props();
</script>

{#each data.companies as company (company.id)}
  {#if company.instrumentType === InstrumentType.EQUITY}
    <div>{company.name}</div>
    <div>{company.symbol}</div>
    <div>{company.instrumentType}</div>
    <div>{company.sector}</div>
{/if}
  
{/each}
