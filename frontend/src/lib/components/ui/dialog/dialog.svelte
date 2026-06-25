<script lang="ts">
	let { open = $bindable(false), children }: { open?: boolean; children?: import('svelte').Snippet } = $props();
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<div class="bg-background/80 fixed inset-0 backdrop-blur-sm" onclick={() => (open = false)} role="button" tabindex="-1" onkeydown={() => {}}></div>
		<div class="relative z-50">
			{#if children}{@render children()}{/if}
		</div>
	</div>
{/if}
