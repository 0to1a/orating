<script lang="ts">
	import { cn } from '$lib/utils.js';
	let { class: className = '', children }: { class?: string; children?: import('svelte').Snippet } = $props();
</script>
<p class={cn('text-muted-foreground text-sm', className)}>{#if children}{@render children()}{/if}</p>
