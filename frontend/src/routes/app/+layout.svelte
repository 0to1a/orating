<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { isAuthenticated, getProfile } from '$lib/auth.js';
	import { createQuery } from '@tanstack/svelte-query';
	import { listCompanies } from '$lib/api-client.js';
	import AppSidebar from '$lib/components/layout/app-sidebar.svelte';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';

	let { children } = $props();

	onMount(() => {
		if (!isAuthenticated()) goto('/login');
	});

	const profile = $derived(getProfile());

	const companiesQuery = createQuery({
		queryKey: ['companies'],
		queryFn: () => listCompanies().then((r) => r.data?.companies ?? [])
	});

	const activeCompany = $derived(
		$companiesQuery.data?.find((c) => c.id === profile?.selectedCompanyId) ?? null
	);

	const pageTitle = $derived.by(() => {
		const seg = $page.url.pathname.split('/').filter(Boolean).at(-1) ?? '';
		return seg.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
	});
</script>

<svelte:head>
	<title>App | {pageTitle}</title>
</svelte:head>

<Sidebar.Provider
	style="--sidebar-width: calc(var(--spacing) * 72); --header-height: calc(var(--spacing) * 12);"
>
	<AppSidebar variant="inset" />
	<Sidebar.Inset>
		<header
			class="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear"
		>
			<div class="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
				<Sidebar.Trigger class="-ms-1" />
				<Separator orientation="vertical" class="mx-2 data-[orientation=vertical]:h-4" />
				<nav class="flex items-center gap-1 text-sm" aria-label="Breadcrumb">
					{#if activeCompany}
						<span class="text-muted-foreground">{activeCompany.name}</span>
						<span class="text-muted-foreground mx-1">/</span>
					{/if}
					<span class="font-semibold">{pageTitle}</span>
				</nav>
			</div>
		</header>
		<div class="flex flex-1 flex-col">
			<div class="@container/main flex flex-1 flex-col gap-2">
				{@render children()}
			</div>
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
