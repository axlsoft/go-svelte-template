import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// cn merges Tailwind class lists, resolving conflicts (last wins). The shadcn
// convention used by every component in this directory.
export function cn(...inputs: ClassValue[]): string {
	return twMerge(clsx(inputs));
}
