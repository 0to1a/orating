<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import Search from '@lucide/svelte/icons/search';
	import Trash2 from '@lucide/svelte/icons/trash-2';
	import X from '@lucide/svelte/icons/x';
	import UserPlus from '@lucide/svelte/icons/user-plus';
	import type { Member } from '$lib/api/types.gen.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogDescription,
		DialogFooter
	} from '$lib/components/ui/dialog/index.js';
	import { getProfile } from '$lib/auth.js';
	import {
		listCompanyMembers,
		inviteCompanyMember,
		removeCompanyMember,
		updateCompanyMemberRole,
		listCompanyRoles
	} from '$lib/api-client.js';

	const qc = useQueryClient();

	const selfId = $derived(getProfile()?.id ?? -1);

	const membersQuery = createQuery({
		queryKey: ['members'],
		queryFn: () => listCompanyMembers().then((r) => r.data?.members ?? [])
	});

	const rolesQuery = createQuery({
		queryKey: ['roles'],
		queryFn: () => listCompanyRoles().then((r) => r.data?.roles ?? [])
	});

	let inviteOpen = $state(false);
	let inviteChips = $state<Array<{ name: string; email: string }>>([]);
	let inviteInput = $state('');
	let inviteRole = $state('member');
	let inviteTagEl = $state<HTMLInputElement | null>(null);
	let inviteTagFocused = $state(false);

	$effect(() => {
		if (!inviteOpen) {
			inviteChips = [];
			inviteInput = '';
			inviteRole = 'member';
		}
	});
	let removeTarget = $state<Member | null>(null);
	let removeOpen = $state(false);
	let searchQuery = $state('');
	let roleFilter = $state('');

	const AVATAR_COLORS = ['#7c5cff', '#0ea5a4', '#d8754a', '#5e7cf5', '#ec4899', '#10b981'];
	function avatarColor(name: string) {
		return AVATAR_COLORS[
			name.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % AVATAR_COLORS.length
		];
	}
	function initials(name: string) {
		return name
			.split(' ')
			.map((w) => w[0])
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	const filteredMembers = $derived(
		($membersQuery.data ?? []).filter((m) => {
			const q = searchQuery.toLowerCase();
			const matchesSearch =
				!q || m.name.toLowerCase().includes(q) || m.email.toLowerCase().includes(q);
			const matchesRole = !roleFilter || m.role === roleFilter;
			return matchesSearch && matchesRole;
		})
	);

	function parseEntries(raw: string): Array<{ name: string; email: string }> {
		return raw
			.split(/[,\n]+/)
			.map((s) => s.trim())
			.filter(Boolean)
			.flatMap((entry) => {
				const match = entry.match(/^(.+?)\s*<([^>]+)>\s*$/);
				if (match) {
					const name = match[1].trim();
					const email = match[2].trim().toLowerCase();
					if (!email.includes('@')) return [];
					return [{ name, email }];
				}
				const email = entry.trim().toLowerCase();
				if (!email.includes('@')) return [];
				const local = email.split('@')[0];
				const name = local
					.split(/[._-]/)
					.map((w) => w.charAt(0).toUpperCase() + w.slice(1))
					.join(' ');
				return [{ name, email }];
			});
	}

	function addEntries(raw: string) {
		const existing = new Set(inviteChips.map((c) => c.email));
		const fresh = parseEntries(raw).filter((p) => !existing.has(p.email));
		inviteChips = [...inviteChips, ...fresh];
	}

	function removeChip(email: string) {
		inviteChips = inviteChips.filter((c) => c.email !== email);
	}

	function handleTagKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' || e.key === ',' || e.key === 'Tab') {
			if (inviteInput.trim()) {
				e.preventDefault();
				addEntries(inviteInput);
				inviteInput = '';
			}
		} else if (e.key === 'Backspace' && !inviteInput && inviteChips.length > 0) {
			inviteChips = inviteChips.slice(0, -1);
		}
	}

	function handleTagPaste(e: ClipboardEvent) {
		e.preventDefault();
		const text = e.clipboardData?.getData('text') ?? '';
		addEntries(text);
		inviteInput = '';
	}

	const inviteMut = createMutation({
		mutationFn: async () => {
			const results = await Promise.allSettled(
				inviteChips.map(({ name, email }) =>
					inviteCompanyMember({ body: { email, name, role: inviteRole } })
				)
			);
			const succeeded = results.filter((r) => r.status === 'fulfilled').length;
			const failed = results.filter((r) => r.status === 'rejected').length;
			return { succeeded, failed };
		},
		onSuccess: ({ succeeded, failed }) => {
			qc.invalidateQueries({ queryKey: ['members'] });
			inviteOpen = false;
			inviteChips = [];
			inviteInput = '';
			inviteRole = 'member';
			if (failed > 0 && succeeded > 0) {
				toast.warning(`${succeeded} added, ${failed} failed`);
			} else if (failed > 0) {
				toast.error(`Failed to add ${failed} member${failed !== 1 ? 's' : ''}`);
			} else {
				toast.success(`${succeeded} member${succeeded !== 1 ? 's' : ''} added`);
			}
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to send invitation')
	});

	const removeMut = createMutation({
		mutationFn: (userId: number) => removeCompanyMember({ path: { userId } }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ['members'] });
			removeOpen = false;
			removeTarget = null;
			toast.success('Member removed');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to remove member')
	});

	const updateRoleMut = createMutation({
		mutationFn: ({ userId, role }: { userId: number; role: string }) =>
			updateCompanyMemberRole({ path: { userId }, body: { role } }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ['members'] });
			toast.success('Role updated');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to update role')
	});
</script>

<div class="flex flex-col gap-6 py-6">
	<!-- Page header -->
	<div class="flex items-end justify-between px-4 lg:px-6">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Team</h1>
			<p class="text-muted-foreground mt-1 text-sm">Manage members in your selected company.</p>
		</div>
		<button
			onclick={() => (inviteOpen = true)}
			class="inline-flex items-center gap-2 rounded-[9px] bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
		>
			<UserPlus class="size-4" />
			Invite
		</button>
	</div>

	<div class="px-4 lg:px-6">
		<!-- Toolbar -->
		<div class="mb-4 flex items-center gap-2.5">
			<div class="relative max-w-xs flex-1">
				<Search
					class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2"
					style="color: var(--t3)"
				/>
				<input
					bind:value={searchQuery}
					placeholder="Search by name or email"
					class="w-full rounded-[9px] border py-2 pl-9 pr-3 text-[13.5px] outline-none transition-shadow"
					style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
					onfocus={(e) => (e.currentTarget.style.boxShadow = '0 0 0 3px var(--brand-soft)')}
					onblur={(e) => (e.currentTarget.style.boxShadow = '')}
				/>
			</div>
			<select
				bind:value={roleFilter}
				class="rounded-[9px] border px-3 py-2 text-[13px] outline-none transition-shadow"
				style="background: var(--s3); border-color: var(--border-h2); color: var(--t2); font-family: inherit"
				onfocus={(e) => (e.currentTarget.style.boxShadow = '0 0 0 3px var(--brand-soft)')}
				onblur={(e) => (e.currentTarget.style.boxShadow = '')}
			>
				<option value="">All roles</option>
				{#each $rolesQuery.data ?? [] as r}
					<option value={r.value}>{r.label}</option>
				{/each}
			</select>
		</div>

		{#if $membersQuery.isLoading}
			<Skeleton class="h-32 w-full" />
		{:else if filteredMembers.length === 0}
			<div
				class="rounded-[var(--radius-card)] border py-10 text-center text-sm"
				style="background: var(--s2); border-color: var(--border-h); color: var(--t3)"
			>
				{searchQuery || roleFilter ? 'No members match your filter.' : 'No members yet — invite someone to get started.'}
			</div>
		{:else}
			<div
				class="overflow-hidden rounded-[var(--radius-card)] border"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				<table class="w-full border-collapse">
					<thead>
						<tr>
							{#each ['Member', 'Role', 'Status', 'Joined', ''] as col}
								<th
									class="px-[18px] py-[13px] text-left text-[11px] font-semibold uppercase tracking-[0.05em]"
									style="background: var(--s1); color: var(--t3); border-bottom: 1px solid var(--border-h)"
								>{col}</th>
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each filteredMembers as member}
							<tr class="group">
								<td
									class="px-[18px] py-[14px]"
									style="border-bottom: 1px solid var(--border-h)"
								>
									<div class="flex items-center gap-3">
										<span
											class="avatar-initials size-8"
											style="background: {avatarColor(member.name)}"
										>{initials(member.name)}</span>
										<div>
											<div class="text-[14px] font-semibold">{member.name}</div>
											<div class="mono text-[12.5px]" style="color: var(--t3)">{member.email}</div>
										</div>
									</div>
								</td>
								<td
									class="px-[18px] py-[14px]"
									style="border-bottom: 1px solid var(--border-h)"
								>
									{#if member.userId === selfId}
										<span class="text-[13px] capitalize" style="color: var(--t3)">{member.role}</span>
									{:else}
										<select
											value={member.role}
											onchange={(e) =>
												$updateRoleMut.mutate({
													userId: member.userId,
													role: (e.target as HTMLSelectElement).value
												})}
											class="rounded-lg border px-2.5 py-1.5 text-[13px] outline-none transition-shadow"
											style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
											onfocus={(e) => (e.currentTarget.style.boxShadow = '0 0 0 3px var(--brand-soft)')}
											onblur={(e) => (e.currentTarget.style.boxShadow = '')}
										>
											{#each $rolesQuery.data ?? [] as r}
												<option value={r.value}>{r.label}</option>
											{/each}
										</select>
									{/if}
								</td>
								<td
									class="px-[18px] py-[14px]"
									style="border-bottom: 1px solid var(--border-h)"
								>
									<span
										class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[12px] font-medium"
										style="background: var(--ok-soft); color: var(--ok)"
									>
										<span class="size-1.5 rounded-full bg-current"></span>
										Active
									</span>
								</td>
								<td
									class="mono px-[18px] py-[14px] text-[13px]"
									style="border-bottom: 1px solid var(--border-h); color: var(--t2)"
								>
									{new Date(member.joinedAt).toLocaleDateString()}
								</td>
								<td
									class="px-[18px] py-[14px]"
									style="border-bottom: 1px solid var(--border-h)"
								>
									{#if member.userId !== selfId}
										<button
											onclick={() => { removeTarget = member; removeOpen = true; }}
											aria-label="Remove member"
											class="grid size-8 place-items-center rounded-lg border-none bg-transparent opacity-0 transition-all group-hover:opacity-100 hover:bg-[var(--danger-soft)] hover:!text-[var(--danger)]"
											style="color: var(--t3)"
										>
											<Trash2 class="size-4" />
										</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

<!-- Invite modal -->
<Dialog bind:open={inviteOpen}>
	<DialogContent class="w-[500px] max-w-full p-6">
		<DialogHeader class="mb-5">
			<div class="flex items-start justify-between">
				<div>
					<DialogTitle class="text-[18px] font-semibold">Invite members</DialogTitle>
					<DialogDescription class="mt-1 text-[13px] text-muted-foreground">
						Members are added immediately — no email confirmation required.
					</DialogDescription>
				</div>
				<button
					onclick={() => (inviteOpen = false)}
					aria-label="Close"
					class="grid size-8 place-items-center rounded-lg border-none bg-transparent transition-colors hover:bg-[var(--s3)]"
					style="color: var(--t2)"
				>
					<X class="size-4" />
				</button>
			</div>
		</DialogHeader>

		<div class="space-y-4">
			<div>
				<label
					for="invite-tag-input"
					class="mb-1.5 block text-[12.5px] font-medium"
					style="color: var(--t2)"
				>Email addresses</label>
				<!-- Tag input container -->
				<div
					role="presentation"
					onclick={() => inviteTagEl?.focus()}
					class="flex min-h-[80px] flex-wrap content-start gap-1.5 rounded-[9px] border p-2.5 transition-shadow"
					style="background: var(--s3); border-color: var(--border-h2); cursor: text; box-shadow: {inviteTagFocused ? '0 0 0 3px var(--brand-soft)' : 'none'}"
				>
					{#each inviteChips as { name, email }}
						<span
							class="inline-flex max-w-full items-center gap-1 rounded-full border py-0.5 pl-2.5 pr-1 text-[12px]"
							style="background: var(--s1); border-color: var(--border-h2); color: var(--t1)"
						>
							<span class="truncate font-medium">{name}</span>
							<span class="mono truncate" style="color: var(--t3)">{email}</span>
							<button
								type="button"
								onclick={(e) => { e.stopPropagation(); removeChip(email); }}
								aria-label="Remove {name}"
								class="ml-0.5 grid size-4 flex-shrink-0 place-items-center rounded-full text-[11px] transition-colors hover:bg-[var(--danger-soft)] hover:text-[var(--danger)]"
								style="color: var(--t3)"
							>×</button>
						</span>
					{/each}
					<input
						id="invite-tag-input"
						bind:this={inviteTagEl}
						bind:value={inviteInput}
						onkeydown={handleTagKeydown}
						onpaste={handleTagPaste}
						onfocusin={() => (inviteTagFocused = true)}
						onfocusout={() => (inviteTagFocused = false)}
						placeholder={inviteChips.length === 0 ? 'Paste or type emails…' : ''}
						class="min-w-[160px] flex-1 bg-transparent text-[13px] outline-none"
						style="color: var(--t1); font-family: inherit"
					/>
				</div>
				<p class="mt-1.5 text-[12px]" style="color: var(--t3)">
					Paste a list or type and press <span class="mono">Enter</span> / <span class="mono">,</span> to add. Supports <span class="mono">Name &lt;email&gt;</span> format.
				</p>
			</div>

			<div>
				<label
					for="invite-role"
					class="mb-1.5 block text-[12.5px] font-medium"
					style="color: var(--t2)"
				>Role</label>
				<select
					id="invite-role"
					bind:value={inviteRole}
					class="w-full rounded-[9px] border px-3 py-2.5 text-[13.5px] outline-none transition-shadow"
					style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
					onfocus={(e) => (e.currentTarget.style.boxShadow = '0 0 0 3px var(--brand-soft)')}
					onblur={(e) => (e.currentTarget.style.boxShadow = '')}
				>
					{#each $rolesQuery.data ?? [] as r}
						<option value={r.value}>{r.label}</option>
					{/each}
				</select>
				<p class="mt-1.5 text-[12px]" style="color: var(--t3)">
					Admins can manage members, API keys, and billing.
				</p>
			</div>
		</div>

		<DialogFooter class="mt-5 gap-2.5 sm:justify-end">
			<button
				onclick={() => (inviteOpen = false)}
				class="rounded-[9px] border px-4 py-2 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
				style="background: none; border-color: var(--border-h2); color: var(--t1)"
			>Cancel</button>
			<button
				onclick={() => $inviteMut.mutate()}
				disabled={$inviteMut.isPending || inviteChips.length === 0}
				class="rounded-[9px] bg-primary px-4 py-2 text-[13px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
			>
				{#if $inviteMut.isPending}
					Adding…
				{:else if inviteChips.length > 0}
					Add {inviteChips.length} {inviteChips.length === 1 ? 'member' : 'members'}
				{:else}
					Add members
				{/if}
			</button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<!-- Confirm remove modal -->
<Dialog bind:open={removeOpen}>
	<DialogContent class="w-[400px] max-w-full p-6">
		<DialogHeader>
			<DialogTitle class="text-[18px] font-semibold">Remove member</DialogTitle>
			<DialogDescription class="mt-1 text-[13px] text-muted-foreground">
				Remove <strong class="text-foreground">{removeTarget?.name}</strong> from the company?
				They will lose access immediately.
			</DialogDescription>
		</DialogHeader>
		<DialogFooter class="mt-5 gap-2.5 sm:justify-end">
			<button
				onclick={() => (removeOpen = false)}
				class="rounded-[9px] border px-4 py-2 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
				style="background: none; border-color: var(--border-h2); color: var(--t1)"
			>Cancel</button>
			<button
				onclick={() => removeTarget && $removeMut.mutate(removeTarget.userId)}
				disabled={$removeMut.isPending}
				class="rounded-[9px] px-4 py-2 text-[13px] font-medium transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
				style="background: var(--danger-soft); color: var(--danger)"
			>
				{$removeMut.isPending ? 'Removing…' : 'Remove'}
			</button>
		</DialogFooter>
	</DialogContent>
</Dialog>
