<script lang="ts">
	import { Menu } from 'lucide-svelte';

	let { ontoggleSidebar }: { ontoggleSidebar?: () => void } = $props();
</script>

<header class="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-40 w-full border-b backdrop-blur">
	<div class="flex h-14 items-center px-4">
		<button
			onclick={ontoggleSidebar}
			class="hover:bg-accent hover:text-accent-foreground -ml-2 flex h-9 w-9 items-center justify-center rounded-md"
			aria-label="Toggle sidebar"
		>
			<Menu class="h-5 w-5" />
		</button>
	</div>
</header>
