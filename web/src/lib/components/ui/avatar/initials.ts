// initialsFrom derives up to two uppercase initials from a display name, falling
// back to the email's first character, then '?'. Deterministic — the same inputs
// always yield the same initials.
export function initialsFrom(name?: string | null, email?: string | null): string {
	const n = (name ?? '').trim();
	if (n) {
		const parts = n.split(/\s+/).filter(Boolean);
		if (parts.length === 1) {
			return parts[0].slice(0, 2).toUpperCase();
		}
		const first = parts[0][0];
		const last = parts[parts.length - 1][0];
		return (first + last).toUpperCase();
	}
	const e = (email ?? '').trim();
	if (e) return e[0].toUpperCase();
	return '?';
}
