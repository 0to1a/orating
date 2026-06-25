<script lang="ts">
	import { createMutation } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { createEvent } from '$lib/api-client.js';
	import { toast } from 'svelte-sonner';

	let name = $state('');
	let description = $state('');
	let visibility = $state<'public' | 'private'>('public');
	let cycles = $state<{ name: string }[]>([{ name: '' }, { name: '' }]);
	let forms = $state<{ type: string; label: string }[]>([{ type: 'rating', label: '' }]);
	let memberEmails = $state<string[]>(['']);

	const createMut = createMutation({
		mutationFn: () =>
			createEvent({
				body: {
					name,
					description: description || undefined,
					visibility,
					cycles: cycles.filter((c) => c.name.trim()).map((c) => ({ name: c.name })),
					forms: forms.filter((f) => f.label.trim()).map((f) => ({ type: f.type, label: f.label })),
					members: visibility === 'private' ? memberEmails.filter((e) => e.trim()) : null
				}
			}),
		onSuccess: (r) => {
			const id = r.data?.id;
			if (id) goto(`/app/events/${id}/control`);
		},
		onError: () => toast.error('Failed to create event')
	});

	function addCycle() { cycles = [...cycles, { name: '' }] }
	function removeCycle(i: number) { if (cycles.length > 1) cycles = cycles.filter((_, idx) => idx !== i) }
	function addForm() { forms = [...forms, { type: 'rating', label: '' }] }
	function removeForm(i: number) { if (forms.length > 1) forms = forms.filter((_, idx) => idx !== i) }
	function addEmail() { memberEmails = [...memberEmails, ''] }
	function removeEmail(i: number) { if (memberEmails.length > 1) memberEmails = memberEmails.filter((_, idx) => idx !== i) }
</script>

<div class="canvas">
	<div class="wrap narrow">
		<button class="back" onclick={() => goto('/app/dashboard')}>← Back to events</button>

		<div class="eyebrow">New event</div>
		<h1 class="title">Set up your panel</h1>
		<p class="sub" style="margin-bottom:28px">Define the rounds and the questions raters will score.</p>

		<!-- Event info -->
		<div class="card" style="margin-bottom:18px">
			<div class="card-h" style="margin-bottom:18px">Event info</div>

			<div style="margin-bottom:16px">
				<label class="fld" for="ev-name">Name</label>
				<input id="ev-name" class="ipt" bind:value={name} placeholder="e.g. Pitch Night — Cohort 12" />
			</div>

			<div style="margin-bottom:16px">
				<label class="fld" for="ev-desc">Description</label>
				<textarea id="ev-desc" class="ipt" bind:value={description} placeholder="What are raters evaluating?"></textarea>
			</div>

			<div>
				<div class="fld" role="group" aria-label="Who can join">Who can join</div>
				<div class="seg">
					<button class:on={visibility === 'public'} onclick={() => (visibility = 'public')}>Public</button>
					<button class:on={visibility === 'private'} onclick={() => (visibility = 'private')}>Invite only</button>
				</div>
				<p style="font-size:12px;color:var(--t3);margin-top:8px">
					{visibility === 'public'
						? "Public events show on everyone's dashboard."
						: 'Invite-only stays with the people you add.'}
				</p>
			</div>
		</div>

		<!-- Cycles -->
		<div class="card" style="margin-bottom:18px">
			<div class="card-h" style="margin-bottom:8px">Cycles</div>
			<p style="font-size:12.5px;color:var(--t3);margin-bottom:18px">One per thing being judged — a team, a sample, a pitch.</p>

			{#each cycles as cycle, i}
				<div class="row-b">
					<div class="num">{i + 1}</div>
					<input class="ipt" bind:value={cycle.name} placeholder="Cycle name…" />
					<button class="x" onclick={() => removeCycle(i)} aria-label="Remove cycle">×</button>
				</div>
			{/each}
			<button class="add-row" onclick={addCycle}>+ Add cycle</button>
		</div>

		<!-- Scoring form -->
		<div class="card" style="margin-bottom:24px">
			<div class="card-h" style="margin-bottom:8px">Scoring form</div>
			<p style="font-size:12.5px;color:var(--t3);margin-bottom:18px">The same questions apply to every cycle.</p>

			{#each forms as form, i}
				<div class="row-b">
					<select class="ipt" style="max-width:148px;flex-shrink:0" bind:value={form.type}>
						<option value="rating">Rating 1–5</option>
						<option value="mood">Mood 1–4</option>
						<option value="free_text">Free text</option>
					</select>
					<input class="ipt" bind:value={form.label} placeholder="Question label…" />
					<button class="x" onclick={() => removeForm(i)} aria-label="Remove question">×</button>
				</div>
			{/each}
			<button class="add-row" onclick={addForm}>+ Add question</button>
		</div>

		<!-- Members (private only) -->
		{#if visibility === 'private'}
			<div class="card" style="margin-bottom:24px">
				<div class="card-h" style="margin-bottom:8px">Invite members</div>
				<p style="font-size:12.5px;color:var(--t3);margin-bottom:18px">Add email addresses of people who can join this event.</p>

				{#each memberEmails as _, i}
					<div class="row-b">
						<input
							class="ipt"
							type="email"
							bind:value={memberEmails[i]}
							placeholder="email@example.com"
						/>
						<button class="x" onclick={() => removeEmail(i)} aria-label="Remove email">×</button>
					</div>
				{/each}
				<button class="add-row" onclick={addEmail}>+ Add email</button>
			</div>
		{/if}

		<!-- Actions -->
		<div style="display:flex;justify-content:flex-end;gap:12px;padding-bottom:40px">
			<button class="btn ghost" onclick={() => goto('/app/dashboard')}>Cancel</button>
			<button
				class="btn primary lg"
				disabled={!name.trim() || $createMut.isPending}
				onclick={() => $createMut.mutate()}
			>
				{$createMut.isPending ? 'Creating…' : 'Create event'}
			</button>
		</div>
	</div>
</div>
