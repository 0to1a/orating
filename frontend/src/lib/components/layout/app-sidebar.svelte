<script lang="ts">
	import { page } from '$app/stores';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import NavUser from './nav-user.svelte';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import { createQuery } from '@tanstack/svelte-query';
	import { listCompanies } from '$lib/api-client.js';
	import { getProfile } from '$lib/auth.js';
	import type { ComponentProps } from 'svelte';

	let { ...restProps }: ComponentProps<typeof Sidebar.Root> = $props();

	const profile = $derived(getProfile());

	const companiesQuery = createQuery({
		queryKey: ['companies'],
		queryFn: () => listCompanies().then((r) => r.data?.companies ?? [])
	});

	const activeCompany = $derived(
		($companiesQuery.data ?? []).find((c) => c.id === profile?.selectedCompanyId) ?? null
	);

	const navItems = [{ href: '/app/dashboard', label: 'Dashboard', icon: LayoutDashboard }];
</script>

<Sidebar.Root collapsible="offcanvas" {...restProps}>
	<Sidebar.Header>
		<div style="display:flex;align-items:center;gap:11px;padding:8px 8px 16px">
			<div style="
				width:38px;height:38px;border-radius:10px;flex-shrink:0;
				background:linear-gradient(135deg,var(--brand),var(--brand-2));
				display:grid;place-items:center;
				font-family:var(--display);font-weight:700;font-size:18px;color:#fff;
				box-shadow:0 6px 18px rgba(109,94,246,.35);
			">O</div>
			<div>
				<div style="font-weight:600;font-size:14px;letter-spacing:-.01em;color:var(--t1)">
					{activeCompany?.name ?? 'Orating'}
				</div>
				<div style="font-size:11px;color:var(--t3);text-transform:uppercase;letter-spacing:.08em;margin-top:2px">
					Workspace
				</div>
			</div>
		</div>
	</Sidebar.Header>

	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>Workspace</Sidebar.GroupLabel>
			<Sidebar.Menu>
				{#each navItems as item (item.href)}
					{@const isActive = $page.url.pathname.startsWith('/app/dashboard')}
					<Sidebar.MenuItem>
						<Sidebar.MenuButton tooltipContent={item.label} isActive={isActive}>
							{#snippet child({ props })}
								<a href={item.href} {...props}>
									<item.icon />
									<span>{item.label}</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				{/each}
			</Sidebar.Menu>
		</Sidebar.Group>
	</Sidebar.Content>

	<Sidebar.Footer>
		<NavUser />
	</Sidebar.Footer>
</Sidebar.Root>
