// Pure theme-resolution helpers (no runes) so they're trivially unit-testable
// and reusable by the anti-flash inline script's logic.

export type ThemePreference = 'light' | 'dark' | 'system';
export type EffectiveTheme = 'light' | 'dark';

// STORAGE_KEY is the localStorage key the preference is persisted under. The
// inline anti-flash script in app.html reads the same key before first paint.
export const STORAGE_KEY = 'theme';

// normalizePreference coerces an arbitrary stored value into a valid preference,
// defaulting to 'system'.
export function normalizePreference(raw: string | null | undefined): ThemePreference {
	return raw === 'light' || raw === 'dark' || raw === 'system' ? raw : 'system';
}

// resolveEffective turns a preference + the current OS dark-mode signal into the
// concrete theme to apply. 'system' follows the OS; explicit choices ignore it.
export function resolveEffective(pref: ThemePreference, systemDark: boolean): EffectiveTheme {
	if (pref === 'system') return systemDark ? 'dark' : 'light';
	return pref;
}
