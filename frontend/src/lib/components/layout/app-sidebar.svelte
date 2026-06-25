<script lang="ts">
	import { page } from '$app/stores';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import CompanySwitcher from './company-switcher.svelte';
	import NavUser from './nav-user.svelte';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import Star from '@lucide/svelte/icons/star';
	import type { ComponentProps } from 'svelte';

	let { ...restProps }: ComponentProps<typeof Sidebar.Root> = $props();

	const navItems = [
		{ href: '/app/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/app/events', label: 'Events', icon: Star }
	];
</script>

<Sidebar.Root collapsible="offcanvas" {...restProps}>
	<Sidebar.Header>
		<CompanySwitcher />
	</Sidebar.Header>

	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>Workspace</Sidebar.GroupLabel>
			<Sidebar.Menu>
				{#each navItems as item (item.href)}
					{@const isActive = $page.url.pathname === item.href}
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
