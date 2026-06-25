<script lang="ts">
	import { onMount } from 'svelte';

	class Particle {
		x: number;
		y: number;
		dx: number;
		dy: number;
		size: number;

		constructor(x: number, y: number, dx: number, dy: number, size: number) {
			this.x = x;
			this.y = y;
			this.dx = dx;
			this.dy = dy;
			this.size = size;
		}

		draw(ctx: CanvasRenderingContext2D) {
			ctx.beginPath();
			ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2, false);
			ctx.fillStyle = 'rgba(109, 94, 246, 0.75)';
			ctx.fill();
		}

		update(
			ctx: CanvasRenderingContext2D,
			w: number,
			h: number,
			mouse: { x: number | null; y: number | null; radius: number }
		) {
			if (this.x > w || this.x < 0) this.dx = -this.dx;
			if (this.y > h || this.y < 0) this.dy = -this.dy;

			if (mouse.x !== null && mouse.y !== null) {
				const ex = mouse.x - this.x;
				const ey = mouse.y - this.y;
				const dist = Math.sqrt(ex * ex + ey * ey);
				if (dist < mouse.radius + this.size) {
					const force = (mouse.radius - dist) / mouse.radius;
					this.x -= (ex / dist) * force * 5;
					this.y -= (ey / dist) * force * 5;
				}
			}

			this.x += this.dx;
			this.y += this.dy;
			this.draw(ctx);
		}
	}

	let canvas: HTMLCanvasElement;

	onMount(() => {
		const ctx = canvas.getContext('2d')!;
		let rafId: number;
		let particles: Particle[] = [];
		const mouse = { x: null as number | null, y: null as number | null, radius: 200 };

		function init() {
			particles = [];
			const count = (canvas.height * canvas.width) / 9000;
			for (let i = 0; i < count; i++) {
				const size = Math.random() * 2 + 1;
				particles.push(
					new Particle(
						Math.random() * (canvas.width - size * 4) + size * 2,
						Math.random() * (canvas.height - size * 4) + size * 2,
						Math.random() * 0.4 - 0.2,
						Math.random() * 0.4 - 0.2,
						size
					)
				);
			}
		}

		function connect() {
			for (let a = 0; a < particles.length; a++) {
				for (let b = a + 1; b < particles.length; b++) {
					const ex = particles[a].x - particles[b].x;
					const ey = particles[a].y - particles[b].y;
					const dist = ex * ex + ey * ey;
					const threshold = (canvas.width / 7) * (canvas.height / 7);
					if (dist < threshold) {
						const opacity = (1 - dist / 20000) * 0.6;
						const nearMouse =
							mouse.x !== null &&
							Math.hypot(particles[a].x - mouse.x, particles[a].y - (mouse.y ?? 0)) <
								mouse.radius;
						ctx.strokeStyle = nearMouse
							? `rgba(139, 124, 255, ${opacity + 0.3})`
							: `rgba(109, 94, 246, ${opacity})`;
						ctx.lineWidth = 0.8;
						ctx.beginPath();
						ctx.moveTo(particles[a].x, particles[a].y);
						ctx.lineTo(particles[b].x, particles[b].y);
						ctx.stroke();
					}
				}
			}
		}

		function animate() {
			rafId = requestAnimationFrame(animate);
			ctx.fillStyle = '#0B0D12';
			ctx.fillRect(0, 0, canvas.width, canvas.height);
			for (const p of particles) p.update(ctx, canvas.width, canvas.height, mouse);
			connect();
		}

		function resize() {
			canvas.width = window.innerWidth;
			canvas.height = window.innerHeight;
			init();
		}

		const onMouseMove = (e: MouseEvent) => {
			mouse.x = e.clientX;
			mouse.y = e.clientY;
		};
		const onMouseOut = () => {
			mouse.x = null;
			mouse.y = null;
		};

		window.addEventListener('resize', resize);
		window.addEventListener('mousemove', onMouseMove);
		window.addEventListener('mouseout', onMouseOut);

		resize();
		animate();

		return () => {
			window.removeEventListener('resize', resize);
			window.removeEventListener('mousemove', onMouseMove);
			window.removeEventListener('mouseout', onMouseOut);
			cancelAnimationFrame(rafId);
		};
	});
</script>

<canvas bind:this={canvas} class="fixed inset-0 -z-10 h-full w-full"></canvas>
