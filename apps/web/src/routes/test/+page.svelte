<script lang="ts">
    let {data} = $props();
    let companies = data.companies;
</script>

{#each companies as company (company.id)}
    <div>{company.name}</div>
    <div>{company.symbol}</div>
{/each}
