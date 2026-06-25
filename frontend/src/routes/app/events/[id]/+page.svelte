<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getEvent, joinEvent, getEventSession, respondToCycle } from '$lib/api-client.js';
	import { getProfile } from '$lib/auth.js';
	import { toast } from 'svelte-sonner';
	import type { EventDetailReadable, SessionResponseReadable, FormInfo } from '$lib/api/types.gen.js';

	const id = Number($page.params.id);

	let event = $state<EventDetailReadable | null>(null);
	let session = $state<SessionResponseReadable | null>(null);
	let formValues = $state<Record<number, number | string | null>>({});
	let submitting = $state(false);
	let pollInterval: ReturnType<typeof setInterval> | undefined;

	// Source of truth for join status comes from the backend, not local state.
	const joined = $derived(session?.isParticipant ?? false);

	async function loadEvent() {
		const r = await getEvent({ path: { id } });
		if (!r.data) return;
		event = r.data;

		const profile = getProfile();
		if (profile && event.hostId === profile.id) {
			goto(`/app/events/${id}/control`);
		}
	}

	async function poll() {
		const r = await getEventSession({ path: { id } });
		if (r.data) {
			const prev = session;
			session = r.data;
			// Reset form values when a new waiting stage begins
			if (r.data.currentStage === 'waiting' && prev?.currentStage !== 'waiting') {
				formValues = {};
			}
		}
	}

	async function handleJoin() {
		const r = await joinEvent({ path: { id } });
		if (r.error) {
			toast.error('Could not join event');
			return;
		}
		// Poll immediately so session.isParticipant updates
		await poll();
		if (!pollInterval) {
			pollInterval = setInterval(poll, 3000);
		}
	}

	async function handleSave() {
		if (!session?.forms) return;
		submitting = true;

		const items = session.forms.map((f: FormInfo) => {
			const val = formValues[f.id];
			if (f.type === 'free_text') {
				return { formId: f.id, valueText: val as string };
			} else {
				return { formId: f.id, valueNumber: val as number };
			}
		});

		const r = await respondToCycle({ path: { id }, body: { items } });
		submitting = false;
		if (r.error) {
			toast.error('Failed to save responses');
			return;
		}
		// Immediately show waiting screen — don't wait for poll
		session = { ...session, myResponseSubmitted: true };
		formValues = {};
	}

	const allFilled = $derived(
		session?.forms?.every((f: FormInfo) => {
			const val = formValues[f.id];
			return val !== null && val !== undefined && val !== '';
		}) ?? false
	);

	const isEnded = $derived(
		session?.currentStage === 'ended' || event?.status === 'ended'
	);

	const showForm = $derived(
		joined &&
		session?.currentStage === 'form_open' &&
		!session?.myResponseSubmitted
	);

	const showWaiting = $derived(
		joined &&
		!showForm &&
		!isEnded &&
		(session?.currentStage === 'waiting' || session?.myResponseSubmitted)
	);

	onMount(async () => {
		await loadEvent();
		if (!event) return;

		// Initial session fetch — isParticipant in response tells us if already joined
		await poll();

		// Start polling
		pollInterval = setInterval(poll, 3000);
	});

	onDestroy(() => {
		if (pollInterval !== undefined) clearInterval(pollInterval);
	});
</script>

<div class="flex flex-col items-center justify-center min-h-[60vh] px-4">
	{#if !event}
		<!-- Loading -->
		<div class="text-[14px]" style="color: var(--t3)">Loading...</div>

	{:else if isEnded}
		<!-- Event ended -->
		<div
			class="w-full max-w-md rounded-[var(--radius-card)] border p-8 flex flex-col items-center gap-4 text-center"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<div class="text-4xl">🎉</div>
			<h1 class="text-xl font-semibold" style="color: var(--t1)">Event has ended</h1>
			<p class="text-[14px]" style="color: var(--t3)">Thank you for participating!</p>
			<button
				onclick={() => goto('/app/events')}
				class="mt-2 rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90"
			>
				Back to events
			</button>
		</div>

	{:else if event.status === 'draft'}
		<!-- Draft — not started -->
		<div
			class="w-full max-w-md rounded-[var(--radius-card)] border p-8 flex flex-col items-center gap-3 text-center"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<p class="text-[14px]" style="color: var(--t3)">This event hasn't started yet.</p>
			<button
				onclick={() => goto('/app/events')}
				class="mt-1 text-[13px] hover:underline"
				style="color: var(--brand)"
			>
				← Back to events
			</button>
		</div>

	{:else if event.status === 'active' && !joined}
		<!-- Pre-join -->
		<div
			class="w-full max-w-md rounded-[var(--radius-card)] border p-8 flex flex-col gap-5"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<div>
				<h2 class="text-xl font-semibold" style="color: var(--t1)">{event.name}</h2>
				{#if event.description}
					<p class="mt-2 text-[14px]" style="color: var(--t3)">{event.description}</p>
				{/if}
			</div>
			<button
				onclick={handleJoin}
				class="self-start rounded-[9px] bg-primary px-5 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90"
			>
				Join Event
			</button>
		</div>

	{:else if showWaiting}
		<!-- Waiting for host -->
		<div
			class="w-full max-w-md rounded-[var(--radius-card)] border p-8 flex flex-col items-center gap-4 text-center"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<!-- Animated pulse indicator -->
			<div class="relative flex items-center justify-center">
				<div
					class="size-12 rounded-full opacity-30 animate-ping absolute"
					style="background: var(--brand)"
				></div>
				<div class="size-6 rounded-full" style="background: var(--brand)"></div>
			</div>
			<h2 class="text-lg font-semibold mt-2" style="color: var(--t1)">Waiting for host...</h2>
			{#if session?.activeCycleName}
				<p class="text-[14px] font-medium" style="color: var(--brand)">{session.activeCycleName}</p>
			{/if}
			<p class="text-[13px]" style="color: var(--t3)">The host will open the form when ready.</p>
		</div>

	{:else if showForm && session}
		<!-- Form is open -->
		<div class="w-full max-w-md flex flex-col gap-6">
			<div>
				<button
					onclick={() => goto('/app/events')}
					class="text-[13px] hover:underline mb-4 block"
					style="color: var(--t3)"
				>← Back to events</button>
				<h2 class="text-xl font-semibold" style="color: var(--t1)">
					{session.activeCycleName || 'Current cycle'}
				</h2>
				<p class="text-[13px] mt-1" style="color: var(--t3)">Fill in all fields to submit.</p>
			</div>

			<div
				class="rounded-[var(--radius-card)] border p-6 flex flex-col gap-6"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				{#each (session.forms ?? []) as form (form.id)}
					<div>
						<p class="text-[14px] font-medium mb-2" style="color: var(--t1)">{form.label}</p>

						{#if form.type === 'rating'}
							<div class="flex gap-2 flex-wrap" role="group" aria-label={form.label}>
								{#each [1, 2, 3, 4, 5] as val}
									<button
										class="w-10 h-10 rounded-full border-2 text-[14px] font-semibold transition-all"
										style="
											background: {formValues[form.id] === val ? 'var(--brand)' : 'var(--s3)'};
											border-color: {formValues[form.id] === val ? 'var(--brand)' : 'var(--border-h2)'};
											color: {formValues[form.id] === val ? 'white' : 'var(--t2)'}
										"
										aria-pressed={formValues[form.id] === val}
										onclick={() => { formValues[form.id] = val; }}
									>{val}</button>
								{/each}
							</div>

						{:else if form.type === 'mood'}
							<div class="flex gap-2 flex-wrap" role="group" aria-label={form.label}>
								{#each [['😞', 1], ['😐', 2], ['😊', 3], ['😄', 4]] as [emoji, val]}
									<button
										class="text-2xl p-2 rounded-[9px] border-2 transition-all"
										style="
											border-color: {formValues[form.id] === val ? 'var(--brand)' : 'var(--border-h2)'};
											background: {formValues[form.id] === val ? 'var(--brand-soft)' : 'var(--s3)'}
										"
										aria-pressed={formValues[form.id] === val}
										onclick={() => { formValues[form.id] = val as number; }}
									>{emoji}</button>
								{/each}
							</div>

						{:else if form.type === 'free_text'}
							<textarea
								id="form-{form.id}"
								aria-label={form.label}
								class="w-full rounded-[9px] border px-3 py-2.5 text-[13.5px] outline-none transition-shadow resize-none"
								style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
								rows={3}
								placeholder="Your response..."
								value={formValues[form.id] as string ?? ''}
								onchange={(e) => { formValues[form.id] = (e.currentTarget as HTMLTextAreaElement).value; }}
								oninput={(e) => { formValues[form.id] = (e.currentTarget as HTMLTextAreaElement).value; }}
							></textarea>
						{/if}
					</div>
				{/each}

				<button
					disabled={!allFilled || submitting}
					onclick={handleSave}
					class="w-full rounded-[9px] bg-primary py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
				>
					{submitting ? 'Saving...' : 'Save All'}
				</button>
			</div>
		</div>

	{:else if joined}
		<!-- Joined but stage is idle — waiting for host to start a cycle -->
		<div
			class="w-full max-w-md rounded-[var(--radius-card)] border p-8 flex flex-col items-center gap-4 text-center"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<div class="relative flex items-center justify-center">
				<div
					class="size-12 rounded-full opacity-30 animate-ping absolute"
					style="background: var(--brand)"
				></div>
				<div class="size-6 rounded-full" style="background: var(--brand)"></div>
			</div>
			<h2 class="text-lg font-semibold mt-2" style="color: var(--t1)">You've joined!</h2>
			<p class="text-[13px]" style="color: var(--t3)">Waiting for the host to begin.</p>
		</div>
	{/if}
</div>
