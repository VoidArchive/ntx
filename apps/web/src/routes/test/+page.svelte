<script lang="ts">
    let {data} = $props();
    let companies = data.companies;
</script>

{#each companies as company}
    <div>{company.name}</div>
{/each}