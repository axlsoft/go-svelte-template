import {
	STORAGE_KEY,
	normalizePreference,
	resolveEffective,
	type EffectiveTheme,
	type ThemePreference
} from './resolve';

// ThemeState owns the theme preference (light | dark | system, default system),
// persists it to localStorage, and keeps the `.dark` class on <html> in sync.
// When the preference is 'system' it live-updates on OS theme changes via a
// matchMedia listener. A tiny inline script in app.html applies the same logic
// before first paint to avoid a flash-of-wrong-theme; this store takes over once
// the app mounts.
class ThemeState {
	preference = $state<ThemePreference>('system');
	systemDark = $state(false);

	#mql: MediaQueryList | null = null;

	// effective is the concrete theme currently applied.
	get effective(): EffectiveTheme {
		return resolveEffective(this.preference, this.systemDark);
	}

	// init reads the stored preference and subscribes to OS theme changes. Call
	// once from the root layout's onMount. Safe to call in non-browser contexts.
	init(): void {
		if (typeof window === 'undefined') return;
		this.preference = normalizePreference(localStorage.getItem(STORAGE_KEY));
		this.#mql = window.matchMedia('(prefers-color-scheme: dark)');
		this.systemDark = this.#mql.matches;
		this.#mql.addEventListener('change', this.#onSystemChange);
		this.#apply();
	}

	// set persists a new preference and applies it immediately.
	set(pref: ThemePreference): void {
		this.preference = pref;
		if (typeof window !== 'undefined') localStorage.setItem(STORAGE_KEY, pref);
		this.#apply();
	}

	#onSystemChange = (e: MediaQueryListEvent): void => {
		this.systemDark = e.matches;
		this.#apply();
	};

	#apply(): void {
		if (typeof document === 'undefined') return;
		document.documentElement.classList.toggle('dark', this.effective === 'dark');
	}
}

export const theme = new ThemeState();
