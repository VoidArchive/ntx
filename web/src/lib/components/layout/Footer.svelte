<script lang="ts">
	import { GithubIcon, ExternalLinkIcon } from '@lucide/svelte';
</script>

<footer class="border-t bg-muted/30">
	<div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
		<div class="grid gap-8 md:grid-cols-4">
			<!-- About -->
			<div class="md:col-span-2">
				<h3 class="text-lg font-semibold">NTX</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Open-source NEPSE data aggregator with screening capabilities. Access stock data, company
					fundamentals, and market insights - all in one place.
				</p>
				<div class="mt-4 flex items-center gap-4">
					<a
						href="https://github.com/your-repo/ntx"
						target="_blank"
						rel="noopener noreferrer"
						class="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
					>
						<GithubIcon class="h-4 w-4" />
						Star on GitHub
					</a>
					<span
						class="inline-flex items-center gap-1 rounded-full border bg-background px-3 py-1 text-xs font-medium"
					>
						Open Source
					</span>
				</div>
			</div>

			<!-- Quick Links -->
			<div>
				<h4 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
					Quick Links
				</h4>
				<ul class="mt-4 space-y-2">
					<li>
						<a href="/companies" class="text-sm text-foreground/80 transition-colors hover:text-foreground"
							>All Companies</a
						>
					</li>
					<li>
						<a href="/sectors" class="text-sm text-foreground/80 transition-colors hover:text-foreground"
							>Sectors</a
						>
					</li>
					<li>
						<a href="/screener" class="text-sm text-foreground/80 transition-colors hover:text-foreground"
							>Stock Screener</a
						>
					</li>
				</ul>
			</div>

			<!-- Resources -->
			<div>
				<h4 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
					Resources
				</h4>
				<ul class="mt-4 space-y-2">
					<li>
						<a
							href="https://nepalstock.com"
							target="_blank"
							rel="noopener noreferrer"
							class="inline-flex items-center gap-1 text-sm text-foreground/80 transition-colors hover:text-foreground"
						>
							NEPSE Official
							<ExternalLinkIcon class="h-3 w-3" />
						</a>
					</li>
					<li>
						<a
							href="https://github.com/your-repo/ntx"
							target="_blank"
							rel="noopener noreferrer"
							class="inline-flex items-center gap-1 text-sm text-foreground/80 transition-colors hover:text-foreground"
						>
							Documentation
							<ExternalLinkIcon class="h-3 w-3" />
						</a>
					</li>
					<li>
						<a
							href="https://github.com/your-repo/ntx/issues"
							target="_blank"
							rel="noopener noreferrer"
							class="inline-flex items-center gap-1 text-sm text-foreground/80 transition-colors hover:text-foreground"
						>
							Report Issue
							<ExternalLinkIcon class="h-3 w-3" />
						</a>
					</li>
				</ul>
			</div>
		</div>

		<!-- Bottom -->
		<div class="mt-8 border-t pt-8">
			<p class="text-center text-xs text-muted-foreground">
				Data aggregated from NEPSE. This is not financial advice. NTX is open source and
				available under the MIT License.
			</p>
		</div>
	</div>
</footer>
