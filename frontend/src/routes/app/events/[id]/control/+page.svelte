<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { pageTitle } from '$lib/page-title.js';
	import {
		getEvent,
		activateEvent,
		startCycle,
		showForm,
		nextCycle,
		endEvent,
		getEventMonitor
	} from '$lib/api-client.js';
	import type { MonitorResponseReadable, CycleInfo } from '$lib/api/types.gen.js';

	const qc = useQueryClient();
	const id = Number($page.params.id);

	const eventQuery = createQuery({
		queryKey: ['event', id],
		queryFn: () => getEvent({ path: { id } }).then((r) => r.data)
	});

	let monitorData = $state<MonitorResponseReadable | null>(null);
	let monitorInterval: ReturnType<typeof setInterval> | undefined;

	onMount(() => {
		monitorInterval = setInterval(async () => {
			const r = await getEventMonitor({ path: { id } });
			if (r.data) monitorData = r.data;
		}, 3000);
	});
	onDestroy(() => { if (monitorInterval !== undefined) clearInterval(monitorInterval) });

	let selectedCycleId = $state<number | null>(null);
	// Track which cycles were explicitly locked by "Next cycle" — not by orderIndex comparison.
	let completedCycleIds = $state<Set<number>>(new Set());

	const activateMut  = createMutation({ mutationFn: () => activateEvent({ path: { id } }), onSuccess: () => { toast.success('Event activated'); qc.invalidateQueries({ queryKey: ['event', id] }) }, onError: () => toast.error('Failed to activate') });
	const startCycleMut= createMutation({ mutationFn: () => startCycle({ path: { id }, body: { cycleId: selectedCycleId! } }), onSuccess: () => { toast.success('Cycle started'); qc.invalidateQueries({ queryKey: ['event', id] }) }, onError: () => toast.error('Failed to start cycle') });
	const showFormMut  = createMutation({ mutationFn: () => showForm({ path: { id } }), onSuccess: () => { toast.success('Form opened'); qc.invalidateQueries({ queryKey: ['event', id] }) }, onError: () => toast.error('Failed to open form') });
	const nextCycleMut = createMutation({
		mutationFn: () => nextCycle({ path: { id }, body: { cycleId: selectedCycleId! } }),
		onSuccess: () => {
			// Mark the cycle that just ended as completed before the query refreshes.
			if (monitorData?.activeCycleId != null) {
				completedCycleIds = new Set([...completedCycleIds, monitorData.activeCycleId]);
			}
			selectedCycleId = null;
			toast.success('Next cycle');
			qc.invalidateQueries({ queryKey: ['event', id] });
		},
		onError: () => toast.error('Failed')
	});
	const endEventMut  = createMutation({ mutationFn: () => endEvent({ path: { id } }), onSuccess: () => { toast.success('Event ended'); qc.invalidateQueries({ queryKey: ['event', id] }) }, onError: () => toast.error('Failed to end') });

	const event  = $derived($eventQuery.data);

	$effect(() => {
		if (event?.name) pageTitle.set(event.name);
		return () => pageTitle.set(null);
	});
	const cycles = $derived(
		[...(event?.cycles ?? [])].sort((a: CycleInfo, b: CycleInfo) => a.orderIndex - b.orderIndex)
	);

	const activeCycle = $derived(
		cycles.find((c: CycleInfo) => c.id === monitorData?.activeCycleId) ?? null
	);
	const activeCycleOrderIndex = $derived(activeCycle?.orderIndex ?? -1);
	// Any cycle that hasn't been locked and isn't currently active can still be visited.
	const hasMoreCycles = $derived(
		cycles.some((c: CycleInfo) => !completedCycleIds.has(c.id) && c.id !== activeCycle?.id)
	);

	/* SVG ring math — r=72, circumference≈452 */
	const fraction = $derived(
		monitorData && (monitorData.participantCount ?? 0) > 0
			? (monitorData.respondedCount ?? 0) / (monitorData.participantCount ?? 1)
			: 0
	);
	const ringOffset = $derived(Math.round(452 * (1 - fraction)));

	function cycleStatus(c: CycleInfo): 'done' | 'now' | 'next' {
		// A cycle is "done" only if explicitly moved past via "Next cycle", not just because
		// another cycle with a higher orderIndex is currently active.
		if (completedCycleIds.has(c.id)) return 'done';
		if (activeCycle && c.id === activeCycle.id) return 'now';
		return 'next';
	}
	function cycleStatusLabel(c: CycleInfo): string {
		const s = cycleStatus(c);
		if (s === 'done') return 'Locked';
		if (s === 'now' && event?.currentStage === 'form_open') return '● Scoring now';
		if (s === 'now') return '● Waiting';
		return 'Queued';
	}

	// Show every cycle that isn't locked and isn't currently active — order doesn't restrict choice.
	const nextCycles = $derived(
		cycles.filter((c: CycleInfo) => !completedCycleIds.has(c.id) && c.id !== activeCycle?.id)
	);
</script>

<div class="canvas">
	<div class="wrap">
		<button class="back" onclick={() => goto('/app/dashboard')}>← Back to events</button>

		{#if $eventQuery.isLoading}
			<div style="color:var(--t3);font-size:14px">Loading…</div>
		{:else if !event}
			<div style="color:var(--danger);font-size:14px">Event not found.</div>
		{:else}
			<!-- Header -->
			<div style="display:flex;align-items:center;justify-content:space-between;gap:16px;flex-wrap:wrap;margin-bottom:26px">
				<div>
					<div class="eyebrow">Host control room</div>
					<h1 class="title">{event.name}</h1>
				</div>
				{#if event.status === 'active'}
					<span class="pill live"><span class="dot"></span>Live</span>
				{:else if event.status === 'ended'}
					<span class="pill draft" style="color:var(--t3)"><span class="dot"></span>Ended</span>
				{:else}
					<span class="pill draft"><span class="dot"></span>Draft</span>
				{/if}
			</div>

			<!-- Hero card: ring + legend + actions -->
			{#if event.status === 'active'}
				<div class="card" style="margin-bottom:18px">
					<div class="control-hero">
						<!-- Response ring -->
						<div class="ring-wrap">
							<svg width="168" height="168" viewBox="0 0 168 168">
								<circle cx="84" cy="84" r="72" fill="none" stroke="var(--s3)" stroke-width="14"/>
								<circle cx="84" cy="84" r="72" fill="none" stroke="var(--live)" stroke-width="14"
									stroke-linecap="round"
									stroke-dasharray="452"
									stroke-dashoffset={ringOffset}
									transform="rotate(-90 84 84)"/>
							</svg>
							<div class="center">
								<div class="big">
									{monitorData?.respondedCount ?? 0}<span style="color:var(--t3);font-size:24px">/{monitorData?.participantCount ?? 0}</span>
								</div>
								<div class="small">responded</div>
							</div>
						</div>

						<!-- Legend + action buttons -->
						<div>
							{#if activeCycle}
								<div style="font-size:12.5px;color:var(--t3);text-transform:uppercase;letter-spacing:.1em;margin-bottom:16px">
									Now scoring · {activeCycle.name}
								</div>
							{/if}
							<div class="legend" style="margin-bottom:20px">
								<div class="li">
									<span class="swatch" style="background:var(--live)"></span>
									<span class="lv">{monitorData?.respondedCount ?? 0}</span>
									<span class="ll">submitted &amp; locked</span>
								</div>
								<div class="li">
									<span class="swatch" style="background:var(--s3)"></span>
									<span class="lv" style="color:var(--t2)">{(monitorData?.participantCount ?? 0) - (monitorData?.respondedCount ?? 0)}</span>
									<span class="ll">still scoring</span>
								</div>
							</div>

							<!-- Context-sensitive action buttons -->
							<div style="display:flex;gap:10px;flex-wrap:wrap">
								{#if event.currentStage === 'idle'}
									<div style="display:flex;flex-direction:column;gap:10px;align-items:flex-start">
										<p style="font-size:13px;color:var(--t3)">Choose a cycle to start.</p>
										<select
											class="ipt"
											style="min-width:180px;max-width:260px"
											bind:value={selectedCycleId}
										>
											<option value={null} disabled selected>Select a cycle…</option>
											{#each cycles as c}
												<option value={c.id}>{c.name}</option>
											{/each}
										</select>
										<button
											class="btn primary"
											disabled={selectedCycleId === null || $startCycleMut.isPending}
											onclick={() => $startCycleMut.mutate()}
										>{$startCycleMut.isPending ? 'Starting…' : 'Start cycle'}</button>
									</div>
								{:else if event.currentStage === 'waiting'}
									<button
										class="btn ghost"
										disabled={$showFormMut.isPending}
										onclick={() => $showFormMut.mutate()}
									>{$showFormMut.isPending ? 'Opening…' : 'Show form'}</button>
								{:else if event.currentStage === 'form_open'}
									{#if hasMoreCycles}
										<div style="display:flex;flex-direction:column;gap:10px">
											<select
												class="ipt"
												style="min-width:180px;max-width:260px"
												bind:value={selectedCycleId}
											>
												<option value={null} disabled selected>Next cycle…</option>
												{#each nextCycles as c}
													<option value={c.id}>{c.name}</option>
												{/each}
											</select>
											<button
												class="btn btn-live"
												disabled={selectedCycleId === null || $nextCycleMut.isPending}
												onclick={() => $nextCycleMut.mutate()}
											>{$nextCycleMut.isPending ? 'Advancing…' : 'Next cycle →'}</button>
										</div>
									{:else}
										<button
											class="btn ghost"
											disabled={$endEventMut.isPending}
											onclick={() => $endEventMut.mutate()}
											style="border-color:var(--danger);color:var(--danger)"
										>{$endEventMut.isPending ? 'Ending…' : 'End event'}</button>
									{/if}
								{/if}
							</div>
						</div>
					</div>
				</div>

			{:else if event.status === 'draft'}
				<!-- Draft: activate card -->
				<div class="card" style="margin-bottom:18px">
					<div class="card-h" style="margin-bottom:12px">Activate event</div>
					<p style="font-size:13px;color:var(--t3);margin-bottom:18px">Open the event so raters can join and you can begin scoring.</p>
					<button
						class="btn primary"
						disabled={$activateMut.isPending}
						onclick={() => $activateMut.mutate()}
					>{$activateMut.isPending ? 'Activating…' : 'Activate event'}</button>
				</div>

			{:else if event.status === 'ended'}
				<!-- Ended: results card -->
				<div class="card" style="margin-bottom:18px">
					<div class="card-h" style="margin-bottom:12px">Event complete</div>
					<p style="font-size:13px;color:var(--t3);margin-bottom:18px">All cycles locked. View the final tally.</p>
					<button
						class="btn primary"
						onclick={() => goto(`/app/events/${id}/results`)}
					>View results →</button>
				</div>
			{/if}

			<!-- Cycle rail -->
			{#if cycles.length > 0}
				<div style="font-size:12.5px;color:var(--t3);text-transform:uppercase;letter-spacing:.1em;margin:0 2px 10px">Run of show</div>
				<div class="rail" style="margin-bottom:22px">
					{#each cycles as c, i}
						{@const status = cycleStatus(c)}
						<div class="rail-step {status}">
							<div class="si">
								<div class="badge">
									{#if status === 'done'}✓{:else}{i + 1}{/if}
								</div>
								<div class="nm">{c.name}</div>
							</div>
							<div class="st">{cycleStatusLabel(c)}</div>
						</div>
					{/each}
				</div>

				<!-- Rater count card -->
				{#if event.status === 'active' && monitorData}
					<div class="card">
						<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
							<div class="card-h">Participation</div>
							<span style="font-size:12px;color:var(--t3)">status only · scores stay hidden</span>
						</div>
						<div style="display:flex;gap:32px">
							<div>
								<div style="font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--t3);margin-bottom:4px">Joined</div>
								<div style="font-family:var(--display);font-size:28px;font-weight:700;color:var(--t1)">{monitorData.participantCount ?? 0}</div>
							</div>
							<div>
								<div style="font-size:11px;text-transform:uppercase;letter-spacing:.08em;color:var(--t3);margin-bottom:4px">Responded</div>
								<div style="font-family:var(--display);font-size:28px;font-weight:700;color:var(--live)">{monitorData.respondedCount ?? 0}</div>
							</div>
						</div>
					</div>
				{/if}
			{/if}
		{/if}
	</div>
</div>
