<script lang="ts">
	import { goto } from '$app/navigation';
	import { isAuthenticated } from '$lib/auth';
	import { onMount } from 'svelte';

	onMount(() => {
		if (isAuthenticated()) {
			goto('/app/dashboard');
		} else {
			goto('/login');
		}
	});
</script>
