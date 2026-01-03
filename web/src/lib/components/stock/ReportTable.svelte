<script lang="ts">
	import { formatNumber, formatDate } from '$lib/utils/format';
	import * as Table from '$lib/components/ui/table';
	import type { Report } from '@ntx/api/ntx/v1/common_pb';
	import { ReportType } from '@ntx/api/ntx/v1/common_pb';

	let { reports }: { reports: Report[] } = $props();

	function formatReportType(type: ReportType): string {
		if (type === ReportType.QUARTERLY) return 'Q';
		if (type === ReportType.ANNUAL) return 'Annual';
		return '-';
	}

	function formatPeriod(report: Report): string {
		if (report.type === ReportType.QUARTERLY) {
			return `Q${report.quarter} ${report.fiscalYear}`;
		}
		return `FY ${report.fiscalYear}`;
	}
</script>

{#if reports.length > 0}
	<div class="rounded-lg border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Period</Table.Head>
					<Table.Head class="text-right">Revenue</Table.Head>
					<Table.Head class="text-right">Net Income</Table.Head>
					<Table.Head class="text-right">EPS</Table.Head>
					<Table.Head class="text-right">Book Value</Table.Head>
					<Table.Head class="hidden text-right md:table-cell">Published</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each reports as report}
					<Table.Row>
						<Table.Cell class="font-medium">{formatPeriod(report)}</Table.Cell>
						<Table.Cell class="text-right font-mono tabular-nums">
							{report.revenue ? `Rs. ${formatNumber(report.revenue / 1_000_000, 2)}M` : '-'}
						</Table.Cell>
						<Table.Cell class="text-right font-mono tabular-nums">
							{report.netIncome ? `Rs. ${formatNumber(report.netIncome / 1_000_000, 2)}M` : '-'}
						</Table.Cell>
						<Table.Cell class="text-right font-mono tabular-nums">
							{report.eps ? `Rs. ${formatNumber(report.eps, 2)}` : '-'}
						</Table.Cell>
						<Table.Cell class="text-right font-mono tabular-nums">
							{report.bookValue ? `Rs. ${formatNumber(report.bookValue, 2)}` : '-'}
						</Table.Cell>
						<Table.Cell class="hidden text-right text-muted-foreground md:table-cell">
							{report.publishedAt ? formatDate(new Date(Number(report.publishedAt.seconds) * 1000)) : '-'}
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
{:else}
	<div class="rounded-lg border bg-muted/30 py-12 text-center">
		<p class="text-muted-foreground">No financial reports available</p>
	</div>
{/if}
