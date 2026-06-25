<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { toast } from 'svelte-sonner';
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
	onDestroy(() => {
		if (monitorInterval !== undefined) clearInterval(monitorInterval);
	});

	let selectedCycleId = $state<number | null>(null);

	const activateMut = createMutation({
		mutationFn: () => activateEvent({ path: { id } }),
		onSuccess: () => {
			toast.success('Event activated');
			qc.invalidateQueries({ queryKey: ['event', id] });
		},
		onError: () => toast.error('Failed to activate event')
	});

	const startCycleMut = createMutation({
		mutationFn: () => startCycle({ path: { id }, body: { cycleId: selectedCycleId! } }),
		onSuccess: () => {
			toast.success('Cycle started');
			qc.invalidateQueries({ queryKey: ['event', id] });
		},
		onError: () => toast.error('Failed to start cycle')
	});

	const showFormMut = createMutation({
		mutationFn: () => showForm({ path: { id } }),
		onSuccess: () => {
			toast.success('Form shown to participants');
			qc.invalidateQueries({ queryKey: ['event', id] });
		},
		onError: () => toast.error('Failed to show form')
	});

	const nextCycleMut = createMutation({
		mutationFn: () => nextCycle({ path: { id }, body: { cycleId: selectedCycleId! } }),
		onSuccess: () => {
			toast.success('Moving to next cycle');
			qc.invalidateQueries({ queryKey: ['event', id] });
		},
		onError: () => toast.error('Failed to advance cycle')
	});

	const endEventMut = createMutation({
		mutationFn: () => endEvent({ path: { id } }),
		onSuccess: () => {
			toast.success('Event ended');
			qc.invalidateQueries({ queryKey: ['event', id] });
		},
		onError: () => toast.error('Failed to end event')
	});

	const event = $derived($eventQuery.data);
	const cycles = $derived(event?.cycles ?? []);

	const activeCycle = $derived(
		cycles.find((c: CycleInfo) => c.id === monitorData?.activeCycleId) ?? null
	);

	const activeCycleOrderIndex = $derived(activeCycle?.orderIndex ?? -1);
	const hasMoreCycles = $derived(
		cycles.some((c: CycleInfo) => c.orderIndex > activeCycleOrderIndex)
	);

	function statusColor(status: string) {
		if (status === 'active') return 'var(--ok)';
		if (status === 'ended') return 'var(--t3)';
		return 'var(--brand)';
	}
	function statusBg(status: string) {
		if (status === 'active') return 'var(--ok-soft)';
		if (status === 'ended') return 'var(--s1)';
		return 'var(--brand-soft)';
	}
</script>

<div class="flex flex-col gap-6 py-6 px-4 lg:px-6">
	<!-- Back link -->
	<button
		onclick={() => goto('/app/events')}
		class="self-start text-[13px] hover:underline"
		style="color: var(--t3)"
	>← Back to events</button>

	{#if $eventQuery.isLoading}
		<div class="text-[14px]" style="color: var(--t3)">Loading event…</div>
	{:else if !event}
		<div class="text-[14px]" style="color: var(--danger)">Event not found.</div>
	{:else}
		<!-- Header -->
		<div class="flex items-start justify-between gap-4">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">{event.name}</h1>
				{#if event.description}
					<p class="mt-1 text-[13px]" style="color: var(--t3)">{event.description}</p>
				{/if}
			</div>
			<div class="flex items-center gap-2 shrink-0">
				<span
					class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[12px] font-medium capitalize"
					style="background: {statusBg(event.status)}; color: {statusColor(event.status)}"
				>
					<span class="size-1.5 rounded-full bg-current"></span>
					{event.status}
				</span>
				{#if event.currentStage && event.currentStage !== 'idle'}
					<span
						class="inline-flex items-center rounded-full px-2.5 py-1 text-[12px] font-medium capitalize"
						style="background: var(--s3); color: var(--t2)"
					>
						{event.currentStage.replace('_', ' ')}
					</span>
				{/if}
			</div>
		</div>

		<!-- Monitor panel (active only) -->
		{#if event.status === 'active'}
			<div
				class="rounded-[var(--radius-card)] border p-5 flex items-center gap-8"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				<div>
					<div class="text-[11px] font-semibold uppercase tracking-wide mb-0.5" style="color: var(--t3)">Participants</div>
					<div class="text-[28px] font-bold" style="color: var(--t1)">
						{monitorData?.participantCount ?? '—'}
					</div>
				</div>
				<div style="width: 1px; height: 40px; background: var(--border-h)"></div>
				<div>
					<div class="text-[11px] font-semibold uppercase tracking-wide mb-0.5" style="color: var(--t3)">Responded</div>
					<div class="text-[28px] font-bold" style="color: var(--t1)">
						{monitorData?.respondedCount ?? '—'}
					</div>
				</div>
				{#if activeCycle}
					<div style="width: 1px; height: 40px; background: var(--border-h)"></div>
					<div>
						<div class="text-[11px] font-semibold uppercase tracking-wide mb-0.5" style="color: var(--t3)">Active Cycle</div>
						<div class="text-[16px] font-semibold" style="color: var(--t1)">{activeCycle.name}</div>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Action section -->
		<div
			class="rounded-[var(--radius-card)] border p-6 flex flex-col gap-4"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<h2 class="text-[15px] font-semibold">Host Controls</h2>

			{#if event.status === 'draft'}
				<div class="flex flex-col gap-2">
					<p class="text-[13px]" style="color: var(--t3)">Activate the event so participants can join.</p>
					<button
						onclick={() => $activateMut.mutate()}
						disabled={$activateMut.isPending}
						class="self-start rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
					>
						{$activateMut.isPending ? 'Activating…' : 'Activate Event'}
					</button>
				</div>
			{:else if event.status === 'active' && event.currentStage === 'idle'}
				<div class="flex flex-col gap-3">
					<p class="text-[13px]" style="color: var(--t3)">Choose a cycle to start.</p>
					<select
						bind:value={selectedCycleId}
						class="rounded-[9px] border px-3 py-2.5 text-[13.5px] outline-none transition-shadow self-start min-w-[200px]"
						style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
					>
						<option value={null} disabled selected>Select a cycle…</option>
						{#each cycles as c}
							<option value={c.id}>{c.name}</option>
						{/each}
					</select>
					<button
						onclick={() => $startCycleMut.mutate()}
						disabled={selectedCycleId === null || $startCycleMut.isPending}
						class="self-start rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
					>
						{$startCycleMut.isPending ? 'Starting…' : 'Start Cycle'}
					</button>
				</div>
			{:else if event.status === 'active' && event.currentStage === 'waiting'}
				<div class="flex flex-col gap-2">
					<p class="text-[13px]" style="color: var(--t3)">Open the rating form to participants.</p>
					<button
						onclick={() => $showFormMut.mutate()}
						disabled={$showFormMut.isPending}
						class="self-start rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
					>
						{$showFormMut.isPending ? 'Opening…' : 'Show Form'}
					</button>
				</div>
			{:else if event.status === 'active' && event.currentStage === 'form_open'}
				<div class="flex flex-col gap-3">
					{#if hasMoreCycles}
						<p class="text-[13px]" style="color: var(--t3)">Advance to the next cycle.</p>
						<select
							bind:value={selectedCycleId}
							class="rounded-[9px] border px-3 py-2.5 text-[13.5px] outline-none transition-shadow self-start min-w-[200px]"
							style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
						>
							<option value={null} disabled selected>Select next cycle…</option>
							{#each cycles.filter((c: CycleInfo) => c.orderIndex > activeCycleOrderIndex) as c}
								<option value={c.id}>{c.name}</option>
							{/each}
						</select>
						<button
							onclick={() => $nextCycleMut.mutate()}
							disabled={selectedCycleId === null || $nextCycleMut.isPending}
							class="self-start rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
						>
							{$nextCycleMut.isPending ? 'Advancing…' : 'Next Cycle'}
						</button>
					{:else}
						<p class="text-[13px]" style="color: var(--t3)">All cycles complete. End the event to lock in results.</p>
						<button
							onclick={() => $endEventMut.mutate()}
							disabled={$endEventMut.isPending}
							class="self-start rounded-[9px] px-5 py-2.5 text-[14px] font-medium transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
							style="background: var(--danger-soft); color: var(--danger)"
						>
							{$endEventMut.isPending ? 'Ending…' : 'End Event'}
						</button>
					{/if}
				</div>
			{:else if event.status === 'ended'}
				<div class="flex flex-col gap-2">
					<p class="text-[13px]" style="color: var(--t3)">This event has ended. View the final results.</p>
					<button
						onclick={() => goto(`/app/events/${id}/results`)}
						class="self-start rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90"
					>
						View Results
					</button>
				</div>
			{/if}
		</div>

		<!-- Cycles list -->
		<div
			class="rounded-[var(--radius-card)] border overflow-hidden"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<div class="px-5 py-4 border-b" style="border-color: var(--border-h)">
				<h2 class="text-[15px] font-semibold">Cycles</h2>
			</div>
			{#if cycles.length === 0}
				<div class="px-5 py-6 text-[13px]" style="color: var(--t3)">No cycles defined.</div>
			{:else}
				<div>
					{#each [...cycles].sort((a: CycleInfo, b: CycleInfo) => a.orderIndex - b.orderIndex) as cycle, i}
						<div
							class="flex items-center justify-between px-5 py-3.5"
							style="border-bottom: {i < cycles.length - 1 ? '1px solid var(--border-h)' : 'none'};
								   background: {cycle.id === monitorData?.activeCycleId ? 'var(--brand-soft)' : 'transparent'}"
						>
							<div class="flex items-center gap-3">
								<span
									class="size-6 rounded-full grid place-items-center text-[11px] font-bold shrink-0"
									style="background: {cycle.id === monitorData?.activeCycleId ? 'var(--brand)' : 'var(--s3)'}; color: {cycle.id === monitorData?.activeCycleId ? 'white' : 'var(--t3)'}"
								>{i + 1}</span>
								<span class="text-[14px] font-medium" style="color: {cycle.id === monitorData?.activeCycleId ? 'var(--brand)' : 'var(--t1)'}">{cycle.name}</span>
							</div>
							{#if cycle.id === monitorData?.activeCycleId}
								<span class="text-[12px] font-medium" style="color: var(--brand)">Active</span>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
