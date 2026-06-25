<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getEventResults } from '$lib/api-client.js';
	import { getToken } from '$lib/auth.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import type { CycleInfo, FormInfo, CycleAverageResult, FreeTextResult } from '$lib/api/types.gen.js';

	const id = Number($page.params.id);

	const resultsQuery = createQuery({
		queryKey: ['results', id],
		queryFn: () => getEventResults({ path: { id } }).then((r) => r.data)
	});

	const results = $derived($resultsQuery.data);

	const sortedCycles = $derived(
		[...(results?.cycles ?? [])].sort(
			(a: CycleInfo, b: CycleInfo) => a.orderIndex - b.orderIndex
		)
	);

	const ratingForms = $derived(
		(results?.forms ?? []).filter(
			(f: FormInfo) => f.type === 'rating' || f.type === 'mood'
		).sort((a: FormInfo, b: FormInfo) => a.orderIndex - b.orderIndex)
	);

	const freeTextForms = $derived(
		(results?.forms ?? []).filter((f: FormInfo) => f.type === 'free_text')
	);

	function getAvg(cycleId: number, formId: number): string {
		const entry = (results?.avgTable ?? []).find(
			(r: CycleAverageResult) => r.cycleId === cycleId && r.formId === formId
		);
		if (!entry) return '—';
		return entry.average.toFixed(2);
	}

	function getFreeTexts(cycleId: number, formId: number): string[] {
		const entry = (results?.freeTexts ?? []).find(
			(r: FreeTextResult) => r.cycleId === cycleId && r.formId === formId
		);
		return entry?.texts ?? [];
	}

	function downloadCSV() {
		const token = getToken();
		fetch(`/api/events/${id}/results/export`, {
			headers: token ? { Authorization: `Bearer ${token}` } : {}
		})
			.then((r) => r.blob())
			.then((blob) => {
				const url = URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `event-${id}-results.csv`;
				a.click();
				URL.revokeObjectURL(url);
			})
			.catch(() => {});
	}
</script>

<div class="flex flex-col gap-6 py-6 px-4 lg:px-6">
	<!-- Back + header -->
	<div class="flex items-center justify-between gap-4">
		<div>
			<button
				onclick={() => goto(`/app/events/${id}/control`)}
				class="mb-2 block text-[13px] hover:underline"
				style="color: var(--t3)"
			>← Back to control panel</button>
			<h1 class="text-2xl font-semibold tracking-tight">Results</h1>
		</div>
		<button
			onclick={downloadCSV}
			class="rounded-[9px] border px-4 py-2 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
			style="background: none; border-color: var(--border-h2); color: var(--t1)"
		>
			Download CSV
		</button>
	</div>

	{#if $resultsQuery.isLoading}
		<Skeleton class="h-40 w-full" />
	{:else if !results}
		<div class="text-[14px]" style="color: var(--danger)">Results not available.</div>
	{:else}
		<!-- Averages table (rating + mood forms only) -->
		{#if ratingForms.length > 0 && sortedCycles.length > 0}
			<div
				class="rounded-[var(--radius-card)] border overflow-hidden"
				style="border-color: var(--border-h)"
			>
				<div class="px-5 py-4 border-b" style="background: var(--s2); border-color: var(--border-h)">
					<h2 class="text-[15px] font-semibold">Averages</h2>
				</div>
				<div class="overflow-x-auto">
					<table class="w-full border-collapse" style="background: var(--s2)">
						<thead>
							<tr>
								<th
									class="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.05em]"
									style="background: var(--s1); color: var(--t3); border-bottom: 1px solid var(--border-h)"
								>Cycle</th>
								{#each ratingForms as form}
									<th
										class="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.05em]"
										style="background: var(--s1); color: var(--t3); border-bottom: 1px solid var(--border-h)"
									>{form.label}</th>
								{/each}
							</tr>
						</thead>
						<tbody>
							{#each sortedCycles as cycle, i}
								<tr>
									<td
										class="px-5 py-3.5 text-[14px] font-medium"
										style="border-bottom: {i < sortedCycles.length - 1 ? '1px solid var(--border-h)' : 'none'}; color: var(--t1)"
									>{cycle.name}</td>
									{#each ratingForms as form}
										<td
											class="px-5 py-3.5 text-[14px] mono"
											style="border-bottom: {i < sortedCycles.length - 1 ? '1px solid var(--border-h)' : 'none'}; color: var(--t2)"
										>{getAvg(cycle.id, form.id)}</td>
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}

		<!-- Free text responses -->
		{#if freeTextForms.length > 0}
			<div class="flex flex-col gap-4">
				<h2 class="text-[15px] font-semibold">Free Text Responses</h2>
				{#each sortedCycles as cycle}
					{#each freeTextForms as form}
						{@const texts = getFreeTexts(cycle.id, form.id)}
						{#if texts.length > 0}
							<div
								class="rounded-[var(--radius-card)] border p-5"
								style="background: var(--s2); border-color: var(--border-h)"
							>
								<div class="flex items-center gap-2 mb-3">
									<span class="text-[13px] font-semibold" style="color: var(--t1)">{cycle.name}</span>
									<span class="text-[12px]" style="color: var(--t3)">—</span>
									<span class="text-[13px]" style="color: var(--t2)">{form.label}</span>
								</div>
								<ul class="flex flex-col gap-1.5">
									{#each texts as text}
										<li class="flex items-start gap-2 text-[13.5px]" style="color: var(--t1)">
											<span class="mt-1.5 size-1.5 rounded-full shrink-0" style="background: var(--t3)"></span>
											{text}
										</li>
									{/each}
								</ul>
							</div>
						{/if}
					{/each}
				{/each}
			</div>
		{/if}

		{#if ratingForms.length === 0 && freeTextForms.length === 0}
			<div
				class="rounded-[var(--radius-card)] border py-12 text-center text-[13px]"
				style="background: var(--s2); border-color: var(--border-h); color: var(--t3)"
			>
				No results to display yet.
			</div>
		{/if}
	{/if}
</div>
