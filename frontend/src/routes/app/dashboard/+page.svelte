<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { getProfile } from '$lib/auth.js';
	import { listEvents } from '$lib/api-client.js';
	import type { EventInfo } from '$lib/api/types.gen.js';

	const profile = $derived(getProfile());

	const eventsQuery = createQuery({
		queryKey: ['events'],
		queryFn: () => listEvents().then((r) => r.data?.events ?? [])
	});

	const events = $derived(($eventsQuery.data ?? []) as EventInfo[]);
	const activeEvents = $derived(events.filter((e) => e.status === 'active'));
	const draftEvents = $derived(events.filter((e) => e.status === 'draft'));

	function isHost(e: EventInfo) {
		return profile && e.hostId === profile.id;
	}

	function eventHref(e: EventInfo): string {
		if (isHost(e)) {
			if (e.status === 'ended') return `/app/events/${e.id}/results`;
			return `/app/events/${e.id}/control`;
		}
		if (e.status === 'active') return `/app/events/${e.id}`;
		return '';
	}

	function eventActionLabel(e: EventInfo): string {
		if (isHost(e)) {
			if (e.status === 'ended') return 'Results';
			if (e.status === 'active') return 'Control';
			return 'Setup';
		}
		if (e.status === 'active') return 'Open';
		return '';
	}

	function statusIcon(e: EventInfo, i: number): string {
		if (e.status === 'active') return '▶';
		if (e.status === 'ended') return '✓';
		return String(i + 1);
	}
</script>

<div class="canvas">
	<div class="wrap">
		<!-- Header -->
		<div style="display:flex;align-items:flex-end;justify-content:space-between;gap:20px;flex-wrap:wrap;margin-bottom:28px">
			<div>
				<div class="eyebrow">Workspace overview</div>
				<h1 class="title">Welcome back, {profile?.name?.split(' ')[0] ?? 'there'}</h1>
				<p class="sub">Manage your rating events across your workspace.</p>
			</div>
			<button class="btn primary" onclick={() => goto('/app/events/new')}>
				<svg viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
				New event
			</button>
		</div>

		<!-- Stat tiles -->
		<div class="grid stats-3 gap-[18px] mb-[18px]" style="grid-template-columns:repeat(3,1fr)">
			<div class="stat">
				<div class="k">
					Active events
					<svg viewBox="0 0 24 24"><polygon points="12 2 15 9 22 9 16 14 18 21 12 17 6 21 8 14 2 9 9 9"/></svg>
				</div>
				<div class="v">{activeEvents.length}</div>
				<div class="trend">{activeEvents.length === 1 ? '1 live now' : activeEvents.length > 0 ? `${activeEvents.length} live now` : 'none live'}</div>
			</div>
			<div class="stat">
				<div class="k">
					Total events
					<svg viewBox="0 0 24 24"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
				</div>
				<div class="v">{events.length}</div>
				<div class="trend">{events.length === 0 ? 'create your first' : `${draftEvents.length} draft`}</div>
			</div>
			<div class="stat hide-mobile">
				<div class="k">
					Draft events
					<svg viewBox="0 0 24 24"><path d="M3 3v18h18"/><path d="M7 14l4-4 3 3 5-6"/></svg>
				</div>
				<div class="v">{draftEvents.length}</div>
				<div class="trend">ready to activate</div>
			</div>
		</div>

		<!-- Events card -->
		<div class="card">
			<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:18px">
				<div class="card-h">Your events</div>
				<button
					onclick={() => goto('/app/events/new')}
					style="font-size:12.5px;color:var(--brand-2);cursor:pointer;background:none;border:none;padding:0"
				>+ New event</button>
			</div>

			{#if $eventsQuery.isLoading}
				<div style="color:var(--t3);font-size:14px;padding:24px 0">Loading…</div>
			{:else if events.length === 0}
				<div style="color:var(--t3);font-size:14px;padding:32px 0;text-align:center">
					No events yet.
					<button
						onclick={() => goto('/app/events/new')}
						style="color:var(--brand-2);background:none;border:none;cursor:pointer;margin-left:4px"
					>Create your first →</button>
				</div>
			{:else}
				<div style="display:flex;flex-direction:column;gap:12px">
					{#each events as event, i}
						{@const href = eventHref(event)}
						{@const label = eventActionLabel(event)}
						<button
							class="ev-row"
							style="width:100%;text-align:left;cursor:{href ? 'pointer' : 'default'}"
							onclick={() => href && goto(href)}
						>
							<div style="display:flex;align-items:center;gap:14px">
								<div class="ev-icon">{statusIcon(event, i)}</div>
								<div>
									<div style="font-weight:600;color:var(--t1)">{event.name}</div>
									<div style="font-size:12px;color:var(--t3)">
										{event.visibility === 'private' ? 'Private' : 'Public'} ·
										{isHost(event) ? 'Host' : 'Rater'}
									</div>
								</div>
							</div>
							<div style="display:flex;align-items:center;gap:14px">
								{#if event.status === 'active'}
									<span class="pill live"><span class="dot"></span>Live</span>
								{:else if event.status === 'ended'}
									<span class="pill draft" style="color:var(--t3)"><span class="dot"></span>Ended</span>
								{:else}
									<span class="pill draft"><span class="dot"></span>Draft</span>
								{/if}
								{#if label}
									<span class="btn ghost" style="padding:8px 14px;pointer-events:none">{label}</span>
								{/if}
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>
