<script lang="ts">
	import Star from '@lucide/svelte/icons/star';

	interface Props {
		totalStars?: number;
		value?: number;
		onRate?: (rating: number) => void;
		size?: 'sm' | 'md' | 'lg';
		class?: string;
		disabled?: boolean;
	}

	let {
		totalStars = 5,
		value = $bindable(0),
		onRate,
		size = 'md',
		class: className = '',
		disabled = false
	}: Props = $props();

	let hover = $state(0);

	const sizePx = { sm: 20, md: 28, lg: 36 };

	function handleRate(star: number) {
		if (disabled) return;
		value = star;
		onRate?.(star);
	}
</script>

<div class="star-root {className}" class:disabled>
	{#each Array.from({ length: totalStars }, (_, i) => i + 1) as star}
		<button
			type="button"
			class="star-btn"
			class:active={(hover || value) >= star}
			class:hovered={hover >= star}
			{disabled}
			onclick={() => handleRate(star)}
			onmouseenter={() => !disabled && (hover = star)}
			onmouseleave={() => !disabled && (hover = 0)}
			aria-label="Rate {star} out of {totalStars}"
			aria-pressed={(value) >= star}
		>
			<Star size={sizePx[size]} class="star-icon" />
		</button>
	{/each}
</div>

<style>
	.star-root {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.star-root.disabled {
		opacity: 0.5;
	}

	.star-btn {
		border: none;
		background: none;
		padding: 2px;
		cursor: pointer;
		color: var(--s4);
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 6px;
		transition: color 0.2s ease, transform 0.15s ease;
		transform-origin: center;
	}
	.star-btn:disabled {
		cursor: not-allowed;
	}
	.star-btn:hover:not(:disabled) {
		transform: scale(1.3) rotate(-10deg);
	}
	.star-btn:active:not(:disabled) {
		transform: scale(0.9) rotate(15deg);
	}
	.star-btn.active {
		color: #FBBF24;
	}
	.star-btn:focus-visible {
		outline: none;
		box-shadow: 0 0 0 3px var(--brand-soft);
	}

	/* Star icon fill */
	.star-btn :global(.star-icon) {
		fill: transparent;
		stroke: currentColor;
		stroke-width: 1.5px;
		transition: fill 0.2s ease, transform 0.2s ease;
	}
	.star-btn.active :global(.star-icon) {
		fill: currentColor;
	}
</style>
