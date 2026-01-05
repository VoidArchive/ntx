<script lang="ts">
    let {data} = $props();
</script>

{#each data.companies as company (company.id)}
    <a href="/company/{company.symbol}">
    <div>{company.name}</div>
  </a>
  
{/each}
