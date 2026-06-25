<script lang="ts">
	import { goto } from '$app/navigation';
	import { createQuery } from '@tanstack/svelte-query';
	import { clearAuth, getProfile } from '$lib/auth.js';
	import { listCompanies } from '$lib/api-client.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import ChevronsUpDown from '@lucide/svelte/icons/chevrons-up-down';
	import LogOut from '@lucide/svelte/icons/log-out';
	import UserPen from '@lucide/svelte/icons/user-pen';
	import Users from '@lucide/svelte/icons/users';
	import Key from '@lucide/svelte/icons/key';
	import Flag from '@lucide/svelte/icons/flag';

	const sidebar = Sidebar.useSidebar();
	const profile = $derived(getProfile());

	const companiesQuery = createQuery({
		queryKey: ['companies'],
		queryFn: () => listCompanies().then((r) => r.data?.companies ?? [])
	});

	const activeCompany = $derived(
		$companiesQuery.data?.find((c) => c.id === profile?.selectedCompanyId) ?? null
	);

	const isAdmin = $derived(
		activeCompany?.role === 'admin' || activeCompany?.isOwner === true
	);
	const isOwner = $derived(activeCompany?.isOwner === true);

	const initials = $derived(
		profile?.name
			?.split(' ')
			.map((w) => w[0])
			.slice(0, 2)
			.join('')
			.toUpperCase() ?? '?'
	);

	const AVATAR_COLORS = ['#7c5cff', '#0ea5a4', '#d8754a', '#5e7cf5', '#ec4899', '#10b981'];
	const avatarColor = $derived(
		AVATAR_COLORS[
			(profile?.name ?? '').split('').reduce((a, c) => a + c.charCodeAt(0), 0) % AVATAR_COLORS.length
		]
	);

	async function logout() {
		clearAuth();
		await goto('/login');
	}
</script>

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						size="lg"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						{...props}
					>
						<span
							class="avatar-initials size-8"
							style="background: {avatarColor}"
						>{initials}</span>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-medium">{profile?.name ?? ''}</span>
							<span class="mono truncate text-xs opacity-70">{profile?.email ?? ''}</span>
						</div>
						<ChevronsUpDown class="ms-auto size-4" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
				side={sidebar.isMobile ? 'bottom' : 'top'}
				align="end"
				sideOffset={8}
			>
				<DropdownMenu.Label class="p-0 font-normal">
					<div class="flex items-center gap-2 px-1 py-1.5 text-start text-sm">
						<span
							class="avatar-initials size-8"
							style="background: {avatarColor}"
						>{initials}</span>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-medium">{profile?.name ?? ''}</span>
							<span class="mono truncate text-xs opacity-70">{profile?.email ?? ''}</span>
						</div>
					</div>
				</DropdownMenu.Label>
				<DropdownMenu.Separator />
				<DropdownMenu.Group>
					<DropdownMenu.Item onSelect={() => goto('/app/profile')}>
						<UserPen class="size-4" />
						Edit profile
					</DropdownMenu.Item>
					{#if isAdmin}
						<DropdownMenu.Item onSelect={() => goto('/app/team')}>
							<Users class="size-4" />
							Team
						</DropdownMenu.Item>
						<DropdownMenu.Item onSelect={() => goto('/app/api-keys')}>
							<Key class="size-4" />
							API Keys
						</DropdownMenu.Item>
					{/if}
					{#if isOwner}
						<DropdownMenu.Item onSelect={() => goto('/app/feature-flags')}>
							<Flag class="size-4" />
							Feature Flags
						</DropdownMenu.Item>
					{/if}
				</DropdownMenu.Group>
				<DropdownMenu.Separator />
				<DropdownMenu.Item onSelect={logout}>
					<LogOut class="size-4" />
					Log out
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
