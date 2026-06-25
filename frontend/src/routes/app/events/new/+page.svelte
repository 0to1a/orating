<script lang="ts">
	import { createMutation } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { createEvent } from '$lib/api-client.js';
	import { toast } from 'svelte-sonner';

	let name = $state('');
	let description = $state('');
	let visibility = $state<'public' | 'private'>('public');
	let cycles = $state<{ name: string }[]>([{ name: '' }]);
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
					members:
						visibility === 'private'
							? memberEmails.filter((e) => e.trim())
							: null
				}
			}),
		onSuccess: (r) => {
			const id = r.data?.id;
			if (id) goto(`/app/events/${id}/control`);
		},
		onError: () => toast.error('Failed to create event')
	});

	function addCycle() {
		cycles = [...cycles, { name: '' }];
	}
	function removeCycle(i: number) {
		if (cycles.length > 1) cycles = cycles.filter((_, idx) => idx !== i);
	}

	function addForm() {
		forms = [...forms, { type: 'rating', label: '' }];
	}
	function removeForm(i: number) {
		if (forms.length > 1) forms = forms.filter((_, idx) => idx !== i);
	}

	function addEmail() {
		memberEmails = [...memberEmails, ''];
	}
	function removeEmail(i: number) {
		if (memberEmails.length > 1) memberEmails = memberEmails.filter((_, idx) => idx !== i);
	}

	const inputStyle =
		'background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit';
	const inputClass =
		'w-full rounded-[9px] border px-3 py-2.5 text-[13.5px] outline-none transition-shadow';
	const labelClass = 'mb-1.5 block text-[12.5px] font-medium';

	function focusStyle(e: FocusEvent) {
		(e.currentTarget as HTMLElement).style.boxShadow = '0 0 0 3px var(--brand-soft)';
	}
	function blurStyle(e: FocusEvent) {
		(e.currentTarget as HTMLElement).style.boxShadow = '';
	}
</script>

<div class="mx-auto max-w-2xl py-8 px-4 lg:px-0 flex flex-col gap-8">
	<!-- Header -->
	<div>
		<button
			onclick={() => goto('/app/events')}
			class="mb-3 text-[13px] hover:underline"
			style="color: var(--t3)"
		>← Back to events</button>
		<h1 class="text-2xl font-semibold tracking-tight">Create Event</h1>
	</div>

	<!-- Section 1: Event Info -->
	<section
		class="rounded-[var(--radius-card)] border p-6 flex flex-col gap-4"
		style="background: var(--s2); border-color: var(--border-h)"
	>
		<h2 class="text-[15px] font-semibold">Event Info</h2>

		<div>
			<label class="{labelClass}" style="color: var(--t2)" for="event-name">Name *</label>
			<input
				id="event-name"
				bind:value={name}
				placeholder="e.g. Q2 Retrospective"
				class={inputClass}
				style={inputStyle}
				onfocus={focusStyle}
				onblur={blurStyle}
			/>
		</div>

		<div>
			<label class="{labelClass}" style="color: var(--t2)" for="event-desc">Description</label>
			<textarea
				id="event-desc"
				bind:value={description}
				placeholder="Optional description"
				rows={3}
				class="{inputClass} resize-none"
				style={inputStyle}
				onfocus={focusStyle}
				onblur={blurStyle}
			></textarea>
		</div>

		<div>
			<span class="{labelClass}" style="color: var(--t2)">Visibility</span>
			<div class="flex gap-2 mt-1">
				<button
					onclick={() => (visibility = 'public')}
					class="rounded-[9px] border px-4 py-2 text-[13px] font-medium transition-all"
					style={visibility === 'public'
						? 'background: var(--brand-soft); border-color: var(--brand); color: var(--brand)'
						: 'background: var(--s3); border-color: var(--border-h2); color: var(--t2)'}
				>Public</button>
				<button
					onclick={() => (visibility = 'private')}
					class="rounded-[9px] border px-4 py-2 text-[13px] font-medium transition-all"
					style={visibility === 'private'
						? 'background: var(--brand-soft); border-color: var(--brand); color: var(--brand)'
						: 'background: var(--s3); border-color: var(--border-h2); color: var(--t2)'}
				>Private</button>
			</div>
		</div>
	</section>

	<!-- Section 2: Cycles -->
	<section
		class="rounded-[var(--radius-card)] border p-6 flex flex-col gap-4"
		style="background: var(--s2); border-color: var(--border-h)"
	>
		<h2 class="text-[15px] font-semibold">Cycles</h2>

		<div class="flex flex-col gap-2">
			{#each cycles as cycle, i}
				<div class="flex items-center gap-2">
					<input
						bind:value={cycle.name}
						placeholder="Cycle name, e.g. Round 1"
						class="{inputClass} flex-1"
						style={inputStyle}
						onfocus={focusStyle}
						onblur={blurStyle}
					/>
					<button
						onclick={() => removeCycle(i)}
						disabled={cycles.length === 1}
						class="size-9 grid place-items-center rounded-lg border transition-colors disabled:opacity-30 disabled:cursor-not-allowed hover:bg-[var(--danger-soft)] hover:text-[var(--danger)]"
						style="border-color: var(--border-h2); color: var(--t3)"
						aria-label="Remove cycle"
					>×</button>
				</div>
			{/each}
		</div>

		<button
			onclick={addCycle}
			class="self-start rounded-[9px] border px-3 py-1.5 text-[13px] transition-colors hover:bg-[var(--s3)]"
			style="border-color: var(--border-h2); color: var(--t2)"
		>+ Add Cycle</button>
	</section>

	<!-- Section 3: Forms -->
	<section
		class="rounded-[var(--radius-card)] border p-6 flex flex-col gap-4"
		style="background: var(--s2); border-color: var(--border-h)"
	>
		<h2 class="text-[15px] font-semibold">Forms</h2>

		<div class="flex flex-col gap-2">
			{#each forms as form, i}
				<div class="flex items-center gap-2">
					<select
						bind:value={form.type}
						class="rounded-[9px] border px-3 py-2.5 text-[13px] outline-none transition-shadow shrink-0"
						style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
						onfocus={focusStyle}
						onblur={blurStyle}
					>
						<option value="rating">Rating</option>
						<option value="mood">Mood</option>
						<option value="free_text">Free Text</option>
					</select>
					<input
						bind:value={form.label}
						placeholder="Form label, e.g. Overall score"
						class="{inputClass} flex-1"
						style={inputStyle}
						onfocus={focusStyle}
						onblur={blurStyle}
					/>
					<button
						onclick={() => removeForm(i)}
						disabled={forms.length === 1}
						class="size-9 grid place-items-center rounded-lg border transition-colors disabled:opacity-30 disabled:cursor-not-allowed hover:bg-[var(--danger-soft)] hover:text-[var(--danger)]"
						style="border-color: var(--border-h2); color: var(--t3)"
						aria-label="Remove form"
					>×</button>
				</div>
			{/each}
		</div>

		<button
			onclick={addForm}
			class="self-start rounded-[9px] border px-3 py-1.5 text-[13px] transition-colors hover:bg-[var(--s3)]"
			style="border-color: var(--border-h2); color: var(--t2)"
		>+ Add Form</button>
	</section>

	<!-- Section 4: Members (private only) -->
	{#if visibility === 'private'}
		<section
			class="rounded-[var(--radius-card)] border p-6 flex flex-col gap-4"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<div>
				<h2 class="text-[15px] font-semibold">Members</h2>
				<p class="text-[12.5px] mt-0.5" style="color: var(--t3)">Invite specific people to this private event.</p>
			</div>

			<div class="flex flex-col gap-2">
				{#each memberEmails as _, i}
					<div class="flex items-center gap-2">
						<input
							bind:value={memberEmails[i]}
							type="email"
							placeholder="email@example.com"
							class="{inputClass} flex-1"
							style={inputStyle}
							onfocus={focusStyle}
							onblur={blurStyle}
						/>
						<button
							onclick={() => removeEmail(i)}
							disabled={memberEmails.length === 1}
							class="size-9 grid place-items-center rounded-lg border transition-colors disabled:opacity-30 disabled:cursor-not-allowed hover:bg-[var(--danger-soft)] hover:text-[var(--danger)]"
							style="border-color: var(--border-h2); color: var(--t3)"
							aria-label="Remove email"
						>×</button>
					</div>
				{/each}
			</div>

			<button
				onclick={addEmail}
				class="self-start rounded-[9px] border px-3 py-1.5 text-[13px] transition-colors hover:bg-[var(--s3)]"
				style="border-color: var(--border-h2); color: var(--t2)"
			>+ Add Email</button>
		</section>
	{/if}

	<!-- Submit -->
	<div class="flex justify-end pb-6">
		<button
			onclick={() => $createMut.mutate()}
			disabled={!name.trim() || $createMut.isPending}
			class="rounded-[9px] bg-primary px-6 py-2.5 text-[14px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
		>
			{$createMut.isPending ? 'Creating…' : 'Create Event'}
		</button>
	</div>
</div>
