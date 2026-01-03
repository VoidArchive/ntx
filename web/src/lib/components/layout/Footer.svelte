<script lang="ts">
	import { GithubIcon, ExternalLinkIcon } from '@lucide/svelte';
</script>

<footer class="border-t bg-muted/30">
	<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
		<div class="flex flex-col items-center justify-between gap-4 md:flex-row">
			<!-- Brand -->
			<div class="flex items-center gap-4">
				<span class="text-lg font-bold">NTX</span>
				<span class="text-sm text-muted-foreground">NEPSE Stock Screener</span>
			</div>

			<!-- Links -->
			<div class="flex items-center gap-6 text-sm">
				<a href="/screener" class="text-muted-foreground hover:text-foreground">Screener</a>
				<a href="/companies" class="text-muted-foreground hover:text-foreground">Companies</a>
				<a
					href="https://nepalstock.com"
					target="_blank"
					rel="noopener noreferrer"
					class="inline-flex items-center gap-1 text-muted-foreground hover:text-foreground"
				>
					NEPSE
					<ExternalLinkIcon class="h-3 w-3" />
				</a>
				<a
					href="https://github.com/voidarchive/ntx"
					target="_blank"
					rel="noopener noreferrer"
					class="inline-flex items-center gap-1 text-muted-foreground hover:text-foreground"
				>
					<GithubIcon class="h-4 w-4" />
				</a>
			</div>
		</div>

		<div class="mt-6 text-center text-xs text-muted-foreground">
			Data aggregated from NEPSE. Not financial advice.
		</div>
	</div>
</footer>
