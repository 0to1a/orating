<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { listEvents } from '$lib/api-client.js';
	import { getProfile } from '$lib/auth.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import type { EventInfo } from '$lib/api/types.gen.js';

	const profileId = $derived(getProfile()?.id ?? -1);

	const eventsQuery = createQuery({
		queryKey: ['events'],
		queryFn: () => listEvents().then((r) => r.data?.events ?? [])
	});

	function statusColor(status: string): string {
		if (status === 'active') return 'var(--ok)';
		if (status === 'ended') return 'var(--t3)';
		return 'var(--brand)';
	}

	function statusBg(status: string): string {
		if (status === 'active') return 'var(--ok-soft)';
		if (status === 'ended') return 'var(--s1)';
		return 'var(--brand-soft)';
	}

	function handleCardClick(event: EventInfo) {
		if (event.hostId === profileId) {
			goto(`/app/events/${event.id}/control`);
		} else {
			goto(`/app/events/${event.id}`);
		}
	}
</script>

<div class="flex flex-col gap-6 py-6">
	<!-- Page header -->
	<div class="flex items-end justify-between px-4 lg:px-6">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Events</h1>
			<p class="text-muted-foreground mt-1 text-sm">Browse and manage live rating events.</p>
		</div>
		<button
			onclick={() => goto('/app/events/new')}
			class="inline-flex items-center gap-2 rounded-[9px] bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
		>
			+ Create Event
		</button>
	</div>

	<div class="px-4 lg:px-6">
		{#if $eventsQuery.isLoading}
			<div class="flex flex-col gap-3">
				<Skeleton class="h-20 w-full" />
				<Skeleton class="h-20 w-full" />
				<Skeleton class="h-20 w-full" />
			</div>
		{:else if ($eventsQuery.data ?? []).length === 0}
			<div
				class="rounded-[var(--radius-card)] border py-16 text-center text-sm"
				style="background: var(--s2); border-color: var(--border-h); color: var(--t3)"
			>
				No events yet — create one to get started.
			</div>
		{:else}
			<div class="flex flex-col gap-3">
				{#each $eventsQuery.data ?? [] as event}
					<button
						onclick={() => handleCardClick(event)}
						class="group w-full rounded-[var(--radius-card)] border px-5 py-4 text-left transition-all hover:shadow-sm"
						style="background: var(--s2); border-color: var(--border-h)"
					>
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2.5 flex-wrap">
									<span class="text-[15px] font-semibold truncate" style="color: var(--t1)">{event.name}</span>
									{#if event.hostId === profileId}
										<span
											class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium"
											style="background: var(--brand-soft); color: var(--brand)"
										>My event</span>
									{/if}
								</div>
								{#if event.description}
									<p class="mt-1 text-[13px] truncate" style="color: var(--t3)">{event.description}</p>
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
								<span
									class="inline-flex items-center rounded-full px-2.5 py-1 text-[12px] font-medium capitalize"
									style="background: var(--s3); color: var(--t2)"
								>
									{event.visibility}
								</span>
							</div>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>
</div>
