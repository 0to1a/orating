<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { getProfile } from '$lib/auth.js';
	import { listCompanies, listCompanyMembers } from '$lib/api-client.js';
	import Building2 from '@lucide/svelte/icons/building-2';
	import Users from '@lucide/svelte/icons/users';
	import KeyRound from '@lucide/svelte/icons/key-round';
	import Flag from '@lucide/svelte/icons/flag';
	import UserPlus from '@lucide/svelte/icons/user-plus';
	import Plus from '@lucide/svelte/icons/plus';

	const profile = $derived(getProfile());

	const companiesQuery = createQuery({
		queryKey: ['companies'],
		queryFn: () => listCompanies().then((r) => r.data?.companies ?? [])
	});
	const membersQuery = createQuery({
		queryKey: ['members'],
		queryFn: () => listCompanyMembers().then((r) => r.data?.members ?? [])
	});

	const activeCompany = $derived(
		($companiesQuery.data ?? []).find((c) => c.id === profile?.selectedCompanyId) ?? null
	);

	const metrics = $derived([
		{
			label: 'Companies',
			value: $companiesQuery.isLoading ? '…' : ($companiesQuery.data?.length ?? '—'),
			icon: Building2,
			href: '/app/companies'
		},
		{
			label: 'Members',
			value: $membersQuery.isLoading ? '…' : ($membersQuery.data?.length ?? '—'),
			icon: Users,
			href: '/app/team'
		},
		{
			label: 'API Keys',
			value: '—',
			icon: KeyRound,
			href: '/app/api-keys'
		},
		{
			label: 'Feature Flags',
			value: '—',
			icon: Flag,
			href: '/app/feature-flags'
		}
	]);

	const recentMembers = $derived(($membersQuery.data ?? []).slice(0, 4));

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
</script>

<div class="flex flex-col gap-6 py-6">
	<!-- Header -->
	<div class="flex items-end justify-between px-4 lg:px-6">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">
				{profile?.name ? `Welcome back, ${profile.name.split(' ')[0]}` : 'Welcome back'}
			</h1>
			<p class="text-muted-foreground mt-1 text-sm">
				Here's what's happening across {activeCompany?.name ?? 'your workspace'}.
			</p>
		</div>
		<a
			href="/app/team"
			class="inline-flex items-center gap-2 rounded-[9px] bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
		>
			<UserPlus class="size-4" />
			Invite member
		</a>
	</div>

	<!-- Metric cards -->
	<div class="grid grid-cols-2 gap-3.5 px-4 lg:grid-cols-4 lg:px-6">
		{#each metrics as m}
			<a
				href={m.href}
				class="block rounded-[var(--radius-card)] border p-[18px] transition-colors hover:bg-[var(--s3)]"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				<div class="flex items-center justify-between" style="color: var(--t2)">
					<span class="text-[13px]">{m.label}</span>
					<m.icon class="size-[17px]" style="color: var(--t3)" />
				</div>
				<div class="mono my-3 text-3xl font-medium tracking-tight">{m.value}</div>
			</a>
		{/each}
	</div>

	<!-- Two-column layout -->
	<div class="grid grid-cols-1 gap-[18px] px-4 lg:grid-cols-[1.6fr_1fr] lg:px-6">
		<!-- Recent members -->
		<div
			class="overflow-hidden rounded-[var(--radius-card)] border"
			style="background: var(--s2); border-color: var(--border-h)"
		>
			<div
				class="flex items-center justify-between border-b px-[18px] py-4"
				style="border-color: var(--border-h)"
			>
				<h3 class="text-[14px] font-semibold">Recent members</h3>
				<a href="/app/team" class="text-[13px] transition-opacity hover:opacity-80" style="color: var(--brand-text)">
					View all
				</a>
			</div>
			<div>
				{#if $membersQuery.isLoading}
					<div class="px-[18px] py-6 text-sm" style="color: var(--t3)">Loading…</div>
				{:else if recentMembers.length === 0}
					<div class="px-[18px] py-6 text-sm" style="color: var(--t3)">
						No members yet — invite someone to get started.
					</div>
				{:else}
					{#each recentMembers as member}
						<div class="flex items-center gap-3 px-[18px] py-[11px] transition-colors hover:bg-[var(--s3)]">
							<span
								class="avatar-initials mt-0.5 size-[30px] rounded-lg text-[11px]"
								style="background: {avatarColor(member.name)}"
							>{initials(member.name)}</span>
							<div class="min-w-0 flex-1">
								<div class="text-[13.5px] font-medium">{member.name}</div>
								<div class="mono text-xs" style="color: var(--t3)">{member.email}</div>
							</div>
							<span class="text-xs capitalize" style="color: var(--t3)">{member.role}</span>
						</div>
					{/each}
				{/if}
			</div>
		</div>

		<!-- Quick actions -->
		<div class="flex flex-col gap-[18px]">
			<div
				class="overflow-hidden rounded-[var(--radius-card)] border"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				<div
					class="flex items-center border-b px-[18px] py-4"
					style="border-color: var(--border-h)"
				>
					<h3 class="text-[14px] font-semibold">Quick actions</h3>
				</div>
				<div class="flex flex-col gap-2 p-3">
					<button
						onclick={() => goto('/app/team')}
						class="flex w-full cursor-pointer items-center gap-2 rounded-[9px] border px-3 py-2.5 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
						style="background: none; border-color: var(--border-h); color: var(--t1)"
					>
						<UserPlus class="size-4" style="color: var(--t3)" />
						Invite a member
					</button>
					<button
						onclick={() => goto('/app/api-keys')}
						class="flex w-full cursor-pointer items-center gap-2 rounded-[9px] border px-3 py-2.5 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
						style="background: none; border-color: var(--border-h); color: var(--t1)"
					>
						<KeyRound class="size-4" style="color: var(--t3)" />
						Create API key
					</button>
					<button
						onclick={() => goto('/app/companies')}
						class="flex w-full cursor-pointer items-center gap-2 rounded-[9px] border px-3 py-2.5 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
						style="background: none; border-color: var(--border-h); color: var(--t1)"
					>
						<Plus class="size-4" style="color: var(--t3)" />
						Add a company
					</button>
				</div>
			</div>
		</div>
	</div>
</div>
