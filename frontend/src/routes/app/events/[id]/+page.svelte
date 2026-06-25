<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getEvent, joinEvent, getEventSession, respondToCycle } from '$lib/api-client.js';
	import { getProfile } from '$lib/auth.js';
	import { toast } from 'svelte-sonner';
	import type { EventDetailReadable, SessionResponseReadable, FormInfo } from '$lib/api/types.gen.js';
	import StarRating from '$lib/components/ui/star-rating.svelte';
	import EmojiRating from '$lib/components/ui/emoji-rating.svelte';
	import { pageTitle } from '$lib/page-title.js';

	const MOOD_ITEMS = [
		{ emoji: '😔', label: 'Terrible' },
		{ emoji: '😕', label: 'Poor' },
		{ emoji: '🙂', label: 'Good' },
		{ emoji: '😍', label: 'Amazing' },
	];

	const id = Number($page.params.id);

	let event   = $state<EventDetailReadable | null>(null);
	let session = $state<SessionResponseReadable | null>(null);
	let formValues = $state<Record<number, number | string | null>>({});
	let submitting = $state(false);
	let pollInterval: ReturnType<typeof setInterval> | undefined;

	const joined   = $derived(session?.isParticipant ?? false);
	const isEnded  = $derived(session?.currentStage === 'ended' || event?.status === 'ended');
	const showForm = $derived(joined && session?.currentStage === 'form_open' && !session?.myResponseSubmitted);
	const showWait = $derived(joined && !showForm && !isEnded &&
		(session?.currentStage === 'waiting' || session?.myResponseSubmitted || session?.currentStage === 'idle'));

	const allFilled = $derived(
		session?.forms?.every((f: FormInfo) => {
			const v = formValues[f.id];
			if (v === null || v === undefined) return false;
			if (f.type === 'free_text') return true; // free text is optional
			return typeof v === 'number' && v > 0;
		}) ?? false
	);

	async function loadEvent() {
		const r = await getEvent({ path: { id } });
		if (!r.data) return;
		event = r.data;
		pageTitle.set(event.name);
		const profile = getProfile();
		if (profile && event.hostId === profile.id) goto(`/app/events/${id}/control`);
	}

	async function poll() {
		const r = await getEventSession({ path: { id } });
		if (r.data) {
			const prev = session;
			session = r.data;
			if (r.data.currentStage === 'waiting' && prev?.currentStage !== 'waiting') formValues = {};
		}
	}

	async function handleJoin() {
		const r = await joinEvent({ path: { id } });
		if (r.error) { toast.error('Could not join event'); return }
		await poll();
		if (!pollInterval) pollInterval = setInterval(poll, 3000);
	}

	async function handleSave() {
		if (!session?.forms) return;
		submitting = true;
		const items = session.forms.map((f: FormInfo) => {
			const val = formValues[f.id];
			return f.type === 'free_text'
				? { formId: f.id, valueText: val as string }
				: { formId: f.id, valueNumber: val as number };
		});
		const r = await respondToCycle({ path: { id }, body: { items } });
		submitting = false;
		if (r.error) { toast.error('Failed to save responses'); return }
		session = { ...session, myResponseSubmitted: true };
		formValues = {};
	}

	onMount(async () => {
		await loadEvent();
		if (!event) return;
		await poll();
		pollInterval = setInterval(poll, 3000);
	});
	onDestroy(() => {
		if (pollInterval !== undefined) clearInterval(pollInterval);
		pageTitle.set(null);
	});
</script>

<div class="canvas" style="padding-top:20px">
	{#if !event}
		<div class="stage">
			<div style="color:var(--t3);font-size:14px">Loading…</div>
		</div>

	{:else if isEnded}
		<div class="stage">
			<div class="stage-card">
				<div style="font-size:48px;margin-bottom:20px">🎉</div>
				<h2 style="font-family:var(--display);font-weight:700;font-size:26px;letter-spacing:-.02em">Event has ended</h2>
				<p style="color:var(--t2);margin-top:10px">Thank you for participating!</p>
				<div class="ev" style="margin-top:22px;font-size:12px;color:var(--t3);text-transform:uppercase;letter-spacing:.1em">{event.name}</div>
				<button class="btn primary" style="margin-top:24px" onclick={() => goto('/app/dashboard')}>Back to dashboard</button>
			</div>
		</div>

	{:else if event.status === 'draft'}
		<div class="stage">
			<div class="stage-card">
				<div class="beacon"></div>
				<h2 style="font-family:var(--display);font-weight:700;font-size:26px;letter-spacing:-.02em">Not started yet</h2>
				<p style="color:var(--t2);margin-top:10px">The host hasn't activated this event yet.</p>
				<div class="ev" style="margin-top:22px;font-size:12px;color:var(--t3);text-transform:uppercase;letter-spacing:.1em">{event.name}</div>
			</div>
		</div>

	{:else if event.status === 'active' && !joined}
		<div class="stage">
			<div class="stage-card">
				<div class="beacon"></div>
				<h2 style="font-family:var(--display);font-weight:700;font-size:26px;letter-spacing:-.02em">{event.name}</h2>
				{#if event.description}
					<p style="color:var(--t2);margin-top:10px">{event.description}</p>
				{/if}
				<div class="ev" style="margin-top:22px;font-size:12px;color:var(--t3);text-transform:uppercase;letter-spacing:.1em">Tap to join the session</div>
				<button class="btn primary lg" style="margin-top:24px" onclick={handleJoin}>Join event</button>
			</div>
		</div>

	{:else if showWait}
		<div class="stage">
			<div class="stage-card">
				<div class="beacon"></div>
				<h2 style="font-family:var(--display);font-weight:700;font-size:26px;letter-spacing:-.02em">You're in.</h2>
				<p style="color:var(--t2);margin-top:10px">
					{session?.myResponseSubmitted
						? 'Your scores are locked. Waiting for the next cycle.'
						: 'Sit tight — the host will open the form when the panel begins.'}
				</p>
				{#if session?.activeCycleName}
					<p style="margin-top:12px;font-size:13px;color:var(--brand-2);font-weight:600">{session.activeCycleName}</p>
				{/if}
				<div class="ev" style="margin-top:22px;font-size:12px;color:var(--t3);text-transform:uppercase;letter-spacing:.1em">{event.name}</div>
			</div>
		</div>

	{:else if showForm && session}
		<div class="canvas" style="padding:32px 26px 80px">
			<div class="wrap narrow">
				<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px">
					<div class="eyebrow" style="margin:0">{session.activeCycleName ?? 'Current cycle'}</div>
					<span class="pill live"><span class="dot"></span>Live</span>
				</div>
				<h1 class="title">{session.activeCycleName ?? 'Score'}</h1>
				<p class="sub" style="margin-bottom:28px">Score every question, then submit. You can't change it after.</p>

				<div class="card">
					<div class="formstack">
						{#each (session.forms ?? []) as form (form.id)}
							<div class="qblock">
								<div class="q">{form.label}</div>

								{#if form.type === 'rating'}
									<div style="display:flex;justify-content:center">
										<StarRating
											totalStars={5}
											size="lg"
											value={formValues[form.id] as number ?? 0}
											onRate={(v) => { formValues[form.id] = v }}
										/>
									</div>

								{:else if form.type === 'mood'}
									<EmojiRating
										items={MOOD_ITEMS}
										value={formValues[form.id] as number ?? 0}
										onRate={(v) => { formValues[form.id] = v }}
									/>

								{:else}
									<textarea
										class="ipt"
										rows={3}
										placeholder="Optional — what stood out?"
										value={formValues[form.id] as string ?? ''}
										oninput={(e) => { formValues[form.id] = (e.currentTarget as HTMLTextAreaElement).value }}
									></textarea>
								{/if}
							</div>
						{/each}
					</div>

					<div class="save-bar">
						<span class="hint">Submitting locks all answers for this cycle.</span>
						<button
							class="btn primary lg"
							disabled={!allFilled || submitting}
							onclick={handleSave}
						>{submitting ? 'Submitting…' : 'Submit scores'}</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
