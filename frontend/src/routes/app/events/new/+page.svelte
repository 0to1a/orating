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
	let memberChips = $state<string[]>([]);
	let memberInput = $state('');
	let memberTagEl = $state<HTMLInputElement | null>(null);
	let memberTagFocused = $state(false);

	function parseEmails(raw: string): string[] {
		return raw
			.split(/[,\n]+/)
			.map((s) => s.trim())
			.filter(Boolean)
			.flatMap((entry) => {
				const match = entry.match(/^.+?\s*<([^>]+)>\s*$/);
				const email = (match ? match[1] : entry).trim().toLowerCase();
				return email.includes('@') ? [email] : [];
			});
	}

	function addMemberEntries(raw: string) {
		const existing = new Set(memberChips);
		const fresh = parseEmails(raw).filter((e) => !existing.has(e));
		memberChips = [...memberChips, ...fresh];
	}

	function removeMemberChip(email: string) {
		memberChips = memberChips.filter((e) => e !== email);
	}

	function handleMemberKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ',' || e.key === 'Tab') {
			if (memberInput.trim()) {
				e.preventDefault();
				addMemberEntries(memberInput);
				memberInput = '';
			}
		} else if (e.key === 'Backspace' && !memberInput && memberChips.length > 0) {
			memberChips = memberChips.slice(0, -1);
		}
	}

	function handleMemberPaste(e: ClipboardEvent) {
		e.preventDefault();
		addMemberEntries(e.clipboardData?.getData('text') ?? '');
		memberInput = '';
	}

	const createMut = createMutation({
		mutationFn: () =>
			createEvent({
				body: {
					name,
					description: description || undefined,
					visibility,
					cycles: cycles.filter((c) => c.name.trim()).map((c) => ({ name: c.name })),
					forms: forms.filter((f) => f.label.trim()).map((f) => ({ type: f.type, label: f.label })),
					members: visibility === 'private' ? memberChips : null
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

				<!-- Tag input -->
				<div
					role="presentation"
					onclick={() => memberTagEl?.focus()}
					style="display:flex;flex-wrap:wrap;align-content:flex-start;gap:6px;min-height:72px;border-radius:9px;border:1px solid var(--border-h2);background:var(--s2);padding:10px;cursor:text;transition:box-shadow .15s;box-shadow:{memberTagFocused ? '0 0 0 3px var(--brand-soft)' : 'none'}"
				>
					{#each memberChips as email}
						<span style="display:inline-flex;align-items:center;gap:4px;background:var(--s3);border:1px solid var(--border-h2);border-radius:999px;padding:2px 8px 2px 10px;font-size:12px;color:var(--t1)">
							<span style="font-family:var(--font-mono,monospace)">{email}</span>
							<button
								type="button"
								onclick={(e) => { e.stopPropagation(); removeMemberChip(email); }}
								aria-label="Remove {email}"
								style="display:grid;place-items:center;width:16px;height:16px;border-radius:50%;border:none;background:none;cursor:pointer;font-size:13px;color:var(--t3);padding:0;line-height:1"
								onmouseenter={(e) => { e.currentTarget.style.background='var(--danger-soft)'; e.currentTarget.style.color='var(--danger)'; }}
								onmouseleave={(e) => { e.currentTarget.style.background='none'; e.currentTarget.style.color='var(--t3)'; }}
							>×</button>
						</span>
					{/each}
					<input
						bind:this={memberTagEl}
						bind:value={memberInput}
						onkeydown={handleMemberKeydown}
						onpaste={handleMemberPaste}
						onfocusin={() => (memberTagFocused = true)}
						onfocusout={() => (memberTagFocused = false)}
						placeholder={memberChips.length === 0 ? 'Paste or type emails…' : ''}
						style="flex:1;min-width:160px;border:none;background:transparent;outline:none;font-size:13px;color:var(--t1);font-family:inherit"
					/>
				</div>
				{#if memberChips.length > 0}
					<p style="font-size:12px;color:var(--t3);margin-top:8px">{memberChips.length} {memberChips.length === 1 ? 'person' : 'people'} added</p>
				{:else}
					<p style="font-size:12px;color:var(--t3);margin-top:8px">Paste a list or type and press Enter / comma. Supports <span style="font-family:monospace">Name &lt;email&gt;</span> format.</p>
				{/if}
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
