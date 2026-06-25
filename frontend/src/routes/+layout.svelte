<script lang="ts">
	import '../app.css';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from 'svelte-sonner';
	import { QueryClientProvider } from '@tanstack/svelte-query';
	import { createQueryClient } from '$lib/query-client';

	const queryClient = createQueryClient();

	let { children } = $props();
</script>

<svelte:head>
	<title>App</title>
</svelte:head>

<ModeWatcher defaultMode="dark" />
<QueryClientProvider client={queryClient}>
	{@render children()}
</QueryClientProvider>
<Toaster richColors />
