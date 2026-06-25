<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getEventResults } from '$lib/api-client.js';
	import { getToken } from '$lib/auth.js';
	import type { CycleInfo, FormInfo, CycleAverageResult, FreeTextResult } from '$lib/api/types.gen.js';

	const id = Number($page.params.id);

	const resultsQuery = createQuery({
		queryKey: ['results', id],
		queryFn: () => getEventResults({ path: { id } }).then((r) => r.data)
	});

	const results = $derived($resultsQuery.data);

	const sortedCycles = $derived(
		[...(results?.cycles ?? [])].sort((a: CycleInfo, b: CycleInfo) => a.orderIndex - b.orderIndex)
	);

	const ratingForms = $derived(
		(results?.forms ?? [])
			.filter((f: FormInfo) => f.type === 'rating' || f.type === 'mood')
			.sort((a: FormInfo, b: FormInfo) => a.orderIndex - b.orderIndex)
	);

	const freeTextForms = $derived(
		(results?.forms ?? []).filter((f: FormInfo) => f.type === 'free_text')
	);

	function getAvg(cycleId: number, formId: number): string {
		const entry = (results?.avgTable ?? []).find(
			(r: CycleAverageResult) => r.cycleId === cycleId && r.formId === formId
		);
		return entry ? entry.average.toFixed(2) : '—';
	}

	function getOverallAvg(cycleId: number): number {
		const avgs = ratingForms
			.map((f: FormInfo) => {
				const entry = (results?.avgTable ?? []).find(
					(r: CycleAverageResult) => r.cycleId === cycleId && r.formId === f.id
				);
				return entry ? entry.average : null;
			})
			.filter((v: number | null): v is number => v !== null);
		if (avgs.length === 0) return 0;
		return avgs.reduce((a: number, b: number) => a + b, 0) / avgs.length;
	}

	/* Sort cycles by overall avg descending for ranking */
	const rankedCycles = $derived(
		[...sortedCycles].sort((a: CycleInfo, b: CycleInfo) => getOverallAvg(b.id) - getOverallAvg(a.id))
	);

	const maxAvg = $derived(
		rankedCycles.length > 0 ? Math.max(...rankedCycles.map((c: CycleInfo) => getOverallAvg(c.id))) : 1
	);

	function scoreBarWidth(cycleId: number): string {
		const max = maxAvg || 1;
		return `${Math.round((getOverallAvg(cycleId) / max) * 100)}%`;
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

<div class="canvas">
	<div class="wrap">
		<button class="back" onclick={() => goto(`/app/events/${id}/control`)}>← Back to control</button>

		<div style="display:flex;align-items:center;justify-content:space-between;gap:16px;flex-wrap:wrap;margin-bottom:8px">
			<div>
				<div class="eyebrow">Host only · final tally</div>
				<h1 class="title">Results</h1>
			</div>
			<button class="btn ghost" onclick={downloadCSV}>
				<svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
				Export CSV
			</button>
		</div>
		<p class="sub" style="margin-bottom:24px">Averages from raters who responded. Cycles ranked by overall score.</p>

		{#if $resultsQuery.isLoading}
			<div style="color:var(--t3);font-size:14px;padding:24px 0">Loading results…</div>
		{:else if !results}
			<div style="color:var(--danger);font-size:14px">Results not available.</div>
		{:else}
			<!-- Ranked averages table -->
			{#if ratingForms.length > 0 && rankedCycles.length > 0}
				<div class="card" style="padding:6px 6px 8px;margin-bottom:18px">
					<table class="res">
						<thead>
							<tr>
								<th style="width:46px">#</th>
								<th>Cycle</th>
								{#each ratingForms as form}
									<th class="c">{form.label}</th>
								{/each}
								<th>Score</th>
							</tr>
						</thead>
						<tbody>
							{#each rankedCycles as cycle, i}
								<tr class:top={i === 0}>
									<td><span class="rank-badge">{i + 1}</span></td>
									<td style="font-weight:{i===0?'700':'400'}">{cycle.name}</td>
									{#each ratingForms as form}
										<td class="num-c c">{getAvg(cycle.id, form.id)}</td>
									{/each}
									<td>
										<div class="score-bar">
											<i style="width:{scoreBarWidth(cycle.id)}"></i>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}

			<!-- Free text responses -->
			{#if freeTextForms.length > 0}
				{#each freeTextForms as form}
					<div class="card" style="margin-bottom:18px">
						<div class="card-h" style="margin-bottom:16px">
							{form.label}
							<span style="font-size:12px;color:var(--t3);font-weight:400;margin-left:8px">(anonymous)</span>
						</div>
						{#each sortedCycles as cycle}
							{@const texts = getFreeTexts(cycle.id, form.id)}
							{#if texts.length > 0}
								<div style="margin-bottom:16px">
									<div style="font-size:12px;text-transform:uppercase;letter-spacing:.08em;color:var(--t3);margin-bottom:10px">{cycle.name}</div>
									{#each texts as text}
										<div class="quote">{text}</div>
									{/each}
								</div>
							{/if}
						{/each}
					</div>
				{/each}
			{/if}

			{#if ratingForms.length === 0 && freeTextForms.length === 0}
				<div class="card" style="padding:48px;text-align:center;color:var(--t3)">
					No results to display yet.
				</div>
			{/if}
		{/if}
	</div>
</div>
