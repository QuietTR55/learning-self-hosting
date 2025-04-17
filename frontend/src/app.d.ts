// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	// Svelte 5 runes
	// These declarations allow TypeScript to recognize the $state, $derived, etc. syntax
	function $state<T>(initial: T): T;
	function $state<T>(): T | undefined;
	function $derived<T>(expression: T): T;
	const svelteHTML: {
		[key: string]: any;
	};
}

export {};
